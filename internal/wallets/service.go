package wallets

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

func (s *Service) List(ctx context.Context) (api.BalancesResponse, error) {
	cfg, err := s.configStore.Load(ctx)
	if err != nil {
		return api.BalancesResponse{}, err
	}

	active, err := profile.Active(cfg)
	if err != nil {
		return api.BalancesResponse{}, err
	}

	credentials, err := s.authStore.LoadCredentials(ctx, active.Name)
	if err != nil {
		return api.BalancesResponse{}, auth.CredentialLoadError(active.Name, err)
	}

	return s.client.GetWallets(ctx, credentials.ClientID, credentials.SecretKey)
}
