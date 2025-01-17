package git

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGitDiff(t *testing.T) {

	var tests = []struct {
		name  string
		input string
		stat  *DiffStat
	}{
		{
			name:  "empty",
			input: "",
			stat:  &DiffStat{},
		},
		{
			name: "changes",
			input: `commit abcdef1234
1	1	first/foo.go
6	5	second/bar.go
-	-	third/baz.bin
3 files changed, 7 insertions(+), 6 deletions(-)`,
			stat: &DiffStat{
				Commit: "abcdef1234",
				Changes: []FileChange{
					{
						Insertions: 1,
						Deletions:  1,
						Name:       "first/foo.go",
					},
					{
						Insertions: 6,
						Deletions:  5,
						Name:       "second/bar.go",
					},
					{
						Name: "third/baz.bin",
					},
				},
				Files:      3,
				Insertions: 7,
				Deletions:  6,
			},
		},
		{
			name: "insertions",
			input: `commit abcdef1234
1	0	first/foo.go
-	-	third/baz.bin
2 files changed, 1 insertions(+)`,
			stat: &DiffStat{
				Commit: "abcdef1234",
				Changes: []FileChange{
					{
						Insertions: 1,
						Name:       "first/foo.go",
					},
					{
						Name: "third/baz.bin",
					},
				},
				Files:      2,
				Insertions: 1,
			},
		},
		{
			name: "deletions",
			input: `commit abcdef1234
0	5	second/bar.go
-	-	third/baz.bin
2 files changed, 5 deletions(-)`,
			stat: &DiffStat{
				Commit: "abcdef1234",
				Changes: []FileChange{
					{
						Deletions: 5,
						Name:      "second/bar.go",
					},
					{
						Name: "third/baz.bin",
					},
				},
				Files:     2,
				Deletions: 5,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			assert := assert.New(t)

			r := strings.NewReader(tt.input)

			stat, err := parseGitDiff(r)
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tt.stat, stat)
		})
	}
}
