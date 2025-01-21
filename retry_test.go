package record

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/assert"
)

func sleepNoop(t *testing.T) func(time.Duration) {
	return func(d time.Duration) {
		assert.GreaterOrEqual(t, d, initialBackoffDuration)
	}
}

// mockRoundTripper allows us to simulate multiple responses/errors.
type mockRoundTripper struct {
	responses []*http.Response
	errors    []error
	callCount int
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	defer func() { m.callCount++ }()

	// If we've used up all the provided responses/errors,
	// just return the last one for each subsequent call
	idx := m.callCount
	if idx >= len(m.responses) {
		// If we have at least one response, return the last one
		if len(m.responses) > 0 {
			return m.responses[len(m.responses)-1], m.errors[len(m.errors)-1]
		}
		// Otherwise nothing is defined
		return nil, nil
	}
	return m.responses[idx], m.errors[idx]
}

// helper function to create a simple http.Response with a given status
func newHTTPResponse(statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(nil),
	}
}

func TestNoRetryOnSuccess(t *testing.T) {
	assert := assert.New(t)

	l := slogt.New(t)

	mock := &mockRoundTripper{
		responses: []*http.Response{
			newHTTPResponse(200), // immediate success
		},
		errors: []error{nil},
	}

	rt := &retryTransport{
		base:       mock,
		maxRetries: 5,
		log:        l,
		sleep:      sleepNoop(t),
	}

	client := &http.Client{Transport: rt}
	req, _ := http.NewRequest("GET", "http://test", nil)

	resp, err := client.Do(req)
	assert.NoError(err, "expected no error on 200")
	assert.NotNil(resp, "response should not be nil")
	assert.Equal(1, mock.callCount, "should only make one request if successful immediately")
}

func TestRetryOnServerError(t *testing.T) {
	assert := assert.New(t)

	l := slogt.New(t)

	mock := &mockRoundTripper{
		responses: []*http.Response{
			newHTTPResponse(500), // first attempt -> 5xx
			newHTTPResponse(500), // second attempt -> 5xx
			newHTTPResponse(200), // third attempt -> success
		},
		errors: []error{nil, nil, nil},
	}

	rt := &retryTransport{
		base:       mock,
		maxRetries: 5,
		log:        l,
		sleep:      sleepNoop(t),
	}

	client := &http.Client{Transport: rt}
	req, _ := http.NewRequest("GET", "http://test", nil)

	resp, err := client.Do(req)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(3, mock.callCount, "should retry until success (2 failures + 1 success)")
	assert.Equal(200, resp.StatusCode)
}

func TestExhaustRetriesOnServerError(t *testing.T) {
	assert := assert.New(t)

	l := slogt.New(t)

	mock := &mockRoundTripper{
		responses: []*http.Response{
			newHTTPResponse(500),
			newHTTPResponse(500),
			newHTTPResponse(500),
		},
		errors: []error{nil, nil, nil},
	}

	rt := &retryTransport{
		base:       mock,
		maxRetries: 3, // only 3 attempts
		log:        l,
		sleep:      sleepNoop(t),
	}

	client := &http.Client{Transport: rt}
	req, _ := http.NewRequest("GET", "http://test", nil)

	resp, err := client.Do(req)
	// We expect that after 3 attempts (all 500), we still fail
	assert.NotNil(err, "after exhausting retries, err should be the last network or HTTP error")
	// With the current code, if the final attempt yields no error but a 500 response,
	// `err` might be nil, but we treat the request as a failure anyway.
	if resp != nil {
		assert.Equal(500, resp.StatusCode, "last attempt should be 500")
	}
	assert.Equal(3, mock.callCount, "should make exactly 3 attempts")
}

func TestTransportError(t *testing.T) {
	assert := assert.New(t)

	l := slogt.New(t)

	mock := &mockRoundTripper{
		responses: []*http.Response{
			nil,                  // no response first
			newHTTPResponse(200), // then success
		},
		errors: []error{
			fmt.Errorf("simulated network/transport error"),
			nil,
		},
	}

	rt := &retryTransport{
		base:       mock,
		maxRetries: 5,
		log:        l,
		sleep:      sleepNoop(t),
	}

	client := &http.Client{Transport: rt}
	req, _ := http.NewRequest("GET", "http://test", nil)

	resp, err := client.Do(req)
	assert.NoError(err, "should succeed on second attempt")
	assert.NotNil(resp, "response should not be nil")
	assert.Equal(2, mock.callCount, "should only make two attempts (first was error, second is success)")
}
