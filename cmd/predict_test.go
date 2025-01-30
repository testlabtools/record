package cmd

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/assert"
	"github.com/testlabtools/record/client"
	"github.com/testlabtools/record/fake"
)

func TestPredictCommand(t *testing.T) {
	var tests = []struct {
		name  string
		args  []string
		check func(t *testing.T, srv *fake.FakeServer)
	}{
		{
			name: "default",
			check: func(t *testing.T, srv *fake.FakeServer) {
				// TODO
				assert.Empty(t, srv.Files)
			},
		},
		{
			name: "github",
			args: []string{
				"--repo", "../testdata/github/repo",
			},
			check: func(t *testing.T, srv *fake.FakeServer) {
				// TODO
				assert.Empty(t, srv.Files)
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

			os.Args = append([]string{"record", "predict"}, tt.args...)

			err := predictCmd.ExecuteContext(ctx)
			if !assert.NoError(err) {
				return
			}

			if tt.check != nil {
				tt.check(t, srv)
			}
		})
	}
}
