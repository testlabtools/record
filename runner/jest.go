package runner

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

type Jest struct {
	options ParserOptions
	tests   []string
}

type JestTestOutput struct {
	TestMatch []string `json:"testMatch,omitempty"`
}

func NewJest(o ParserOptions) Parser {
	return &Jest{
		options: o,
	}
}

func (p *Jest) Parse(r io.Reader) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		// Omit any lines that don't start with a slash. They could be jset
		// warnings (like jest-haste-map complaining about a duplicated manual
		// mock found).
		if !strings.HasPrefix(line, "/") {
			continue
		}
		p.tests = append(p.tests, line)
	}

	return scanner.Err()
}

func (p *Jest) Format(w io.Writer) error {
	var matches []string

	for _, t := range p.tests {
		match := strings.TrimPrefix(t, p.options.WorkDir)
		matches = append(matches, match)
	}

	o := JestTestOutput{
		TestMatch: matches,
	}
	return json.NewEncoder(w).Encode(o)
}
