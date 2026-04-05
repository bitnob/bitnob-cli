package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestClientDo_RetriesTransientErrorAndSucceeds(t *testing.T) {
	t.Parallel()

	var calls int32
	httpClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			n := atomic.AddInt32(&calls, 1)
			if n == 1 {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Status:     "500 Internal Server Error",
					Body:       io.NopCloser(strings.NewReader(`{"error":"temporary failure"}`)),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}),
	}

	client := NewClientWithOptions(Options{
		BaseURL:                 "https://api.test",
		HTTPClient:              httpClient,
		RetryMaxAttempts:        3,
		RetryBaseDelay:          time.Millisecond,
		RetryMaxDelay:           2 * time.Millisecond,
		CircuitFailureThreshold: 10,
		CircuitOpenDuration:     time.Second,
		Sleep:                   func(_ context.Context, _ time.Duration) error { return nil },
	})

	body, err := client.Do(context.Background(), "GET", "/api/test", "client", "secret", nil)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected response body: %s", string(body))
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("unexpected call count: got=%d want=2", got)
	}
}

func TestClientDo_CircuitBreakerOpensAfterFailures(t *testing.T) {
	t.Parallel()

	var calls int32
	httpClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			atomic.AddInt32(&calls, 1)
			return nil, errors.New("dial tcp: connection refused")
		}),
	}

	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	client := NewClientWithOptions(Options{
		BaseURL:                 "https://api.test",
		HTTPClient:              httpClient,
		RetryMaxAttempts:        1,
		CircuitFailureThreshold: 2,
		CircuitOpenDuration:     time.Minute,
		Now:                     func() time.Time { return now },
		Sleep:                   func(_ context.Context, _ time.Duration) error { return nil },
	})

	_, _ = client.Do(context.Background(), "GET", "/api/test", "client", "secret", nil)
	_, _ = client.Do(context.Background(), "GET", "/api/test", "client", "secret", nil)

	_, err := client.Do(context.Background(), "GET", "/api/test", "client", "secret", nil)
	if err == nil {
		t.Fatal("expected circuit breaker error")
	}

	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.Type != ErrorTypeNetwork {
		t.Fatalf("unexpected error type: %s", apiErr.Type)
	}
	if !strings.Contains(apiErr.Message, "circuit breaker open") {
		t.Fatalf("unexpected error message: %s", apiErr.Message)
	}

	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("unexpected call count: got=%d want=2", got)
	}
}

func TestParseRetryAfterHTTPDate(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	header := now.Add(3 * time.Second).Format(http.TimeFormat)
	seconds := parseRetryAfter(header, func() time.Time { return now })
	if seconds != 3 {
		t.Fatalf("unexpected Retry-After parse: got=%d want=3", seconds)
	}
}
