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
		// Omit any lines that don't start with a slash. They could be jest
		// warnings (like jest-haste-map complaining about a duplicated manual
		// mock found).
		if !strings.HasPrefix(line, "/") {
			continue
		}
		file, err := parseFile(line, p.options)
		if err != nil {
			return err
		}
		p.tests = append(p.tests, file)
	}

	return scanner.Err()
}

func (p *Jest) Format(files []string, w io.Writer) error {
	var matches []string

	for _, t := range files {
		matches = append(matches, t)
	}

	o := JestTestOutput{
		TestMatch: matches,
	}
	return json.NewEncoder(w).Encode(o)
}

func (p *Jest) Files() []string {
	return p.tests
}
