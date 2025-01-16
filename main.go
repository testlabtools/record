package main

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/klauspost/compress/zstd"
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

func main() {
	files := map[string][]byte{
		"file1.txt": []byte("This is the content of file1."),
		"file2.txt": []byte("This is the content of file2."),
	}

	var raw bytes.Buffer
	var buf bytes.Buffer

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := createTarball(files, &raw); err != nil {
		log.Fatalf("Error creating tarball: %v", err)
	}

	if err := compressZstd(&raw, &buf); err != nil {
		log.Fatalf("Error compressing tarball: %v", err)
	}

	log.Printf("initial size: %d bytes; compressed size: %d bytes\n", raw.Len(), buf.Len())

	// TODO
	url := "https://example.com/upload"
	if err := uploadFile(ctx, url, &buf); err != nil {
		log.Fatalf("Error uploading compressed data: %v", err)
	}

	fmt.Println("Upload successful!")
}
