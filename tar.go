package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
)

// createTarball creates a tarball from a map of file names and their contents.
// The tarball data is written into out.
func createTarball(files map[string][]byte, out io.Writer) error {
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

// extractTarball extracts a tarball into a map of file names and their
// contents.
func extractTarball(r io.Reader) (map[string][]byte, error) {
	files := make(map[string][]byte)
	tr := tar.NewReader(r)

	for {
		// Read the next header
		header, err := tr.Next()
		if err == io.EOF {
			// End of tarball
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tarball: %w", err)
		}

		// Only process regular files (skip directories, symlinks, etc.)
		if header.Typeflag != tar.TypeReg {
			continue
		}

		// Read the file content
		content := new(bytes.Buffer)
		if _, err := io.Copy(content, tr); err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", header.Name, err)
		}

		// Add the file content to the map
		files[header.Name] = content.Bytes()
	}

	return files, nil
}
