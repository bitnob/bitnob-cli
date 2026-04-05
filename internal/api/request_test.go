package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClient_Do(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		clientID       string
		secretKey      string
		body           []byte
		serverResponse func(w http.ResponseWriter, r *http.Request)
		wantErr        bool
		checkError     func(t *testing.T, err error)
	}{
		{
			name:      "successful GET request",
			method:    "GET",
			path:      "/api/test",
			clientID:  "test-client",
			secretKey: "test-secret",
			body:      nil,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"result": "success"}`))
			},
			wantErr: false,
		},
		{
			name:      "successful POST request with body",
			method:    "POST",
			path:      "/api/test",
			clientID:  "test-client",
			secretKey: "test-secret",
			body:      []byte(`{"data": "test"}`),
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				// Verify request body
				body, _ := io.ReadAll(r.Body)
				if string(body) != `{"data": "test"}` {
					t.Errorf("unexpected request body: %s", body)
				}
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"id": "123"}`))
			},
			wantErr: false,
		},
		{
			name:      "authentication error",
			method:    "GET",
			path:      "/api/test",
			clientID:  "test-client",
			secretKey: "test-secret",
			body:      nil,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Invalid credentials"}`))
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				apiErr, ok := err.(*Error)
				if !ok {
					t.Errorf("expected *Error, got %T", err)
					return
				}
				if apiErr.Type != ErrorTypeAuth {
					t.Errorf("expected ErrorTypeAuth, got %v", apiErr.Type)
				}
				if apiErr.StatusCode != 401 {
					t.Errorf("expected status 401, got %d", apiErr.StatusCode)
				}
			},
		},
		{
			name:      "rate limit error with retry-after",
			method:    "GET",
			path:      "/api/test",
			clientID:  "test-client",
			secretKey: "test-secret",
			body:      nil,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error": "Rate limit exceeded"}`))
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				apiErr, ok := err.(*Error)
				if !ok {
					t.Errorf("expected *Error, got %T", err)
					return
				}
				if apiErr.Type != ErrorTypeRateLimit {
					t.Errorf("expected ErrorTypeRateLimit, got %v", apiErr.Type)
				}
				if apiErr.RetryAfter != 60 {
					t.Errorf("expected RetryAfter 60, got %d", apiErr.RetryAfter)
				}
			},
		},
		{
			name:      "not found error",
			method:    "GET",
			path:      "/api/test",
			clientID:  "test-client",
			secretKey: "test-secret",
			body:      nil,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error": "Resource not found"}`))
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				if !IsNotFoundError(err) {
					t.Errorf("expected IsNotFoundError to return true")
				}
			},
		},
		{
			name:      "server error",
			method:    "GET",
			path:      "/api/test",
			clientID:  "test-client",
			secretKey: "test-secret",
			body:      nil,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Internal server error"}`))
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				apiErr, ok := err.(*Error)
				if !ok {
					t.Errorf("expected *Error, got %T", err)
					return
				}
				if !apiErr.IsRetryable() {
					t.Errorf("expected server error to be retryable")
				}
			},
		},
		{
			name:      "empty client ID",
			method:    "GET",
			path:      "/api/test",
			clientID:  "",
			secretKey: "test-secret",
			body:      nil,
			wantErr:   true,
		},
		{
			name:      "empty secret key",
			method:    "GET",
			path:      "/api/test",
			clientID:  "test-client",
			secretKey: "",
			body:      nil,
			wantErr:   true,
		},
		{
			name:      "invalid path without slash",
			method:    "GET",
			path:      "api/test",
			clientID:  "test-client",
			secretKey: "test-secret",
			body:      nil,
			wantErr:   true,
		},
		{
			name:      "empty method",
			method:    "",
			path:      "/api/test",
			clientID:  "test-client",
			secretKey: "test-secret",
			body:      nil,
			wantErr:   true,
		},
		{
			name:      "request with X-Request-Id header",
			method:    "GET",
			path:      "/api/test",
			clientID:  "test-client",
			secretKey: "test-secret",
			body:      nil,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Request-Id", "req-123")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "Bad request"}`))
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				apiErr, ok := err.(*Error)
				if !ok {
					t.Errorf("expected *Error, got %T", err)
					return
				}
				if apiErr.RequestID != "req-123" {
					t.Errorf("expected RequestID req-123, got %s", apiErr.RequestID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify headers
				if tt.clientID != "" {
					if r.Header.Get("x-auth-client") != tt.clientID {
						t.Errorf("expected x-auth-client %s, got %s", tt.clientID, r.Header.Get("x-auth-client"))
					}
					if r.Header.Get("x-auth-timestamp") == "" {
						t.Error("missing x-auth-timestamp header")
					}
					if r.Header.Get("x-auth-nonce") == "" {
						t.Error("missing x-auth-nonce header")
					}
					if r.Header.Get("x-auth-signature") == "" {
						t.Error("missing x-auth-signature header")
					}
				}

				if tt.serverResponse != nil {
					tt.serverResponse(w, r)
				}
			}))
			defer server.Close()

			// Create client with test server URL
			client := NewClientWithOptions(Options{
				BaseURL: server.URL,
			})
			ctx := context.Background()

			// Execute request
			result, err := client.Do(ctx, tt.method, tt.path, tt.clientID, tt.secretKey, tt.body)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check specific error conditions
			if tt.checkError != nil && err != nil {
				tt.checkError(t, err)
			}

			// Check result for successful requests
			if !tt.wantErr && result == nil {
				t.Error("expected non-nil result for successful request")
			}
		})
	}
}

func TestClient_Signature(t *testing.T) {
	client := NewClient()

	// Override time for consistent testing
	client.now = func() time.Time {
		return time.Unix(1234567890, 0)
	}

	// Override nonce for consistent testing
	client.newNonce = func() (string, error) {
		return "test-nonce", nil
	}

	clientID := "test-client"
	secretKey := "test-secret"
	payload := `{"test": "data"}`
	timestamp := "1234567890"
	nonce := "test-nonce"

	signature := sign(clientID, timestamp, nonce, payload, secretKey)

	// Verify signature format (should be hex)
	if len(signature) != 64 { // SHA256 produces 32 bytes = 64 hex chars
		t.Errorf("unexpected signature length: %d", len(signature))
	}

	// Verify signature is consistent
	signature2 := sign(clientID, timestamp, nonce, payload, secretKey)
	if signature != signature2 {
		t.Error("signature not consistent")
	}

	// Verify signature changes with different inputs
	signature3 := sign(clientID, timestamp, nonce, "different payload", secretKey)
	if signature == signature3 {
		t.Error("signature should change with different payload")
	}
}

func TestError_IsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want bool
	}{
		{
			name: "rate limit error is retryable",
			err: &Error{
				Type:       ErrorTypeRateLimit,
				StatusCode: 429,
			},
			want: true,
		},
		{
			name: "server error is retryable",
			err: &Error{
				Type:       ErrorTypeServer,
				StatusCode: 500,
			},
			want: true,
		},
		{
			name: "network error with 500+ status is retryable",
			err: &Error{
				Type:       ErrorTypeNetwork,
				StatusCode: 502,
			},
			want: true,
		},
		{
			name: "network error with 0 status is retryable",
			err: &Error{
				Type:       ErrorTypeNetwork,
				StatusCode: 0,
			},
			want: true,
		},
		{
			name: "auth error is not retryable",
			err: &Error{
				Type:       ErrorTypeAuth,
				StatusCode: 401,
			},
			want: false,
		},
		{
			name: "not found error is not retryable",
			err: &Error{
				Type:       ErrorTypeNotFound,
				StatusCode: 404,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.IsRetryable(); got != tt.want {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       []byte
		wantType   ErrorType
		wantMsg    string
	}{
		{
			name:       "JSON error response",
			statusCode: 400,
			body:       []byte(`{"error": "Invalid input"}`),
			wantType:   ErrorTypeValidation,
			wantMsg:    "Invalid input",
		},
		{
			name:       "JSON message response",
			statusCode: 403,
			body:       []byte(`{"message": "Permission denied"}`),
			wantType:   ErrorTypePermission,
			wantMsg:    "Permission denied",
		},
		{
			name:       "Plain text response",
			statusCode: 500,
			body:       []byte("Internal server error"),
			wantType:   ErrorTypeServer,
			wantMsg:    "Internal server error",
		},
		{
			name:       "Empty response",
			statusCode: 503,
			body:       []byte(""),
			wantType:   ErrorTypeServer,
			wantMsg:    "Service Unavailable",
		},
		{
			name:       "JSON with details",
			statusCode: 422,
			body:       []byte(`{"message": "Validation failed", "details": "Field 'email' is required"}`),
			wantType:   ErrorTypeValidation,
			wantMsg:    "Validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewError(tt.statusCode, tt.body)

			if err.Type != tt.wantType {
				t.Errorf("NewError() Type = %v, want %v", err.Type, tt.wantType)
			}

			if err.StatusCode != tt.statusCode {
				t.Errorf("NewError() StatusCode = %v, want %v", err.StatusCode, tt.statusCode)
			}

			if err.Message != tt.wantMsg {
				t.Errorf("NewError() Message = %v, want %v", err.Message, tt.wantMsg)
			}
		})
	}
}

func TestClient_ConcurrentRequests(t *testing.T) {
	// Create test server
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": "success"}`))
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClientWithOptions(Options{
		BaseURL: server.URL,
	})
	ctx := context.Background()

	// Run concurrent requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := client.Do(ctx, "GET", "/api/test", "client", "secret", nil)
			if err != nil {
				t.Errorf("concurrent request failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all requests
	for i := 0; i < 10; i++ {
		<-done
	}

	if requestCount != 10 {
		t.Errorf("expected 10 requests, got %d", requestCount)
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClientWithOptions(Options{
		BaseURL: server.URL,
	})

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Execute request
	_, err := client.Do(ctx, "GET", "/api/test", "client", "secret", nil)

	// Should get a context error
	if err == nil {
		t.Error("expected context cancellation error")
	}

	// Check if it's a network error (context cancelled)
	if apiErr, ok := err.(*Error); ok {
		if apiErr.Type != ErrorTypeNetwork {
			t.Errorf("expected ErrorTypeNetwork for context cancellation, got %v", apiErr.Type)
		}
		if !strings.Contains(apiErr.Message, "context") {
			t.Errorf("expected error message to mention context, got %v", apiErr.Message)
		}
	}
}
