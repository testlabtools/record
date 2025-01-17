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
	Committed time.Time `json:"committed"`
	Names     []string  `json:"names"`
}

func parseCommitFiles(r io.Reader) ([]CommitFile, error) {
	var result []CommitFile
	var currentCommit CommitFile
	var inCommit bool

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "commit ") {
			if inCommit {
				// Finish the current commit
				result = append(result, currentCommit)
				currentCommit = CommitFile{}
			}

			commitDate := strings.TrimPrefix(line, "commit ")
			parsedTime, err := time.Parse("2006-01-02", commitDate)
			if err != nil {
				return nil, fmt.Errorf("invalid date format: %s", commitDate)
			}
			currentCommit.Committed = parsedTime
			inCommit = true
		} else if line != "" {
			// Collect file names
			currentCommit.Names = append(currentCommit.Names, line)
		}
	}

	if inCommit {
		// Append the last commit if it exists
		result = append(result, currentCommit)
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
		"--pretty=format:commit %cs",
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
