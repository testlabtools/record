package git

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

type Repo struct {
	Dir string

	mainBranch string

	MaxDays int
}

func NewRepo(dir string) *Repo {
	return &Repo{
		Dir: dir,

		MaxDays: 60,
	}
}

func (r Repo) Exists() bool {
	info, err := os.Stat(r.Dir)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (r Repo) MainBranch() (string, error) {
	if r.mainBranch != "" {
		return r.mainBranch, nil
	}

	cmd := exec.Command(
		"git",
		"-C", r.Dir,
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
