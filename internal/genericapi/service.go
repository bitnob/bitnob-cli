package genericapi

import (
	"context"

	"github.com/bitnob/bitnob-cli/internal/api"
	"github.com/bitnob/bitnob-cli/internal/auth"
	"github.com/bitnob/bitnob-cli/internal/config"
	"github.com/bitnob/bitnob-cli/internal/profile"
)

type Service struct {
	configStore *config.Store
	authStore   *auth.Store
	client      *api.Client
}

func NewService(configStore *config.Store, authStore *auth.Store, client *api.Client) *Service {
	return &Service{
		configStore: configStore,
		authStore:   authStore,
		client:      client,
	}
}

func (s *Service) Do(ctx context.Context, method string, path string, body []byte) ([]byte, error) {
	cfg, err := s.configStore.Load(ctx)
	if err != nil {
		return nil, err
	}

	active, err := profile.Active(cfg)
	if err != nil {
		return nil, err
	}

	credentials, err := s.authStore.LoadCredentials(ctx, active.Name)
	if err != nil {
		return nil, auth.CredentialLoadError(active.Name, err)
	}

	return s.client.Do(ctx, method, path, credentials.ClientID, credentials.SecretKey, body)
}
