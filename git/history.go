package git

import (
	"bufio"
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
	cmd := exec.Command(
		"git",
		"-C", r.dir,
		"log",
		fmt.Sprintf("--since=%ddays", r.MaxDays),
		"--name-only",
		"--pretty=format:commit %H %cs",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Use a single channel for both results and errors
	type parseResult struct {
		parsed []CommitFile
		err    error
	}
	resultChan := make(chan parseResult, 1)

	// Parse the command output concurrently
	go func() {
		parsed, err := parseCommitFiles(stdout)
		resultChan <- parseResult{parsed: parsed, err: err}
		close(resultChan)
	}()

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("error waiting for command: %w", err)
	}

	res := <-resultChan
	return res.parsed, res.err
}
