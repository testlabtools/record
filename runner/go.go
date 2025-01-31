package runner

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type GoTest struct {
	options ParserOptions
	tests   []string
}

func NewGoTest(o ParserOptions) Parser {
	return &GoTest{
		options: o,
	}
}

func (p *GoTest) Parse(r io.Reader) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		// Omit any lines with spaces (containing go build output).
		if strings.Contains(line, " ") {
			continue
		}
		p.tests = append(p.tests, line)
	}

	return scanner.Err()
}

func (p *GoTest) Format(w io.Writer) error {
	// Create a regexp pattern as test format.
	out := strings.Join(p.tests, "|")
	_, err := fmt.Fprintf(w, "^(%s)$", out)
	return err
}
