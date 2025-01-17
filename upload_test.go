package record

import (
	"testing"

	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/assert"

	"github.com/testlabtools/record/client"
)

func TestUploadFromGithub(t *testing.T) {
	var tests = []struct {
		name     string
		options  UploadOptions
		env      map[string]string
		expected map[string][]byte
		err      string
	}{
		{
			name: "default",
			options: UploadOptions{
				Repo: "simple",
			},
			expected: map[string][]byte{
				"file1.txt": []byte("This is the content of file1."),
				"file2.txt": []byte("This is the content of file2."),
			},
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

			srv := newFakeServer(t, l, client.Github)
			defer srv.Close()

			env := srv.env
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
				assert.Len(srv.files, 1)
				files, err := srv.extractFiles(0)
				assert.NoError(err)

				assert.Equal(tt.expected, files)
			} else {
				assert.Empty(srv.files)
			}
		})
	}
}
