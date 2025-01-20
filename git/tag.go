package git

import (
	"os/exec"
	"strings"
)

func (r *Repo) TagsPointedAt(ref string) ([]string, error) {
	cmd := exec.Command(
		"git",
		"-C", r.dir,
		"tag",
		"--points-at", ref,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.TrimSpace(string(out))
	return strings.Split(lines, "\n"), nil
}
