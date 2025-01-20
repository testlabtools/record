package git

import (
	"os/exec"
	"strings"
)

type CommitInfo struct {
	AuthorEmail string
	Subject     string
}

func (r *Repo) CommitInfo(ref string) (*CommitInfo, error) {
	cmd := exec.Command(
		"git",
		"-C", r.Dir,
		"log",
		"-1",
		"--format=%ae%x09%s",
		ref,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	line := strings.TrimSpace(string(out))
	if line == "" {
		return nil, nil
	}

	fields := strings.SplitN(line, "\t", 2)

	return &CommitInfo{
		AuthorEmail: fields[0],
		Subject:     fields[1],
	}, nil
}
