package record

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net"
	"net/http"
	"sync"
	"time"
)

const initialBackoffDuration = 500 * time.Millisecond

// retryTransport implements http.RoundTripper and adds retry logic.
type retryTransport struct {
	base       http.RoundTripper
	maxRetries int

	sleep func(d time.Duration)

	log *slog.Logger

	mu   sync.Mutex
	rand *rand.Rand
}

func (t *retryTransport) randInt64(backoff int64) int64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.rand == nil {
		s := rand.NewPCG(42, uint64(time.Now().UnixNano()))
		t.rand = rand.New(s)
	}

	return t.rand.Int64N(backoff)
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	sync.OnceFunc(func() {
		if t.base == nil {
			t.base = &http.Transport{
				// Use most values from `net/http/transport.go`, but with a
				// shorter dialer timeout (from 30 to 10 sec).
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			}
		}

		if t.sleep == nil {
			t.sleep = time.Sleep
		}

		if t.log == nil {
			t.log = slog.Default()
		}
	})()

	backoff := initialBackoffDuration

	var resp *http.Response
	var err error

	retries := t.maxRetries

	for i := 0; i < retries; i++ {
		resp, err = t.base.RoundTrip(req)
		if err != nil {
			// Could be a network error, DNS issue, etc.—retry
			t.log.Info("request failed with error",
				"method", req.Method,
				"path", req.URL.Path,
				"attempt", i+1,
				"err", err,
			)
		} else if resp.StatusCode < 500 {
			// Return on success or any non-5xx code
			return resp, nil
		} else {
			// If 5xx, we want to retry. Close response body to avoid leaks.
			resp.Body.Close()
			t.log.Info("request failed with server error",
				"method", req.Method,
				"path", req.URL.Path,
				"attempt", i+1,
				"status", resp.StatusCode,
			)
		}

		// Apply exponential backoff with jitter
		jitter := time.Duration(t.randInt64(int64(backoff / 2)))
		sleepDuration := backoff + jitter
		t.sleep(sleepDuration)
		backoff *= 2
	}

	// If we exhausted all retries, return the last error (or a custom one)
	if err == nil {
		return nil, fmt.Errorf("all retry attempts failed with 5xx status codes")
	}
	return nil, err
}
