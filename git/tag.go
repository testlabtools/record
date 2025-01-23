package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func (r *Repo) TagsPointedAt(ref string) ([]string, error) {
	args := []string{
		"-C", r.Dir,
		"tag",
		"--points-at", ref,
	}

	cmd := exec.Command("git", args...)
	out, err := cmd.Output()

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err != nil {
		return nil, fmt.Errorf("failed to get tags pointed at for args %q: stderr=%q err=%w", args, stderr.String(), err)
	}
	lines := strings.TrimSpace(string(out))
	if lines == "" {
		return nil, nil
	}
	return strings.Split(lines, "\n"), nil
}
