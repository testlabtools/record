package record

import (
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
		expected []string
		err      string
	}{
		{
			name: "default",
			options: UploadOptions{
				Reports: "testdata/basic/reports",
			},
			expected: []string{
				"testdata/basic/reports/e2e-1.xml",
				"testdata/basic/reports/e2e-2.xml",
			},
		},
		{
			name: "empty reports",
			options: UploadOptions{
				Reports: "testdata/unknown/reports",
			},
			expected: []string{},
		},
		{
			name: "empty key",
			env:  map[string]string{},
			err:  "env var TESTLAB_KEY is required",
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

func mustReadFiles(files []string) map[string][]byte {
	contents := make(map[string][]byte)
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		content, err := io.ReadAll(f)
		if err != nil {
			panic(err)
		}
		contents[file] = content
	}
	return contents
}
