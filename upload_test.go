package record

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/assert"

	"github.com/testlabtools/record/client"
	"github.com/testlabtools/record/fake"
)

func TestUploadFromGithub(t *testing.T) {
	var tests = []struct {
		name     string
		options  UploadOptions
		env      map[string]string
		expected map[string]string
		err      string
	}{
		{
			name: "default",
			options: UploadOptions{
				Reports: "testdata/basic/reports",
			},
			expected: map[string]string{
				"testdata/basic/reports/e2e-1.xml": "reports/1.xml",
				"testdata/basic/reports/e2e-2.xml": "reports/2.xml",
			},
		},
		{
			name: "github",
			options: UploadOptions{
				Reports: "testdata/github/reports",
				Repo:    "testdata/github/repo",
			},
			expected: map[string]string{
				"testdata/github/reports/e2e-1.xml":       "reports/1.xml",
				"testdata/github/reports/e2e-2.xml":       "reports/2.xml",
				"testdata/github/repo/.github/CODEOWNERS": "CODEOWNERS",
			},
		},
		{
			name: "empty reports",
			options: UploadOptions{
				Reports: "testdata/unknown/reports",
			},
			expected: map[string]string{},
		},
		{
			name: "empty key",
			env:  map[string]string{},
			err:  "env var TESTLAB_KEY is required",
		},
		{
			name: "too many reports",
			options: UploadOptions{
				Reports:    "testdata/basic/reports",
				MaxReports: 1,
			},
			err: "too many files (2 > 1) found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := slogt.New(t)
			assert := assert.New(t)

			srv := fake.NewServer(t, l, client.Github)
			defer srv.Close()

			env := srv.Env
			if tt.env != nil {
				env = tt.env
			}
			err := Upload(l, env, tt.options)

			if tt.err != "" {
				assert.ErrorContains(err, tt.err)
			} else {
				if !assert.NoError(err) {
					return
				}
			}

			if len(tt.expected) > 0 {
				assert.Len(srv.Files, 1)
				files, err := srv.ExtractTar(0)
				assert.NoError(err)

				expected := mustReadFiles(tt.expected)
				assert.Equal(expected, files)
			} else {
				assert.Empty(srv.Files)
			}
		})
	}
}

func mustReadFiles(files map[string]string) map[string][]byte {
	contents := make(map[string][]byte)
	for file, key := range files {
		f, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		content, err := io.ReadAll(f)
		if err != nil {
			panic(err)
		}
		contents[key] = content
	}
	return contents
}

func TestUploadSkipsFilesForSameRun(t *testing.T) {
	var tests = []struct {
		name     string
		options  UploadOptions
		expected map[string]string
	}{
		{
			name: "default",
			options: UploadOptions{
				Reports: "testdata/basic/reports",
			},
			expected: map[string]string{
				"testdata/basic/reports/e2e-1.xml": "reports/1.xml",
				"testdata/basic/reports/e2e-2.xml": "reports/2.xml",
			},
		},
		{
			name: "github",
			options: UploadOptions{
				Reports: "testdata/github/reports",
				Repo:    "testdata/github/repo",
			},
			expected: map[string]string{
				"testdata/github/reports/e2e-1.xml": "reports/1.xml",
				"testdata/github/reports/e2e-2.xml": "reports/2.xml",
				// CODEOWNERS is skipped for non-initial run uploads.
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := slogt.New(t)
			assert := assert.New(t)

			srv := fake.NewServer(t, l, client.Github)
			defer srv.Close()

			// Fake initial run creation.
			runKey := fmt.Sprintf("%s-e2e", srv.Env["GITHUB_RUN_ID"])
			srv.Runs[runKey] = client.CIRunRequest{}

			err := Upload(l, srv.Env, tt.options)
			if !assert.NoError(err) {
				return
			}

			assert.Len(srv.Files, 1)
			files, err := srv.ExtractTar(0)
			assert.NoError(err)

			expected := mustReadFiles(tt.expected)
			assert.Equal(expected, files)
		})
	}
}
