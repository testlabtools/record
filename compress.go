package main

import (
	"bytes"
	"fmt"

	"github.com/klauspost/compress/zstd"
)

// compressZstd compresses data using Zstd and writes it into out.
func compressZstd(data *bytes.Buffer, out *bytes.Buffer) error {
	w, err := zstd.NewWriter(out)
	if err != nil {
		return fmt.Errorf("failed to create Zstd writer: %w", err)
	}

	if _, err := w.Write(data.Bytes()); err != nil {
		return fmt.Errorf("failed to write compressed content: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close Zstd writer: %w", err)
	}

	return nil
}
