package git

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Date(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestParseCommitFiles(t *testing.T) {
	assert := assert.New(t)

	log := strings.NewReader(`
commit 2024-10-16
app/loader.ts

commit 2024-09-24
app/loader.ts
packages/foo/bar.ts

commit 2024-09-24
app/baz.ts

commit 2024-09-23
app/baz.ts
packages/quux/some.test.ts
`)

	expected := []CommitFile{
		{
			Committed: Date("2024-10-16"),
			Names:     []string{"app/loader.ts"},
		}, {
			Committed: Date("2024-09-24"),
			Names:     []string{"app/loader.ts", "packages/foo/bar.ts"},
		}, {
			Committed: Date("2024-09-24"),
			Names:     []string{"app/baz.ts"},
		}, {
			Committed: Date("2024-09-23"),
			Names:     []string{"app/baz.ts", "packages/quux/some.test.ts"},
		},
	}

	commits, err := parseCommitFiles(log)
	if !assert.NoError(err) {
		return
	}

	assert.Equal(expected, commits)
}

func TestCommitFiles(t *testing.T) {
	assert := assert.New(t)

	r := NewRepo("../testdata/github/repo")
	commits, err := r.CommitFiles()
	if !assert.NoError(err) {
		return
	}

	today := Date("2024-10-16")

	expected := []CommitFile{
		{
			Committed: today,
			Names:     []string{".github/CODEOWNERS"},
		},
	}

	// Reset date to today to stabilize test.
	for i := range commits {
		assert.NotEmpty(commits[i].Committed)
		commits[i].Committed = today
	}

	assert.Equal(expected, commits)
}
