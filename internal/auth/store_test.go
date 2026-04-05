package auth

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestCredentialLoadError_NotConfigured(t *testing.T) {
	t.Parallel()

	err := CredentialLoadError("live", os.ErrNotExist)
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); !strings.Contains(got, `no credentials configured for profile "live"`) {
		t.Fatalf("unexpected message: %q", got)
	}
}

func TestCredentialLoadError_BackendFailure(t *testing.T) {
	t.Parallel()

	root := errors.New("permission denied")
	err := CredentialLoadError("live", root)
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); !strings.Contains(got, `load credentials for profile "live": permission denied`) {
		t.Fatalf("unexpected message: %q", got)
	}
}
