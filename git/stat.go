package git

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

type DiffStat struct {
	Commit     string       `json:"commit"`
	Changes    []FileChange `json:"changes"`
	Files      int          `json:"files"`
	Insertions int          `json:"insertions"`
	Deletions  int          `json:"deletions"`
}

type FileChange struct {
	Insertions int    `json:"insertions"`
	Deletions  int    `json:"deletions"`
	Name       string `json:"name"`
}

func parseDiffStat(r io.Reader) (*DiffStat, error) {
	scanner := bufio.NewScanner(r)
	var diffStat DiffStat

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "commit ") {
			diffStat.Commit = strings.TrimPrefix(line, "commit ")
		} else if strings.Contains(line, "\t") {
			parts := strings.Split(line, "\t")
			if len(parts) != 3 {
				return nil, fmt.Errorf("invalid --numstat line: %s", line)
			}

			insertions, err := parseChangeNumber(parts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid insertions number: %w", err)
			}
			deletions, err := parseChangeNumber(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid deletions number: %w", err)
			}

			diffStat.Changes = append(diffStat.Changes, FileChange{
				Insertions: insertions,
				Deletions:  deletions,
				Name:       parts[2],
			})
		} else if strings.Contains(line, " changed") {
			stats := strings.Fields(line)
			if len(stats) < 3 {
				return nil, fmt.Errorf("invalid --shortstat line: %s", line)
			}

			val, err := strconv.Atoi(stats[0])
			if err != nil {
				return nil, err
			}
			diffStat.Files = val

			for i, stat := range stats {
				if strings.Contains(stat, "insert") {
					val, err := strconv.Atoi(stats[i-1])
					if err != nil {
						return nil, err
					}
					diffStat.Insertions = val
				}
				if strings.Contains(stat, "delet") {
					val, err := strconv.Atoi(stats[i-1])
					if err != nil {
						return nil, err
					}
					diffStat.Deletions = val
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading diff: %w", err)
	}

	return &diffStat, nil
}

func parseChangeNumber(s string) (int, error) {
	if s == "-" {
		return 0, nil // Binary files
	}
	return strconv.Atoi(s)
}

func (r Repo) DiffStat(ref string) (*DiffStat, error) {
	base, err := r.MainBranch()
	if err != nil {
		return nil, fmt.Errorf("cannot find merge-base branch: %w", err)
	}

	// Most git version tags are not merged into the main branch. Use
	// `git-diff` for those tags to get a full diff of the changes.
	diff := []string{
		"-C", r.dir,
		"diff",
		"--merge-base", base,
		"--numstat",
		"--shortstat",
		ref,
	}

	// However, git version tags that are merged into the main branch, return
	// no diff output because they are part of that branch. Use `git-show` to
	// get the diff of those (squashed) changes.
	show := []string{
		"-C", r.dir,
		"show",
		"--format=commit %H",
		"--numstat",
		"--shortstat",
		ref,
	}

	for _, args := range [][]string{diff, show} {
		cmd := exec.Command("git", args...)

		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get diff stat for ref %q: %w", ref, err)
		}

		stat, err := parseDiffStat(bytes.NewReader(out))
		if err != nil {
			return nil, fmt.Errorf("failed to get diff stat for ref %q: %w", ref, err)
		}

		if stat.Commit == "" && stat.Files == 0 {
			// Try next command if stat output is empty.
			continue
		}

		return stat, nil
	}

	return nil, fmt.Errorf("cannot get any diff stat for ref %q", ref)
}
