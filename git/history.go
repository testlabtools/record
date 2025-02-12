package git

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

type CommitFile struct {
	Hash      string    `json:"hash"`
	Committed time.Time `json:"committed"`
	Names     []string  `json:"names"`
}

func parseCommitFiles(r io.Reader) ([]CommitFile, error) {
	var result []CommitFile
	var cur CommitFile
	var inCommit bool

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "commit ") {
			if inCommit {
				// Finish the current commit
				result = append(result, cur)
				cur = CommitFile{}
			}

			fields := strings.Fields(line)
			cur.Hash = fields[1]

			commitDate := fields[2]
			parsedTime, err := time.Parse("2006-01-02", commitDate)
			if err != nil {
				return nil, fmt.Errorf("invalid date format: %s", commitDate)
			}
			cur.Committed = parsedTime
			inCommit = true
		} else if line != "" {
			// Collect file names
			cur.Names = append(cur.Names, line)
		}
	}

	if inCommit {
		// Append the last commit if it exists
		result = append(result, cur)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read log: %v", err)
	}

	return result, nil
}

func (r Repo) CommitFiles() ([]CommitFile, error) {
	args := []string{
		"-C", r.Dir,
		"log",
		fmt.Sprintf("--since=%ddays", r.MaxDays),
		"--name-only",
		"--pretty=format:commit %H %cs",
	}
	cmd := exec.Command("git", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.Output()

	if err != nil {
		return nil, fmt.Errorf("failed to get commit files for args %q: stderr=%q err=%w", args, stderr.String(), err)
	}

	return parseCommitFiles(bytes.NewReader(stdout))
}
