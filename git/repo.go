package git

import (
	"bytes"
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

	args := []string{
		"-C", r.Dir,
		"branch",
		"-r",
	}

	cmd := exec.Command("git", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("failed to get main branch with args %q: stderr=%q err=%w", args, stderr.String(), err)
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
		return "", fmt.Errorf("cannot find main branch in git remote output: %q", string(out))
	}

	r.mainBranch = branch

	return branch, nil
}

func (r Repo) MergeBase(ref, main string) (string, error) {
	args := []string{
		"-C", r.Dir,
		"merge-base",
		ref,
		main,
	}

	cmd := exec.Command("git", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get merge base with args %q: stderr=%q err=%w", args, stderr.String(), err)
	}

	return strings.TrimSpace(string(out)), nil
}
