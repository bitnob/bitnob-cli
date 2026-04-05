package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitnob/bitnob-cli/internal/config"
)

func TestRunConfigRestore_RestoresAndBacksUpCurrent(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	configPath := filepath.Join(base, "config.json")
	backupPath := filepath.Join(base, "config.backup.json")

	store := config.NewStore(configPath)
	current := config.DefaultConfig()
	current.ActiveProfile = "live"
	current.Profiles["live"] = config.Profile{AuthMethod: "hmac", CredentialsConfigured: true}
	if err := store.Save(context.Background(), current); err != nil {
		t.Fatalf("save current config: %v", err)
	}

	backup := config.DefaultConfig()
	backup.ActiveProfile = "sandbox"
	backup.Profiles["sandbox"] = config.Profile{AuthMethod: "hmac", CredentialsConfigured: true}
	backupStore := config.NewStore(backupPath)
	if err := backupStore.Save(context.Background(), backup); err != nil {
		t.Fatalf("save backup config: %v", err)
	}

	result, err := runConfigRestore(configPath, backupPath)
	if err != nil {
		t.Fatalf("runConfigRestore returned error: %v", err)
	}

	if result.PreviousBackedUp == "" {
		t.Fatal("expected previous config backup path")
	}
	if _, err := os.Stat(result.PreviousBackedUp); err != nil {
		t.Fatalf("expected previous config backup to exist: %v", err)
	}

	restored, err := store.Load(context.Background())
	if err != nil {
		t.Fatalf("load restored config: %v", err)
	}
	if restored.ActiveProfile != "sandbox" {
		t.Fatalf("unexpected restored active profile: %q", restored.ActiveProfile)
	}
}

func TestRunConfigRestore_InvalidBackupRejected(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	configPath := filepath.Join(base, "config.json")
	backupPath := filepath.Join(base, "config.backup.json")

	store := config.NewStore(configPath)
	current := config.DefaultConfig()
	current.ActiveProfile = "live"
	current.Profiles["live"] = config.Profile{AuthMethod: "hmac", CredentialsConfigured: true}
	if err := store.Save(context.Background(), current); err != nil {
		t.Fatalf("save current config: %v", err)
	}

	if err := os.WriteFile(backupPath, []byte("{not-json"), 0o600); err != nil {
		t.Fatalf("write invalid backup: %v", err)
	}

	_, err := runConfigRestore(configPath, backupPath)
	if err == nil {
		t.Fatal("expected error for invalid backup")
	}
	if !strings.Contains(err.Error(), "valid config JSON") {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, err := store.Load(context.Background())
	if err != nil {
		t.Fatalf("load current config after failed restore: %v", err)
	}
	if loaded.ActiveProfile != "live" {
		t.Fatalf("expected current config to remain unchanged, got %q", loaded.ActiveProfile)
	}
}
