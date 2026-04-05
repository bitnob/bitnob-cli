package auth

import (
	"context"

	"github.com/bitnob/bitnob-cli/internal/api"
	"github.com/bitnob/bitnob-cli/internal/config"
	"github.com/bitnob/bitnob-cli/internal/profile"
)

type IdentityService struct {
	configStore *config.Store
	authStore   *Store
	client      *api.Client
}

func NewIdentityService(configStore *config.Store, authStore *Store, client *api.Client) *IdentityService {
	return &IdentityService{
		configStore: configStore,
		authStore:   authStore,
		client:      client,
	}
}

func (s *IdentityService) WhoAmI(ctx context.Context) (api.WhoAmIResponse, error) {
	cfg, err := s.configStore.Load(ctx)
	if err != nil {
		return api.WhoAmIResponse{}, err
	}

	active, err := profile.Active(cfg)
	if err != nil {
		return api.WhoAmIResponse{}, err
	}

	credentials, err := s.authStore.LoadCredentials(ctx, active.Name)
	if err != nil {
		return api.WhoAmIResponse{}, CredentialLoadError(active.Name, err)
	}

	return s.client.GetWhoAmI(ctx, credentials.ClientID, credentials.SecretKey)
}
