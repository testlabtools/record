package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

// uploadFile uploads the compressed data to the specified URL.
func uploadFile(ctx context.Context, url string, compressedData *bytes.Buffer) error {
	req, err := http.NewRequestWithContext(ctx, "POST", url, compressedData)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/zstd")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s", string(body))
	}

	return nil
}
