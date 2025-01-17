package git

import (
	"bufio"
	"fmt"
	"io"
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

func parseGitDiff(r io.Reader) (*DiffStat, error) {
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
		} else if strings.Contains(line, "files changed") {
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
