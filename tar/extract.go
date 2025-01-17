package tar

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
)

// Extract extracts a tarball into a map of file names and their
// contents.
func Extract(r io.Reader) (map[string][]byte, error) {
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
