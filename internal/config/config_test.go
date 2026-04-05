package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadOrRecover_BacksUpMalformedConfig(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	path := filepath.Join(base, "config.json")
	malformed := []byte("{invalid-json")
	if err := os.WriteFile(path, malformed, 0o600); err != nil {
		t.Fatalf("write malformed config: %v", err)
	}

	store := NewStore(path)
	cfg, warning, err := store.LoadOrRecover(context.Background())
	if err != nil {
		t.Fatalf("LoadOrRecover returned error: %v", err)
	}

	if warning == "" {
		t.Fatal("expected warning for malformed config recovery")
	}
	if !strings.Contains(warning, "reset to defaults") {
		t.Fatalf("unexpected warning: %q", warning)
	}
	if cfg.ActiveProfile != "default" {
		t.Fatalf("unexpected active profile: %q", cfg.ActiveProfile)
	}

	currentData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read recovered config: %v", err)
	}

	var recovered Config
	if err := json.Unmarshal(currentData, &recovered); err != nil {
		t.Fatalf("recovered config is not valid JSON: %v", err)
	}
	if recovered.ActiveProfile != "default" {
		t.Fatalf("unexpected recovered active profile: %q", recovered.ActiveProfile)
	}

	backups, err := filepath.Glob(path + ".broken.*")
	if err != nil {
		t.Fatalf("glob backup files: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("expected exactly one backup file, got %d", len(backups))
	}

	backupData, err := os.ReadFile(backups[0])
	if err != nil {
		t.Fatalf("read backup config: %v", err)
	}
	if string(backupData) != string(malformed) {
		t.Fatalf("backup file content mismatch: got %q", string(backupData))
	}
}

func TestLoadOrRecover_ValidConfig_NoWarning(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	path := filepath.Join(base, "config.json")
	cfg := DefaultConfig()
	cfg.ActiveProfile = "live"
	cfg.Profiles["live"] = Profile{AuthMethod: "hmac", CredentialsConfigured: true}

	store := NewStore(path)
	if err := store.Save(context.Background(), cfg); err != nil {
		t.Fatalf("save valid config: %v", err)
	}

	loaded, warning, err := store.LoadOrRecover(context.Background())
	if err != nil {
		t.Fatalf("LoadOrRecover returned error: %v", err)
	}
	if warning != "" {
		t.Fatalf("did not expect warning, got %q", warning)
	}
	if loaded.ActiveProfile != "live" {
		t.Fatalf("unexpected active profile: %q", loaded.ActiveProfile)
	}
}
