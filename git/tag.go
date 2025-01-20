package git

import (
	"os/exec"
	"strings"
)

func (r *Repo) TagsPointedAt(ref string) ([]string, error) {
	cmd := exec.Command(
		"git",
		"-C", r.Dir,
		"tag",
		"--points-at", ref,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.TrimSpace(string(out))
	if lines == "" {
		return nil, nil
	}
	return strings.Split(lines, "\n"), nil
}
