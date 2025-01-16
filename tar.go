package main

import (
	"archive/tar"
	"bytes"
	"fmt"
)

// createTarball creates a tarball from a map of file names and their contents.
// The tarball data is written into out.
func createTarball(files map[string][]byte, out *bytes.Buffer) error {
	tw := tar.NewWriter(out)

	for name, content := range files {
		header := &tar.Header{
			Name: name,
			Mode: 0600,
			Size: int64(len(content)),
		}

		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header: %w", err)
		}

		if _, err := tw.Write(content); err != nil {
			return fmt.Errorf("failed to write file content: %w", err)
		}
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}

	return nil
}
