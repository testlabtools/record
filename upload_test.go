package main

import (
	"testing"

	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/assert"

	"github.com/testlabtools/record/client"
)

func TestUpload(t *testing.T) {
	l := slogt.New(t)
	assert := assert.New(t)

	srv := newFakeServer(t, l, client.Github)
	defer srv.Close()

	err := upload(l, srv.env)

	assert.NoError(err)

	assert.Len(srv.files, 1)
	files, err := srv.extractFiles(0)
	assert.NoError(err)

	expected := map[string][]byte{
		"file1.txt": []byte("This is the content of file1."),
		"file2.txt": []byte("This is the content of file2."),
	}
	assert.Equal(expected, files)
}

func TestUploadFailsWithEmptyAPIKey(t *testing.T) {
	l := slogt.New(t)
	assert := assert.New(t)

	osEnv := map[string]string{}

	err := upload(l, osEnv)

	assert.ErrorContains(err, "env var TESTLAB_KEY is required")
}
