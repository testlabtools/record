package git

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommitInfo(t *testing.T) {
	ae := "user1@org"

	var tests = []struct {
		repo string
		ref  string
		info *CommitInfo
	}{
		{"github", "HEAD", &CommitInfo{AuthorEmail: ae}},
		{"feature", "HEAD", &CommitInfo{AuthorEmail: ae}},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.repo, tt.ref), func(t *testing.T) {
			assert := assert.New(t)

			r := NewRepo(fmt.Sprintf("../testdata/%s/repo", tt.repo))
			info, err := r.CommitInfo(tt.ref)
			if !assert.NoError(err) {
				return
			}

			// Copy subject to stabilize test.
			assert.NotEmpty(info.Subject)
			tt.info.Subject = info.Subject

			assert.Equal(tt.info, info)
		})
	}
}
