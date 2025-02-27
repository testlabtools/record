package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type CommitInfo struct {
	AuthorEmail string
	Subject     string
}

func (r *Repo) CommitInfo(ref string) (*CommitInfo, error) {
	args := []string{
		"-C", r.Dir,
		"log",
		"-1",
		"--format=%ae%x09%s",
		ref,
	}
	cmd := exec.Command("git", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit info for args %q: stderr=%q err=%w", args, stderr.String(), err)
	}

	ci := &CommitInfo{}

	line := strings.TrimSpace(string(out))
	if line == "" {
		return ci, nil
	}

	fields := strings.SplitN(line, "\t", 2)

	ci.AuthorEmail = fields[0]
	ci.Subject = fields[1]

	return ci, nil
}
