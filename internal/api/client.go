package api

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

const (
	defaultBaseURL = "https://api.bitnob.com"
)

type Client struct {
	httpClient *http.Client
	HTTPClient *http.Client // Exported for access
	baseURL    string
	now        func() time.Time
	newNonce   func() (string, error)
	sleep      func(context.Context, time.Duration) error

	retryMaxAttempts        int
	retryBaseDelay          time.Duration
	retryMaxDelay           time.Duration
	circuitFailureThreshold int
	circuitOpenDuration     time.Duration

	circuitMu                 sync.Mutex
	circuitConsecutiveFailure int
	circuitOpenUntil          time.Time
}

type Options struct {
	HTTPClient *http.Client
	BaseURL    string
	Now        func() time.Time
	NewNonce   func() (string, error)
	Sleep      func(context.Context, time.Duration) error

	RetryMaxAttempts        int
	RetryBaseDelay          time.Duration
	RetryMaxDelay           time.Duration
	CircuitFailureThreshold int
	CircuitOpenDuration     time.Duration
}

func NewClient() *Client {
	return NewClientWithOptions(Options{})
}

func NewClientWithOptions(opts Options) *Client {
	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 15 * time.Second,
		}
	}

	now := opts.Now
	if now == nil {
		now = time.Now().UTC
	}

	newNonce := opts.NewNonce
	if newNonce == nil {
		newNonce = generateNonce
	}

	sleep := opts.Sleep
	if sleep == nil {
		sleep = sleepWithContext
	}

	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	retryMaxAttempts := opts.RetryMaxAttempts
	if retryMaxAttempts <= 0 {
		retryMaxAttempts = 3
	}

	retryBaseDelay := opts.RetryBaseDelay
	if retryBaseDelay <= 0 {
		retryBaseDelay = 250 * time.Millisecond
	}

	retryMaxDelay := opts.RetryMaxDelay
	if retryMaxDelay <= 0 {
		retryMaxDelay = 2 * time.Second
	}
	if retryMaxDelay < retryBaseDelay {
		retryMaxDelay = retryBaseDelay
	}

	circuitFailureThreshold := opts.CircuitFailureThreshold
	if circuitFailureThreshold <= 0 {
		circuitFailureThreshold = 5
	}

	circuitOpenDuration := opts.CircuitOpenDuration
	if circuitOpenDuration <= 0 {
		circuitOpenDuration = 30 * time.Second
	}

	return &Client{
		httpClient: httpClient,
		HTTPClient: httpClient,
		baseURL:    baseURL,
		now:        now,
		newNonce:   newNonce,
		sleep:      sleep,

		retryMaxAttempts:        retryMaxAttempts,
		retryBaseDelay:          retryBaseDelay,
		retryMaxDelay:           retryMaxDelay,
		circuitFailureThreshold: circuitFailureThreshold,
		circuitOpenDuration:     circuitOpenDuration,
	}
}

func sign(clientID string, timestamp string, nonce string, body string, secretKey string) string {
	message := clientID + ":" + timestamp + ":" + nonce + ":" + body
	mac := hmac.New(sha256.New, []byte(secretKey))
	_, _ = mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func generateNonce() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw[:]), nil
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
