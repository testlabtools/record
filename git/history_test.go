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
commit abcdef1234 2024-10-16
app/loader.ts

commit ab12cd34ef 2024-09-24
app/loader.ts
packages/foo/bar.ts

commit abc1234def 2024-09-24
app/baz.ts

commit def1234abc 2024-09-23
app/baz.ts
packages/quux/some.test.ts
`)

	expected := []CommitFile{
		{
			Hash:      "abcdef1234",
			Committed: Date("2024-10-16"),
			Names:     []string{"app/loader.ts"},
		}, {
			Hash:      "ab12cd34ef",
			Committed: Date("2024-09-24"),
			Names:     []string{"app/loader.ts", "packages/foo/bar.ts"},
		}, {
			Hash:      "abc1234def",
			Committed: Date("2024-09-24"),
			Names:     []string{"app/baz.ts"},
		}, {
			Hash:      "def1234abc",
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
	if !assert.True(r.Exists()) {
		return
	}

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

	// Reset date to today and commit hash to stabilize test.
	for i := range commits {
		assert.NotEmpty(commits[i].Committed)
		commits[i].Committed = today

		assert.NotEmpty(commits[i].Hash)
		expected[i].Hash = commits[i].Hash
	}

	assert.Equal(expected, commits)
}
