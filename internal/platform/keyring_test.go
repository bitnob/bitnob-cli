package platform

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/zalando/go-keyring"
)

type stubBackend struct {
	store     map[string]string
	saveErr   error
	loadErr   error
	deleteErr error
}

func (s *stubBackend) Save(_ context.Context, key string, value string) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	if s.store == nil {
		s.store = make(map[string]string)
	}
	s.store[key] = value
	return nil
}

func (s *stubBackend) Load(_ context.Context, key string) (string, error) {
	if s.loadErr != nil {
		return "", s.loadErr
	}
	value, ok := s.store[key]
	if !ok {
		return "", os.ErrNotExist
	}
	return value, nil
}

func (s *stubBackend) Delete(_ context.Context, key string) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	delete(s.store, key)
	return nil
}

func TestFailoverBackend_SaveFallsBackOnUnavailablePrimary(t *testing.T) {
	t.Parallel()

	primary := &stubBackend{saveErr: keyring.ErrUnsupportedPlatform}
	fallback := &stubBackend{store: map[string]string{}}
	backend := &failoverBackend{primary: primary, fallback: fallback}

	if err := backend.Save(context.Background(), "profile.credentials", "secret"); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if !backend.usingFallback() {
		t.Fatal("expected fallback mode to be enabled")
	}
	if got := fallback.store["profile.credentials"]; got != "secret" {
		t.Fatalf("unexpected fallback stored value: %q", got)
	}
}

func TestFailoverBackend_LoadUsesFallbackWhenPrimaryMissing(t *testing.T) {
	t.Parallel()

	primary := &stubBackend{store: map[string]string{}}
	fallback := &stubBackend{store: map[string]string{"profile.credentials": "from-file"}}
	backend := &failoverBackend{primary: primary, fallback: fallback}

	got, err := backend.Load(context.Background(), "profile.credentials")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if got != "from-file" {
		t.Fatalf("unexpected loaded value: %q", got)
	}
	if !backend.usingFallback() {
		t.Fatal("expected fallback mode to be enabled after fallback load")
	}
}

func TestShouldFallback(t *testing.T) {
	t.Parallel()

	if shouldFallback(nil) {
		t.Fatal("nil error should not trigger fallback")
	}
	if shouldFallback(os.ErrNotExist) {
		t.Fatal("os.ErrNotExist should not trigger fallback")
	}
	if !shouldFallback(keyring.ErrUnsupportedPlatform) {
		t.Fatal("ErrUnsupportedPlatform should trigger fallback")
	}
	if !shouldFallback(errors.New("cannot autolaunch D-Bus without X11 $DISPLAY")) {
		t.Fatal("dbus startup error should trigger fallback")
	}
}
