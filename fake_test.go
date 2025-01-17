package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testlabtools/record/client"
)

type FakeServer struct {
	mux    *http.ServeMux
	server *httptest.Server

	env map[string]string

	runs     map[string]client.CIRunRequest
	files    [][]byte
	fileUrls []string
	status   map[int]client.RunFileUploadStatus
}

func (s *FakeServer) Close() {
	s.server.Close()
}

func newFakeServer(t *testing.T, l *slog.Logger, ci client.CIProviderName) *FakeServer {
	t.Helper()

	assert := assert.New(t)

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	env := map[string]string{
		"TESTLAB_KEY":   "some-key",
		"TESTLAB_HOST":  server.URL,
		"TESTLAB_GROUP": "e2e",
	}

	fs := &FakeServer{
		mux:    mux,
		server: server,

		env: env,

		runs:   make(map[string]client.CreateRunJSONRequestBody),
		status: make(map[int]client.RunFileUploadStatus),
	}

	switch ci {
	case client.Github:
		fs.useGitHub()
	default:
		panic(fmt.Sprintf("unknown CI provider name: %q", ci))
	}

	log := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			l.Info("got fake request", "method", r.Method, "url", r.URL)
			handler(w, r)
		}
	}

	secure := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Check X-API-Key is present.
			assert.NotEmpty(r.Header.Get(HeaderAPIKey), HeaderAPIKey+" is missing")

			w.Header().Set("Content-Type", "application/json")
			log(handler)(w, r)
		}
	}

	createRun := func(w http.ResponseWriter, r *http.Request) {
		var run client.CIRunRequest
		mustDecode(r.Body, &run)

		assert.NotEmpty(run.ActorName, "GITHUB_ACTOR")
		assert.NotEmpty(run.CiProviderName, "CI_PROVIDER_NAME")
		assert.NotEmpty(run.GitRef, "GITHUB_REF")
		assert.NotEmpty(run.GitRefName, "GITHUB_REF_NAME")
		assert.NotEmpty(run.GitRepo, "GITHUB_REPO")
		assert.NotEmpty(run.GitSha, "GITHUB_SHA")
		assert.NotEmpty(run.Group, "TESTLAB_GROUP")
		assert.NotEmpty(run.RunAttempt, "GITHUB_RUN_ATTEMPT")
		assert.NotEmpty(run.RunId, "GITHUB_RUN_ID")
		assert.NotEmpty(run.RunNumber, "GITHUB_RUN_NUMBER")

		// Determine if runId-group pair is the first created run.
		idx := fmt.Sprintf("%d-%s", run.RunId, run.Group)
		_, ok := fs.runs[idx]
		fs.runs[idx] = run
		created := !ok

		status := http.StatusOK
		if created {
			status = http.StatusCreated
		}

		w.WriteHeader(status)
		resp := client.CIRunResponse{
			Id: fmt.Sprint(len(fs.runs)),
		}
		json.NewEncoder(w).Encode(resp)
	}

	postFileUpload := func(w http.ResponseWriter, r *http.Request) {
		id := len(fs.fileUrls) + 1
		url := fmt.Sprintf("%s/s3/files/%d", server.URL, id)
		fs.fileUrls = append(fs.fileUrls, url)

		w.WriteHeader(http.StatusCreated)
		resp := client.RunFileUploadRequest{
			Id:  fmt.Sprint(id),
			Url: url,
		}
		json.NewEncoder(w).Encode(resp)
	}

	patchFileInfo := func(w http.ResponseWriter, r *http.Request) {
		fileId, err := strconv.Atoi(r.PathValue("fileId"))
		if err != nil {
			panic(err)
		}

		var info client.UpdateRunFileInfoJSONBody
		mustDecode(r.Body, &info)

		fs.status[fileId] = info.UploadStatus

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(info)
	}

	postS3File := func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fs.files = append(fs.files, body)
		w.WriteHeader(http.StatusOK)
	}

	mux.HandleFunc("POST /api/v1/runs", secure(createRun))

	mux.HandleFunc("POST /api/v1/runs/{runId}/files/upload", secure(postFileUpload))

	mux.HandleFunc("PATCH /api/v1/runs/{runId}/files/{fileId}", secure(patchFileInfo))

	mux.HandleFunc("POST /s3/files/{fileId}", log(postS3File))

	mux.HandleFunc("/", log(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	fs.mux = mux
	fs.server = server

	return fs
}

func (s *FakeServer) useGitHub() {
	s.env["GITHUB_ACTIONS"] = "true"
	s.env["GITHUB_ACTOR"] = "smvv"
	s.env["GITHUB_REF"] = "refs/heads/feature-branch-1"
	s.env["GITHUB_REF_NAME"] = "feature-branch-1"
	s.env["GITHUB_REF_TYPE"] = "branch"
	s.env["GITHUB_REPOSITORY"] = "octocat/Hello-World"
	s.env["GITHUB_RUN_ATTEMPT"] = "1"
	s.env["GITHUB_RUN_ID"] = "1658821493"
	s.env["GITHUB_RUN_NUMBER"] = "3"
	s.env["GITHUB_SHA"] = "ffac537e6cbbf934b08745a378932722df287a53"
}

func (s *FakeServer) extractFiles(i int) (map[string][]byte, error) {
	file := s.files[i]
	r := bytes.NewReader(file)

	var buf bytes.Buffer
	if err := decompressZstd(r, &buf); err != nil {
		return nil, err
	}

	return extractTarball(&buf)
}

func mustDecode(r io.ReadCloser, v interface{}) {
	defer r.Close()
	err := json.NewDecoder(r).Decode(&v)
	if err != nil {
		panic(err)
	}
}
