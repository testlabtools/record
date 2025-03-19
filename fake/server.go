package fake

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
	"github.com/testlabtools/record/tar"
	"github.com/testlabtools/record/zstd"
)

const HeaderAPIKey = "X-API-Key"

type FakeHandlers struct {
	CreateRun http.HandlerFunc

	PostFileUpload http.HandlerFunc

	PatchFileInfo http.HandlerFunc

	Predict http.HandlerFunc

	PutS3File http.HandlerFunc

	NotFound http.HandlerFunc
}

type FakeServer struct {
	mux    *http.ServeMux
	server *httptest.Server

	Handlers *FakeHandlers

	Env map[string]string

	Runs     map[string]client.CIRunRequest
	Files    [][]byte
	fileUrls []string
	status   map[int]client.FileUploadStatus

	Predicts []client.PredictRequest
}

func (s *FakeServer) Close() {
	s.server.Close()
}

func NewServer(t *testing.T, l *slog.Logger, ci client.CIProviderName) *FakeServer {
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

		Env: env,

		Runs:   make(map[string]client.CreateRunJSONRequestBody),
		status: make(map[int]client.FileUploadStatus),
	}

	switch ci {
	case client.Github:
		fs.useGitHub()
	default:
		panic(fmt.Sprintf("unknown CI provider name: %q", ci))
	}

	log := func(handler *http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			l.Info("got fake request", "method", r.Method, "url", r.URL)
			(*handler)(w, r)
		}
	}

	secure := func(handler *http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Check X-API-Key is present.
			assert.NotEmpty(r.Header.Get(HeaderAPIKey), HeaderAPIKey+" is missing")

			w.Header().Set("Content-Type", "application/json")
			log(handler)(w, r)
		}
	}

	h := &FakeHandlers{}
	fs.Handlers = h

	h.CreateRun = func(w http.ResponseWriter, r *http.Request) {
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
		_, ok := fs.Runs[idx]
		fs.Runs[idx] = run
		created := !ok

		status := http.StatusOK
		if created {
			status = http.StatusCreated
		}

		w.WriteHeader(status)
		resp := client.CIRunResponse{
			Id: fmt.Sprint(len(fs.Runs)),
		}
		mustEncode(w, resp)
	}

	mux.HandleFunc("POST /api/v1/runs", secure(&h.CreateRun))

	h.PostFileUpload = func(w http.ResponseWriter, r *http.Request) {
		id := len(fs.fileUrls) + 1
		url := fmt.Sprintf("%s/s3/files/%d", server.URL, id)
		fs.fileUrls = append(fs.fileUrls, url)

		w.WriteHeader(http.StatusCreated)
		resp := client.RunFileUploadRequest{
			Id:  fmt.Sprint(id),
			Url: url,
		}
		mustEncode(w, resp)
	}

	mux.HandleFunc("POST /api/v1/runs/{runId}/files/upload", secure(&h.PostFileUpload))

	h.PatchFileInfo = func(w http.ResponseWriter, r *http.Request) {
		fileId, err := strconv.Atoi(r.PathValue("fileId"))
		if err != nil {
			panic(err)
		}

		var info client.UpdateRunFileInfoJSONBody
		mustDecode(r.Body, &info)

		fs.status[fileId] = info.UploadStatus

		w.WriteHeader(http.StatusOK)
		mustEncode(w, info)
	}

	mux.HandleFunc("PATCH /api/v1/runs/{runId}/files/{fileId}", secure(&h.PatchFileInfo))

	h.Predict = func(w http.ResponseWriter, r *http.Request) {
		var req client.PredictRequest
		mustDecode(r.Body, &req)

		fs.Predicts = append(fs.Predicts, req)

		assert.NotEmpty(req.TestFiles, "TestFiles")
		assert.NotEmpty(req.CiRun.GitRepo, "GITHUB_REPO")
		assert.NotEmpty(req.CiRun.Group, "TESTLAB_GROUP")
		assert.NotEmpty(req.GitSummary, "GitSummary")

		w.WriteHeader(http.StatusOK)
		resp := client.PredictResponse{
			TestFiles: req.TestFiles,
		}
		mustEncode(w, resp)
	}

	mux.HandleFunc("POST /api/v1/predict", secure(&h.Predict))

	h.PutS3File = func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fs.Files = append(fs.Files, body)
		w.WriteHeader(http.StatusOK)
	}

	mux.HandleFunc("PUT /s3/files/{fileId}", log(&h.PutS3File))

	h.NotFound = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}
	mux.HandleFunc("/", log(&h.NotFound))

	fs.mux = mux
	fs.server = server

	return fs
}

func (s *FakeServer) useGitHub() {
	s.Env["GITHUB_ACTIONS"] = "true"
	s.Env["GITHUB_ACTOR"] = "smvv"
	s.Env["GITHUB_REF"] = "refs/heads/feature-branch-1"
	s.Env["GITHUB_REF_NAME"] = "feature-branch-1"
	s.Env["GITHUB_REF_TYPE"] = "branch"
	s.Env["GITHUB_REPOSITORY"] = "octocat/Hello-World"
	s.Env["GITHUB_RUN_ATTEMPT"] = "1"
	s.Env["GITHUB_RUN_ID"] = "1658821493"
	s.Env["GITHUB_RUN_NUMBER"] = "3"
	s.Env["GITHUB_SHA"] = "ffac537e6cbbf934b08745a378932722df287a53"
}

func (s *FakeServer) ExtractTar(i int) (map[string][]byte, error) {
	file := s.Files[i]
	r := bytes.NewReader(file)

	var buf bytes.Buffer
	if err := zstd.Decompress(r, &buf); err != nil {
		return nil, err
	}

	return tar.Extract(&buf)
}

func mustDecode(r io.ReadCloser, v interface{}) {
	defer r.Close()
	err := json.NewDecoder(r).Decode(&v)
	if err != nil {
		panic(err)
	}
}

func mustEncode(w io.Writer, v interface{}) {
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		panic(err)
	}
}
