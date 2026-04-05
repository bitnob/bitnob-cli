package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const defaultActiveProfile = "default"

type Config struct {
	ActiveProfile string             `json:"active_profile"`
	Profiles      map[string]Profile `json:"profiles"`
}

type Profile struct {
	AuthMethod            string `json:"auth_method,omitempty"`
	CredentialsConfigured bool   `json:"credentials_configured,omitempty"`
}

type Store struct {
	path string
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Path() string {
	return s.path
}

func (s *Store) Load(_ context.Context) (Config, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	cfg.ApplyDefaults()
	return cfg, nil
}

func (s *Store) Save(_ context.Context, cfg Config) error {
	cfg.ApplyDefaults()

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')
	return os.WriteFile(s.path, data, 0o600)
}

// LoadOrRecover loads config and recovers from malformed JSON by backing it up
// and writing a fresh default config.
func (s *Store) LoadOrRecover(ctx context.Context) (Config, string, error) {
	cfg, err := s.Load(ctx)
	if err == nil {
		return cfg, "", nil
	}

	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	if !errors.As(err, &syntaxErr) && !errors.As(err, &typeErr) {
		return Config{}, "", err
	}

	backupPath := fmt.Sprintf("%s.broken.%s", s.path, time.Now().UTC().Format("20060102T150405Z"))
	if renameErr := os.Rename(s.path, backupPath); renameErr != nil {
		return Config{}, "", fmt.Errorf("backup malformed config: %w (original error: %v)", renameErr, err)
	}

	cfg = DefaultConfig()
	if saveErr := s.Save(ctx, cfg); saveErr != nil {
		return Config{}, "", fmt.Errorf("write recovered default config: %w (backup at %s)", saveErr, backupPath)
	}

	warning := fmt.Sprintf("warning: malformed config at %s was backed up to %s and reset to defaults", s.path, backupPath)
	return cfg, warning, nil
}

func DefaultConfig() Config {
	cfg := Config{
		ActiveProfile: defaultActiveProfile,
		Profiles: map[string]Profile{
			defaultActiveProfile: {},
		},
	}

	cfg.ApplyDefaults()
	return cfg
}

func (c *Config) ApplyDefaults() {
	if c.ActiveProfile == "" {
		c.ActiveProfile = defaultActiveProfile
	}

	if c.Profiles == nil {
		c.Profiles = make(map[string]Profile)
	}

	if _, ok := c.Profiles[c.ActiveProfile]; !ok {
		c.Profiles[c.ActiveProfile] = Profile{}
	}
}
