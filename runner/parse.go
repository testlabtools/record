package runner

import (
	"fmt"
	"io"
)

type ParserOptions struct {
	WorkDir string
}

type Parser interface {
	Parse(r io.Reader) error
	Format(w io.Writer) error
}

var parsers = map[string]func(o ParserOptions) Parser{
	"go-test": NewGoTest,
	"jest":    NewJest,
}

func New(name string, o ParserOptions) (Parser, error) {
	p := parsers[name]
	if p == nil {
		return nil, fmt.Errorf("unknown runner format: %q", name)
	}
	return p(o), nil
}
