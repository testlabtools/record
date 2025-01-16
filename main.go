package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"
)

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
