package record

import "strings"

// mask replaces the input string with the same number of `X` characters.
func mask(s string) string {
	return strings.Repeat("X", len(s))
}
