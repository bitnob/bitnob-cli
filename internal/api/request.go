package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultMaxResponseSize is the default maximum response size (1MB)
	DefaultMaxResponseSize = 1 << 20 // 1MB
)

func (c *Client) Do(ctx context.Context, method string, path string, clientID string, secretKey string, body []byte) ([]byte, error) {
	if clientID == "" || secretKey == "" {
		return nil, fmt.Errorf("client id and secret key are required")
	}
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("path must start with /")
	}

	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		return nil, fmt.Errorf("http method is required")
	}

	fullURL := c.baseURL + path
	payload := string(body)

	var lastErr *Error
	for attempt := 1; attempt <= c.retryMaxAttempts; attempt++ {
		if err := c.checkCircuit(); err != nil {
			return nil, err
		}

		bodyBytes, apiErr := c.doOnce(ctx, method, fullURL, clientID, secretKey, payload, body)
		if apiErr == nil {
			c.recordSuccess()
			return bodyBytes, nil
		}

		lastErr = apiErr
		if isTransientError(apiErr) {
			c.recordTransientFailure()
		} else {
			c.recordSuccess()
			return nil, apiErr
		}

		if attempt == c.retryMaxAttempts {
			return nil, apiErr
		}

		delay := c.retryDelay(attempt, apiErr)
		if err := c.sleep(ctx, delay); err != nil {
			return nil, &Error{
				Type:       ErrorTypeNetwork,
				StatusCode: 0,
				Message:    err.Error(),
			}
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, &Error{
		Type:       ErrorTypeUnknown,
		StatusCode: 0,
		Message:    "request failed",
	}
}

func readResponseBodyLimited(body io.Reader, maxSize int64) ([]byte, error) {
	limited := &io.LimitedReader{R: body, N: maxSize + 1}
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxSize {
		return nil, fmt.Errorf("response body exceeds limit of %d bytes", maxSize)
	}
	return data, nil
}

func (c *Client) doOnce(ctx context.Context, method string, fullURL string, clientID string, secretKey string, payload string, body []byte) ([]byte, *Error) {
	timestamp := strconv.FormatInt(c.now().Unix(), 10)
	nonce, err := c.newNonce()
	if err != nil {
		return nil, &Error{
			Type:       ErrorTypeUnknown,
			StatusCode: 0,
			Message:    err.Error(),
		}
	}

	signature := sign(clientID, timestamp, nonce, payload, secretKey)

	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, &Error{
			Type:       ErrorTypeUnknown,
			StatusCode: 0,
			Message:    err.Error(),
		}
	}

	req.Header.Set("x-auth-client", clientID)
	req.Header.Set("x-auth-timestamp", timestamp)
	req.Header.Set("x-auth-nonce", nonce)
	req.Header.Set("x-auth-signature", signature)
	req.Header.Set("accept", "application/json")
	if len(body) > 0 {
		req.Header.Set("content-type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &Error{
			Type:       ErrorTypeNetwork,
			StatusCode: 0,
			Message:    err.Error(),
		}
	}
	defer resp.Body.Close()

	bodyBytes, err := readResponseBodyLimited(resp.Body, DefaultMaxResponseSize)
	if err != nil {
		return nil, &Error{
			Type:       ErrorTypeNetwork,
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("failed to read response body: %v", err),
		}
	}

	requestID := resp.Header.Get("X-Request-Id")

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return bodyBytes, nil
	}

	apiErr := NewError(resp.StatusCode, bodyBytes)
	apiErr.RequestID = requestID
	if apiErr.Type == ErrorTypeRateLimit {
		apiErr.RetryAfter = parseRetryAfter(resp.Header.Get("Retry-After"), c.now)
	}

	return nil, apiErr
}

func parseRetryAfter(raw string, now func() time.Time) int {
	retryAfter := strings.TrimSpace(raw)
	if retryAfter == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		if seconds > 0 {
			return seconds
		}
		return 0
	}
	if when, err := http.ParseTime(retryAfter); err == nil {
		seconds := int(math.Ceil(when.Sub(now()).Seconds()))
		if seconds > 0 {
			return seconds
		}
	}
	return 0
}

func isTransientError(err *Error) bool {
	if err == nil {
		return false
	}
	if err.Type == ErrorTypeRateLimit || err.Type == ErrorTypeServer {
		return true
	}
	if err.Type == ErrorTypeNetwork {
		return true
	}
	return false
}

func (c *Client) retryDelay(attempt int, err *Error) time.Duration {
	if err != nil && err.Type == ErrorTypeRateLimit && err.RetryAfter > 0 {
		return time.Duration(err.RetryAfter) * time.Second
	}

	delay := c.retryBaseDelay
	for i := 1; i < attempt; i++ {
		delay *= 2
		if delay >= c.retryMaxDelay {
			return c.retryMaxDelay
		}
	}
	if delay > c.retryMaxDelay {
		return c.retryMaxDelay
	}
	return delay
}

func (c *Client) checkCircuit() *Error {
	if c.circuitFailureThreshold <= 0 || c.circuitOpenDuration <= 0 {
		return nil
	}

	c.circuitMu.Lock()
	defer c.circuitMu.Unlock()

	if c.circuitOpenUntil.IsZero() {
		return nil
	}

	now := c.now()
	if now.After(c.circuitOpenUntil) || now.Equal(c.circuitOpenUntil) {
		c.circuitOpenUntil = time.Time{}
		c.circuitConsecutiveFailure = 0
		return nil
	}

	remaining := int(math.Ceil(c.circuitOpenUntil.Sub(now).Seconds()))
	if remaining < 0 {
		remaining = 0
	}

	return &Error{
		Type:       ErrorTypeNetwork,
		StatusCode: 0,
		Message:    fmt.Sprintf("circuit breaker open; retry after %ds", remaining),
		RetryAfter: remaining,
	}
}

func (c *Client) recordSuccess() {
	c.circuitMu.Lock()
	defer c.circuitMu.Unlock()
	c.circuitConsecutiveFailure = 0
	c.circuitOpenUntil = time.Time{}
}

func (c *Client) recordTransientFailure() {
	if c.circuitFailureThreshold <= 0 || c.circuitOpenDuration <= 0 {
		return
	}

	c.circuitMu.Lock()
	defer c.circuitMu.Unlock()
	c.circuitConsecutiveFailure++
	if c.circuitConsecutiveFailure >= c.circuitFailureThreshold {
		c.circuitOpenUntil = c.now().Add(c.circuitOpenDuration)
	}
}
