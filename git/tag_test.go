package git

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTagsPointedAt(t *testing.T) {
	var tests = []struct {
		repo string
		ref  string
		tags []string
	}{
		{"github", "HEAD", []string{"1.0.2"}},
		{"feature", "HEAD", []string{"1.0.2", "2.my-feature.3"}},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.repo, tt.ref), func(t *testing.T) {
			assert := assert.New(t)

			r := NewRepo(fmt.Sprintf("../testdata/%s/repo", tt.repo))
			tags, err := r.TagsPointedAt(tt.ref)
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tt.tags, tags)
		})
	}
}
