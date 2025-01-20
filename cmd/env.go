package cmd

import (
	"os"
	"strings"
)

// getEnv extracts environment variables from OS.
func getEnv() map[string]string {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) > 1 {
			env[pair[0]] = pair[1]
		}
	}
	return env
}
