package record

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/testlabtools/record/client"
)

const DefaulMaxReports = 100

type UploadOptions struct {
	// Repo is the path to the git repository directory.
	Repo string

	// Reports is the path to the JUnit reports directory.
	Reports string

	// Started is the start time of the run. If nil, `NOW()` is returned from
	// the API.
	Started *time.Time

	// MaxReports is the maximum number of reports that can be found in the
	// reports directory. If it exceeds the threshold, an error is returned.
	//
	// If omitted (or zero), DefaulMaxReports is used.
	MaxReports int

	// Debug enables verbose log messages. By default (false), only messages
	// with level info are visible.
	Debug bool

	// Client is the used HTTP client for all API requests.
	Client *http.Client
}

// createRun creates a CI run.
func (u *api) createRun(ctx context.Context, body client.CreateRunJSONRequestBody) (*client.CIRunResponse, bool, error) {
	run, err := u.api.CreateRunWithResponse(ctx, body)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create run: %w", err)
	}

	code := run.StatusCode()

	switch code {
	case http.StatusOK:
		return run.JSON200, false, nil
	case http.StatusCreated:
		return run.JSON201, true, nil
	default:
		return nil, false, fmt.Errorf("create run returned invalid status code: %d", run.StatusCode())
	}
}

// uploadRun uploads the data to the pre-signed URL of the run file.
func (u *api) uploadRunFile(ctx context.Context, run *client.CIRunResponse, data io.Reader) error {
	runId := run.Id

	// Get pre-signed url for the new run file.
	upload, err := u.api.GetRunFileUploadUrlWithResponse(ctx, runId)
	if err != nil {
		return fmt.Errorf("failed to get run file upload url: %w", err)
	}

	code := upload.StatusCode()
	var url string
	var fileId string

	switch code {
	case http.StatusCreated:
		fileId = upload.JSON201.Id
		url = upload.JSON201.Url
	default:
		return fmt.Errorf("create run returned invalid status code: %d", code)
	}

	u.log.Debug("got run file upload", "fileId", fileId, "url", url)

	// Upload data to pre-signed url.
	if err := uploadFile(ctx, u.hc, url, data); err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	u.log.Info("upload successful", "fileId", fileId)

	// Mark file upload as completed
	resp, err := u.api.UpdateRunFileInfoWithResponse(ctx, runId, fileId, client.UpdateRunFileInfoJSONRequestBody{
		UploadStatus: client.UploadCompleted,
	})
	if err != nil {
		return fmt.Errorf("failed to update file info %w", err)
	}

	code = resp.StatusCode()
	if code != http.StatusOK {
		return fmt.Errorf("update file info returned invalid status code: %d", code)
	}

	return nil
}

// uploadFile uploads the compressed data to the specified URL.
func uploadFile(ctx context.Context, client *http.Client, url string, data io.Reader) error {
	req, err := http.NewRequestWithContext(ctx, "PUT", url, data)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/zstd")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed: %s", string(body))
	}

	return nil
}

func Upload(l *slog.Logger, osEnv map[string]string, o UploadOptions) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	server := osEnv["TESTLAB_HOST"]
	if server == "" {
		server = "https://eu.testlab.tools"
	}

	apiKey := osEnv["TESTLAB_KEY"]
	if apiKey == "" {
		return fmt.Errorf("env var TESTLAB_KEY is required")
	}

	l.Info("upload run", "server", server, "apiKey", mask(apiKey))

	collector, err := NewCollector(l, o.Repo, osEnv)
	if err != nil {
		return err
	}

	env := collector.Env()
	l.Debug("collected env vars", "env", env)

	api, err := newApi(l, o.Client, server, apiKey)
	if err != nil {
		return err
	}

	runReq := env.RunRequest()
	runReq.Started = o.Started

	run, created, err := api.createRun(ctx, runReq)
	if err != nil {
		return fmt.Errorf("failed to create run: %w", err)
	}

	l.Info("created run", "runId", run.Id, "created", created, "reports", o.Reports)

	var data bytes.Buffer
	if err := collector.Bundle(BundleOptions{
		InitialRun: created,
		ReportsDir: o.Reports,
		MaxReports: o.MaxReports,
	}, &data); err != nil {
		return fmt.Errorf("failed to bundle: %w", err)
	}

	if data.Len() == 0 {
		l.Warn("collected tarball is empty. Skip file upload")
	} else {
		l.Info("tarball compressed", "size", data.Len())

		if err := api.uploadRunFile(ctx, run, &data); err != nil {
			return fmt.Errorf("failed to upload run: %w", err)
		}
	}

	return nil
}
