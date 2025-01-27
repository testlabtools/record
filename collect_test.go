package record

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/assert"
	"github.com/testlabtools/record/client"
	"github.com/testlabtools/record/fake"
	"github.com/testlabtools/record/tar"
	"github.com/testlabtools/record/zstd"
)

func TestCollectorAddsCommitFiles(t *testing.T) {
	var tests = []struct {
		name    string
		branch  string
		created bool
		added   bool
	}{
		{
			name:    "main-created",
			branch:  "main",
			created: true,
			added:   true,
		},
		{
			name:    "main-second",
			branch:  "main",
			created: false,
			added:   false,
		},
		{
			name:    "feature-created",
			branch:  "feature",
			created: true,
			added:   false,
		},
		{
			name:    "feature-second",
			branch:  "feature",
			created: false,
			added:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := slogt.New(t)
			assert := assert.New(t)

			srv := fake.NewServer(t, l, client.Github)
			defer srv.Close()

			options := UploadOptions{
				Reports: "testdata/github/reports",
				Repo:    "testdata/github/repo",
			}

			srv.Env["GITHUB_REF"] = "refs/head/" + tt.branch
			srv.Env["GITHUB_REF_NAME"] = tt.branch

			collector, err := NewCollector(l, options, srv.Env)
			if !assert.NoError(err) {
				return
			}

			var data bytes.Buffer
			err = collector.Bundle(tt.created, &data)
			if !assert.NoError(err) {
				return
			}

			var buf bytes.Buffer
			err = zstd.Decompress(&data, &buf)
			if !assert.NoError(err) {
				return
			}

			files, err := tar.Extract(&buf)
			if !assert.NoError(err) {
				return
			}

			file := files[GitSummaryFileName]
			if !tt.created {
				assert.Empty(file)
				return
			}

			var summary GitSummary
			err = json.NewDecoder(bytes.NewReader(file)).Decode(&summary)
			assert.NoError(err)

			if tt.added {
				assert.NotEmpty(summary.CommitFiles)
			} else {
				assert.Empty(summary.CommitFiles)
			}
			assert.NotEmpty(summary.DiffStat)
		})
	}
}
