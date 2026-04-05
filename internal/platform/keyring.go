package platform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/zalando/go-keyring"
)

type SecretStore struct {
	backend secretBackend
}

type secretBackend interface {
	Save(ctx context.Context, key string, value string) error
	Load(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

type keyringBackend struct {
	service string
}

type fileBackend struct {
	baseDir string
}

type failoverBackend struct {
	mu          sync.RWMutex
	useFallback bool
	primary     secretBackend
	fallback    secretBackend
}

func NewSecretStore(service string) *SecretStore {
	if service == "" {
		service = "bitnob-cli"
	}

	return &SecretStore{
		backend: keyringBackend{service: service},
	}
}

func NewFileSecretStore(baseDir string) *SecretStore {
	return &SecretStore{
		backend: fileBackend{baseDir: baseDir},
	}
}

func NewAutoSecretStore(service string, fallbackDir string) (*SecretStore, string, error) {
	if service == "" {
		service = "bitnob-cli"
	}

	primary := keyringBackend{service: service}
	fallback := fileBackend{baseDir: fallbackDir}
	backend := &failoverBackend{
		primary:  primary,
		fallback: fallback,
	}

	probeKey := "__bitnob_cli_probe__"
	probeValue := fmt.Sprintf("probe-%d", time.Now().UTC().UnixNano())
	if err := primary.Save(context.Background(), probeKey, probeValue); err != nil {
		if ensureErr := fallback.Save(context.Background(), probeKey, probeValue); ensureErr != nil {
			return nil, "", fmt.Errorf("keyring unavailable (%v) and file secret store unavailable (%v)", err, ensureErr)
		}
		_ = fallback.Delete(context.Background(), probeKey)
		backend.setUseFallback(true)
		warning := fmt.Sprintf("warning: keyring backend unavailable (%v); using file secret store at %s", err, fallbackDir)
		return &SecretStore{backend: backend}, warning, nil
	}
	_ = primary.Delete(context.Background(), probeKey)

	return &SecretStore{backend: backend}, "", nil
}

func (s *SecretStore) Save(ctx context.Context, key string, value string) error {
	return s.backend.Save(ctx, key, value)
}

func (s *SecretStore) Load(ctx context.Context, key string) (string, error) {
	return s.backend.Load(ctx, key)
}

func (s *SecretStore) Delete(ctx context.Context, key string) error {
	return s.backend.Delete(ctx, key)
}

func (b keyringBackend) Save(_ context.Context, key string, value string) error {
	return keyring.Set(b.service, key, value)
}

func (b keyringBackend) Load(_ context.Context, key string) (string, error) {
	value, err := keyring.Get(b.service, key)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", os.ErrNotExist
	}
	return value, err
}

func (b keyringBackend) Delete(_ context.Context, key string) error {
	err := keyring.Delete(b.service, key)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

func (b *failoverBackend) Save(ctx context.Context, key string, value string) error {
	if b.usingFallback() {
		return b.fallback.Save(ctx, key, value)
	}

	if err := b.primary.Save(ctx, key, value); err != nil {
		if shouldFallback(err) {
			b.setUseFallback(true)
			return b.fallback.Save(ctx, key, value)
		}
		return err
	}

	return nil
}

func (b *failoverBackend) Load(ctx context.Context, key string) (string, error) {
	if b.usingFallback() {
		return b.fallback.Load(ctx, key)
	}

	value, err := b.primary.Load(ctx, key)
	if err == nil {
		return value, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		fallbackValue, fallbackErr := b.fallback.Load(ctx, key)
		if fallbackErr == nil {
			b.setUseFallback(true)
			return fallbackValue, nil
		}
		if errors.Is(fallbackErr, os.ErrNotExist) {
			return "", err
		}
		return "", fallbackErr
	}

	if shouldFallback(err) {
		b.setUseFallback(true)
		return b.fallback.Load(ctx, key)
	}

	return "", err
}

func (b *failoverBackend) Delete(ctx context.Context, key string) error {
	if b.usingFallback() {
		return b.fallback.Delete(ctx, key)
	}

	err := b.primary.Delete(ctx, key)
	if err == nil {
		_ = b.fallback.Delete(ctx, key)
		return nil
	}

	if shouldFallback(err) {
		b.setUseFallback(true)
		return b.fallback.Delete(ctx, key)
	}

	return err
}

func (b *failoverBackend) usingFallback() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.useFallback
}

func (b *failoverBackend) setUseFallback(v bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.useFallback = v
}

func shouldFallback(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, keyring.ErrNotFound) {
		return false
	}
	if errors.Is(err, keyring.ErrUnsupportedPlatform) {
		return true
	}

	var execErr *exec.Error
	if errors.As(err, &execErr) {
		return true
	}

	msg := strings.ToLower(err.Error())
	signals := []string{
		"keyring",
		"keychain",
		"secret service",
		"dbus",
		"unsupported platform",
		"not available",
		"autolaunch",
	}
	for _, token := range signals {
		if strings.Contains(msg, token) {
			return true
		}
	}

	return false
}

func (b fileBackend) Save(_ context.Context, key string, value string) error {
	if err := os.MkdirAll(b.baseDir, 0o700); err != nil {
		return err
	}

	path := filepath.Join(b.baseDir, key)
	return os.WriteFile(path, []byte(value), 0o600)
}

func (b fileBackend) Load(_ context.Context, key string) (string, error) {
	path := filepath.Join(b.baseDir, key)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (b fileBackend) Delete(_ context.Context, key string) error {
	path := filepath.Join(b.baseDir, key)
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
