package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/bitnob/bitnob-cli/internal/platform"
)

type Store struct {
	secrets *platform.SecretStore
}

func NewStore(secrets *platform.SecretStore) *Store {
	return &Store{secrets: secrets}
}

type Credentials struct {
	ClientID  string `json:"client_id"`
	SecretKey string `json:"secret_key"`
}

func (s *Store) SaveCredentials(ctx context.Context, profile string, credentials Credentials) error {
	data, err := json.Marshal(credentials)
	if err != nil {
		return err
	}

	return s.secrets.Save(ctx, keyName(profile), string(data))
}

func (s *Store) LoadCredentials(ctx context.Context, profile string) (Credentials, error) {
	data, err := s.secrets.Load(ctx, keyName(profile))
	if err != nil {
		return Credentials{}, err
	}

	var credentials Credentials
	if err := json.Unmarshal([]byte(data), &credentials); err != nil {
		return Credentials{}, err
	}

	return credentials, nil
}

func (s *Store) DeleteCredentials(ctx context.Context, profile string) error {
	return s.secrets.Delete(ctx, keyName(profile))
}

func keyName(profile string) string {
	return fmt.Sprintf("%s.credentials", profile)
}

func CredentialLoadError(profile string, err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("no credentials configured for profile %q", profile)
	}
	return fmt.Errorf("load credentials for profile %q: %w", profile, err)
}
