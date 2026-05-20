package auth

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitnob/bitnob-cli/internal/platform"
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

func TestValidateProfileName(t *testing.T) {
	t.Parallel()

	valid := []string{"default", "sandbox", "live-1", "team.alpha", "team_alpha"}
	for _, profile := range valid {
		if err := ValidateProfileName(profile); err != nil {
			t.Fatalf("expected profile %q to be valid: %v", profile, err)
		}
	}

	invalid := []string{"", "../escaped", "team/alpha", "team alpha", "team:alpha"}
	for _, profile := range invalid {
		if err := ValidateProfileName(profile); err == nil {
			t.Fatalf("expected profile %q to be invalid", profile)
		}
	}
}

func TestStoreRejectsUnsafeProfileName(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	store := NewStore(platform.NewFileSecretStore(filepath.Join(base, "secrets")))
	err := store.SaveCredentials(context.Background(), "../escaped", Credentials{
		ClientID:  "client",
		SecretKey: "secret",
	})
	if err == nil {
		t.Fatal("expected unsafe profile name to be rejected")
	}

	if _, statErr := os.Stat(filepath.Join(base, "escaped.credentials")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("unsafe profile wrote outside secret directory: %v", statErr)
	}
}
