package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const SignatureHeader = "x-bitnob-signature"

type Service struct {
	httpClient *http.Client
}

type Config struct {
	Secret    string
	ForwardTo string
}

type EventLog struct {
	Type       string `json:"type"`
	Event      string `json:"event,omitempty"`
	ForwardTo  string `json:"forward_to,omitempty"`
	StatusCode int    `json:"status_code,omitempty"`
}

func NewService(client *http.Client) *Service {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &Service{httpClient: client}
}

func Sign(secret string, body []byte) string {
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func Verify(secret string, body []byte, signature string) bool {
	if secret == "" || signature == "" {
		return false
	}
	expected := Sign(secret, body)
	return hmac.Equal([]byte(strings.ToLower(expected)), []byte(strings.ToLower(strings.TrimSpace(signature))))
}

func (s *Service) Handler(cfg Config, logger func(EventLog) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}

		if !Verify(cfg.Secret, body, r.Header.Get(SignatureHeader)) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		var payload struct {
			Event string `json:"event"`
		}
		_ = json.Unmarshal(body, &payload)

		if cfg.ForwardTo == "" {
			if logger != nil {
				_ = logger(EventLog{
					Type:  "webhook.received",
					Event: payload.Event,
				})
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return
		}

		statusCode, err := s.forward(r.Context(), cfg.ForwardTo, body, r.Header)
		if err != nil {
			if logger != nil {
				_ = logger(EventLog{
					Type:       "webhook.forward_failed",
					Event:      payload.Event,
					ForwardTo:  cfg.ForwardTo,
					StatusCode: http.StatusBadGateway,
				})
			}
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		if logger != nil {
			_ = logger(EventLog{
				Type:       "webhook.forwarded",
				Event:      payload.Event,
				ForwardTo:  cfg.ForwardTo,
				StatusCode: statusCode,
			})
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func (s *Service) forward(ctx context.Context, target string, body []byte, headers http.Header) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	for name, values := range headers {
		if strings.EqualFold(name, "host") || strings.EqualFold(name, "content-length") {
			continue
		}
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}
	if req.Header.Get("content-type") == "" {
		req.Header.Set("content-type", "application/json")
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, fmt.Errorf("forward target returned %s", resp.Status)
	}

	return resp.StatusCode, nil
}
