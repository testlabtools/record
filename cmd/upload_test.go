package main

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/assert"
	"github.com/testlabtools/record/client"
	"github.com/testlabtools/record/fake"
)

func TestParseStarted(t *testing.T) {
	// TODO support "2016-07-25T02:22:33Z"

	assert := assert.New(t)
	val, err := parseStarted("2016-07-25T02:22:33+0000")
	if !assert.NoError(err) {
		return
	}
	assert.Equal(time.Date(2016, 7, 25, 2, 22, 33, 0, time.FixedZone("", 0)), val)
}

func TestUploadCommand(t *testing.T) {
	var tests = []struct {
		name  string
		args  []string
		check func(t *testing.T, srv *fake.FakeServer)
	}{
		{
			name: "default",
		},
		{
			name: "explicit started",
			args: []string{
				"--started", "2016-07-25T02:22:33+0000",
			},
			check: func(t *testing.T, srv *fake.FakeServer) {
				key := srv.Env["GITHUB_RUN_ID"] + "-" + srv.Env["TESTLAB_GROUP"]
				if !assert.Contains(t, srv.Runs, key) {
					return
				}

				started := time.Date(2016, 7, 25, 2, 22, 33, 0, time.UTC)
				run := srv.Runs[key]
				assert.Equal(t, started, *run.Started)
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
