package git

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
)

type Repo struct {
	dir string

	mainBranch string

	MaxDays int
}

func NewRepo(dir string) *Repo {
	return &Repo{
		dir: dir,

		MaxDays: 60,
	}
}

func (r Repo) MainBranch() (string, error) {
	if r.mainBranch != "" {
		return r.mainBranch, nil
	}

	cmd := exec.Command(
		"git",
		"-C", r.dir,
		"branch",
		"-r",
	)

	out, err := cmd.Output()

	if err != nil {
		return "", err
	}

	branch := ""

	branches := []string{
		"origin/main",
		"origin/master",
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if slices.Contains(branches, line) {
			branch = line
		}
	}

	if branch == "" {
		return "", fmt.Errorf("cannot find main branch in git remote output")
	}

	r.mainBranch = branch

	return branch, nil
}
