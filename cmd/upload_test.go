package main

import (
	"context"
	"log/slog"
	"maps"
	"os"
	"slices"
	"sort"
	"testing"
	"time"

	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/assert"
	"github.com/testlabtools/record/client"
	"github.com/testlabtools/record/fake"
)

func TestParseStarted(t *testing.T) {
	var tests = []struct {
		name string
		val  string
		out  time.Time
	}{
		{
			name: "iso8601",
			val:  "2016-07-25T02:22:33+0000",
			out:  time.Date(2016, 7, 25, 2, 22, 33, 0, time.UTC),
		},
		{
			name: "rfc3339",
			val:  "2016-07-25T02:22:33Z",
			out:  time.Date(2016, 7, 25, 2, 22, 33, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			val, err := parseStarted(tt.val)
			if !assert.NoError(err) {
				return
			}
			assert.Equal(tt.out, val)
		})
	}

}

func TestUploadCommand(t *testing.T) {
	var tests = []struct {
		name  string
		args  []string
		check func(t *testing.T, srv *fake.FakeServer)
	}{
		{
			name: "default",
			check: func(t *testing.T, srv *fake.FakeServer) {
				// No run files for the default upload command.
				assert.Empty(t, srv.Files)
			},
		},
		{
			name: "github",
			args: []string{
				"--reports", "../testdata/github/reports",
				"--repo", "../testdata/github/repo",
			},
			check: func(t *testing.T, srv *fake.FakeServer) {
				assert.Len(t, srv.Files, 1)

				files, err := srv.ExtractTar(0)
				assert.NoError(t, err)

				names := slices.Collect(maps.Keys(files))
				sort.Strings(names)
				expected := []string{
					"CODEOWNERS",
					"reports/1.xml",
					"reports/2.xml",
				}
				assert.Equal(t, expected, names)
			},
		},
		{
			name: "explicit started",
			args: []string{
				"--started", "2016-07-25T02:22:33+0000",
				"--reports", "../testdata/basic/reports",
			},
			check: func(t *testing.T, srv *fake.FakeServer) {
				key := srv.Env["GITHUB_RUN_ID"] + "-" + srv.Env["TESTLAB_GROUP"]
				if !assert.Contains(t, srv.Runs, key) {
					return
				}

				started := time.Date(2016, 7, 25, 2, 22, 33, 0, time.UTC)
				run := srv.Runs[key]
				assert.Equal(t, started, *run.Started)

				assert.Len(t, srv.Files, 1)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			l := slogt.New(t)
			slog.SetDefault(l)

			srv := fake.NewServer(t, l, client.Github)
			defer srv.Close()

			ctx := context.WithValue(context.Background(), "env", srv.Env)

			os.Args = append([]string{"record", "upload"}, tt.args...)

			err := uploadCmd.ExecuteContext(ctx)
			if !assert.NoError(err) {
				return
			}

			if tt.check != nil {
				tt.check(t, srv)
			}
		})
	}
}
