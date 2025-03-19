package record

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/testlabtools/record/client"
)

const HeaderAPIKey = "X-API-Key"

type api struct {
	hc  *http.Client
	api client.ClientWithResponses
	log *slog.Logger
}

func newApi(l *slog.Logger, hc *http.Client, server, apiKey string) (*api, error) {
	cl, err := client.NewClient(server)
	if err != nil {
		return nil, err
	}

	if hc == nil {
		l.Debug("use default http client with retry")
		hc = http.DefaultClient
		hc.Timeout = 10 * time.Second
		hc.Transport = &retryTransport{
			maxRetries: 5,
			log:        l,
		}
	} else {
		l.Debug("use provided http client")
	}

	cl.Client = hc

	cl.RequestEditors = append(cl.RequestEditors, func(ctx context.Context, r *http.Request) error {
		r.Header.Add(HeaderAPIKey, apiKey)
		return nil
	})

	return &api{
		hc:  hc,
		api: client.ClientWithResponses{ClientInterface: cl},
		log: l,
	}, nil
}
