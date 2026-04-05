package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorType represents the type of API error
type ErrorType string

const (
	ErrorTypeAuth       ErrorType = "authentication_error"
	ErrorTypePermission ErrorType = "permission_error"
	ErrorTypeNotFound   ErrorType = "not_found"
	ErrorTypeValidation ErrorType = "validation_error"
	ErrorTypeRateLimit  ErrorType = "rate_limit_error"
	ErrorTypeServer     ErrorType = "server_error"
	ErrorTypeNetwork    ErrorType = "network_error"
	ErrorTypeUnknown    ErrorType = "unknown_error"
)

// Error represents a structured API error
type Error struct {
	Type       ErrorType `json:"type"`
	StatusCode int       `json:"status_code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	RequestID  string    `json:"request_id,omitempty"`
	RetryAfter int       `json:"retry_after,omitempty"` // For rate limit errors
}

func (e *Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s (status %d): %s - %s", e.Type, e.StatusCode, e.Message, e.Details)
	}
	return fmt.Sprintf("%s (status %d): %s", e.Type, e.StatusCode, e.Message)
}

// IsRetryable returns true if the error is temporary and can be retried
func (e *Error) IsRetryable() bool {
	switch e.Type {
	case ErrorTypeRateLimit, ErrorTypeServer:
		return true
	case ErrorTypeNetwork:
		return e.StatusCode >= 500 || e.StatusCode == 0
	default:
		return false
	}
}

// NewError creates a new API error from an HTTP response
func NewError(statusCode int, body []byte) *Error {
	err := &Error{
		StatusCode: statusCode,
		Type:       getErrorType(statusCode),
	}

	// Try to parse the response body as JSON
	var apiErr struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Details string `json:"details"`
	}

	if json.Unmarshal(body, &apiErr) == nil {
		if apiErr.Error != "" {
			err.Message = apiErr.Error
		} else if apiErr.Message != "" {
			err.Message = apiErr.Message
		}
		err.Details = apiErr.Details
	} else {
		// If not JSON, use the raw body as the message
		err.Message = string(body)
		if err.Message == "" {
			err.Message = http.StatusText(statusCode)
		}
	}

	return err
}

// getErrorType determines the error type based on status code
func getErrorType(statusCode int) ErrorType {
	switch {
	case statusCode == 401:
		return ErrorTypeAuth
	case statusCode == 403:
		return ErrorTypePermission
	case statusCode == 404:
		return ErrorTypeNotFound
	case statusCode == 422 || statusCode == 400:
		return ErrorTypeValidation
	case statusCode == 429:
		return ErrorTypeRateLimit
	case statusCode >= 500:
		return ErrorTypeServer
	case statusCode == 0:
		return ErrorTypeNetwork
	default:
		return ErrorTypeUnknown
	}
}

// IsAuthError checks if the error is an authentication error
func IsAuthError(err error) bool {
	apiErr, ok := err.(*Error)
	return ok && apiErr.Type == ErrorTypeAuth
}

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	apiErr, ok := err.(*Error)
	return ok && apiErr.Type == ErrorTypeNotFound
}

// IsRateLimitError checks if the error is a rate limit error
func IsRateLimitError(err error) bool {
	apiErr, ok := err.(*Error)
	return ok && apiErr.Type == ErrorTypeRateLimit
}
