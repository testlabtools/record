package zstd

import (
	"bytes"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
)

// Compress compresses data using Zstd into the writer.
func Compress(data *bytes.Buffer, w io.Writer) error {
	z, err := zstd.NewWriter(w)
	if err != nil {
		return fmt.Errorf("failed to create Zstd writer: %w", err)
	}

	if _, err := z.Write(data.Bytes()); err != nil {
		return fmt.Errorf("failed to write compressed content: %w", err)
	}

	if err := z.Close(); err != nil {
		return fmt.Errorf("failed to close Zstd writer: %w", err)
	}

	return nil
}

// Decompress decompresses compressed data using Zstd into the writer.
func Decompress(r io.Reader, w io.Writer) error {
	z, err := zstd.NewReader(r)
	if err != nil {
		return fmt.Errorf("failed to create Zstd reader: %w", err)
	}

	if _, err := z.WriteTo(w); err != nil {
		return fmt.Errorf("failed to read compressed content: %w", err)
	}

	z.Close()

	return nil
}
