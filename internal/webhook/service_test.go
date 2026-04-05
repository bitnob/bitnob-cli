package webhook

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVerify(t *testing.T) {
	body := []byte(`{"event":"deposit.success"}`)
	secret := "webhook_secret_test"
	signature := Sign(secret, body)

	if !Verify(secret, body, signature) {
		t.Fatal("expected signature to verify")
	}
	if Verify(secret, body, "bad-signature") {
		t.Fatal("expected invalid signature to fail verification")
	}
}

func TestHandlerVerifiesAndForwards(t *testing.T) {
	forwarded := make(chan []byte, 1)
	service := NewService(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(req.Body)
			forwarded <- body
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(bytes.NewReader(nil)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}),
	})
	var logs []EventLog
	handler := service.Handler(Config{
		Secret:    "webhook_secret_test",
		ForwardTo: "http://forward.test/webhook",
	}, func(entry EventLog) error {
		logs = append(logs, entry)
		return nil
	})

	body := []byte(`{"event":"deposit.success","data":{"amount":"1000"}}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(SignatureHeader, Sign("webhook_secret_test", body))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}
	select {
	case got := <-forwarded:
		if string(got) != string(body) {
			t.Fatalf("unexpected forwarded body: %s", got)
		}
	default:
		t.Fatal("expected event to be forwarded")
	}
	if len(logs) != 1 || logs[0].Type != "webhook.forwarded" {
		t.Fatalf("unexpected logs: %+v", logs)
	}
}

func TestHandlerRejectsInvalidSignature(t *testing.T) {
	service := NewService(nil)
	handler := service.Handler(Config{Secret: "webhook_secret_test"}, nil)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte(`{"event":"deposit.success"}`)))
	req.Header.Set(SignatureHeader, "bad-signature")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
