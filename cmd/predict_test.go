package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/assert"
	"github.com/testlabtools/record/client"
	"github.com/testlabtools/record/fake"
	"github.com/testlabtools/record/runner"
)

func TestPredictCommand(t *testing.T) {
	var tests = []struct {
		name   string
		args   []string
		stdin  string
		stdout interface{}
		check  func(t *testing.T, srv *fake.FakeServer)
	}{
		{
			name: "feature-empty-go-test",
			args: []string{
				"--repo", "../testdata/feature/repo",
				"--runner", "go-test",
			},
			stdout: "^()$",
			check: func(t *testing.T, srv *fake.FakeServer) {
				assert.Empty(t, srv.Predicts)
			},
		},
		{
			name: "github-two-go-test",
			args: []string{
				"--repo", "../testdata/github/repo",
				"--runner", "go-test",
			},
			stdin: `ok  	github.com/testlabtools/record
TestPredictCommand
TestUploadCommand
`,
			stdout: "^(TestPredictCommand|TestUploadCommand)$",
			check: func(t *testing.T, srv *fake.FakeServer) {
				assert.Len(t, srv.Predicts, 1)
				assert.Len(t, srv.Predicts[0].TestFiles, 2)
			},
		},
		{
			name: "github-two-jest",
			args: []string{
				"--repo", "../testdata/github/repo",
				"--runner", "jest",
			},
			stdin: `jest-haste-map: duplicated manual mock found: Foo
  The following files share their name; please delete one of them:
    * <rootDir>/src/foo/__mocks__/Foo.ts
    * <rootDir>/src/bar/__mocks__/Foo.ts
$pwd$/app/web/baz.test.ts
$pwd$/app/web/quux.test.ts
`,
			stdout: runner.JestTestOutput{
				TestMatch: []string{
					"/app/web/baz.test.ts",
					"/app/web/quux.test.ts",
				},
			},
			check: func(t *testing.T, srv *fake.FakeServer) {
				assert.Len(t, srv.Predicts, 1)
				assert.Len(t, srv.Predicts[0].TestFiles, 2)
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

			cwd, _ := os.Getwd()
			cwd = path.Join(cwd, "testdata", "symlink")
			srv.Env["PWD"] = cwd

			ctx := context.WithValue(context.Background(), "env", srv.Env)

			stdin := strings.ReplaceAll(tt.stdin, "$pwd$", cwd)
			ctx = context.WithValue(ctx, "stdin", strings.NewReader(stdin))

			var stdout bytes.Buffer
			ctx = context.WithValue(ctx, "stdout", &stdout)

			os.Args = append([]string{"record", "predict"}, tt.args...)

			err := predictCmd.ExecuteContext(ctx)
			if !assert.NoError(err) {
				return
			}

			if tt.check != nil {
				tt.check(t, srv)
			}

			if _, ok := tt.stdout.(string); ok {
				assert.Equal(tt.stdout, stdout.String())
			} else {
				var buf bytes.Buffer
				json.NewEncoder(&buf).Encode(tt.stdout)
				assert.JSONEq(buf.String(), stdout.String())
			}
		})
	}
}
