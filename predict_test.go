package record

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/assert"
	"github.com/testlabtools/record/client"
	"github.com/testlabtools/record/fake"
)

func TestPredictFromGithub(t *testing.T) {
	var tests = []struct {
		name     string
		options  PredictOptions
		env      map[string]string
		stdin    string
		expected string
		err      string
	}{
		{
			name: "feature-empty-go-test",
			options: PredictOptions{
				Repo:   "testdata/feature/repo",
				Runner: "go-test",
			},
			expected: "^()$",
		},
		{
			name: "feature-two-go-test",
			options: PredictOptions{
				Repo:   "testdata/feature/repo",
				Runner: "go-test",
			},
			stdin:    "TestAB\nTestCD|TestEF\n",
			expected: "^(TestCD|TestEF)$",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := slogt.New(t)
			assert := assert.New(t)

			srv := fake.NewServer(t, l, client.Github)
			defer srv.Close()

			srv.Handlers.Predict = func(w http.ResponseWriter, r *http.Request) {
				var req client.PredictRequest
				mustDecode(t, r.Body, &req)

				files := req.TestFiles
				if len(files) > 1 {
					files = files[1:]
				}

				w.WriteHeader(http.StatusOK)
				resp := client.PredictResponse{
					TestFiles: files,
				}
				mustEncode(t, w, resp)
			}

			env := srv.Env
			if tt.env != nil {
				env = tt.env
			}

			opt := tt.options
			opt.Stdin = strings.NewReader(tt.stdin)

			var out bytes.Buffer
			opt.Stdout = &out

			err := Predict(l, env, opt)

			if tt.err != "" {
				assert.ErrorContains(err, tt.err)
			} else if !assert.NoError(err) {
				return
			}

			assert.Equal(tt.expected, out.String())
		})
	}
}

func TestPredictFailedFallback(t *testing.T) {
	var tests = []struct {
		name     string
		options  PredictOptions
		handler  http.HandlerFunc
		stdin    string
		expected string
	}{
		{
			name: "empty-fallback",
			options: PredictOptions{
				Repo:   "testdata/feature/repo",
				Runner: "go-test",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.False(t, true, "predict handler called")
			},
			expected: "^()$",
		},
		{
			name: "crash-fallback",
			options: PredictOptions{
				Repo:   "testdata/feature/repo",
				Runner: "go-test",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			stdin:    "TestSome\nTestThing\n",
			expected: "^(TestSome|TestThing)$",
		},
		{
			name: "invalid-fallback",
			options: PredictOptions{
				Repo:   "testdata/feature/repo",
				Runner: "go-test",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				// dont return any response, so an invalid JSON response.
			},
			stdin:    "TestSome\nTestThing\n",
			expected: "^(TestSome|TestThing)$",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := slogt.New(t)
			assert := assert.New(t)

			srv := fake.NewServer(t, l, client.Github)
			defer srv.Close()

			srv.Handlers.Predict = tt.handler

			env := srv.Env

			opt := tt.options
			opt.Stdin = strings.NewReader(tt.stdin)

			var out bytes.Buffer
			opt.Stdout = &out

			// Disable HTTP retry logic.
			opt.client = http.DefaultClient
			opt.client.Transport = nil

			err := Predict(l, env, opt)
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tt.expected, out.String())
		})
	}
}

func mustDecode(t *testing.T, r io.ReadCloser, v interface{}) {
	t.Helper()
	defer r.Close()
	err := json.NewDecoder(r).Decode(&v)
	if err != nil {
		panic(err)
	}
}

func mustEncode(t *testing.T, w io.Writer, v interface{}) {
	t.Helper()
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		panic(err)
	}
}
