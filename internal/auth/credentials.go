package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/bitnob/bitnob-cli/internal/api"
	"github.com/bitnob/bitnob-cli/internal/config"
)

type CredentialsService struct {
	configStore *config.Store
	authStore   *Store
	client      *api.Client
}

type LoginResult struct {
	Profile string             `json:"profile"`
	WhoAmI  api.WhoAmIResponse `json:"whoami"`
}

func NewCredentialsService(configStore *config.Store, authStore *Store, client *api.Client) *CredentialsService {
	return &CredentialsService{
		configStore: configStore,
		authStore:   authStore,
		client:      client,
	}
}

func (s *CredentialsService) Login(ctx context.Context, profileName string, clientID string, secretKey string) (LoginResult, error) {
	profileName = strings.TrimSpace(profileName)
	clientID = strings.TrimSpace(clientID)
	secretKey = strings.TrimSpace(secretKey)

	if clientID == "" {
		return LoginResult{}, fmt.Errorf("client id is required")
	}
	if secretKey == "" {
		return LoginResult{}, fmt.Errorf("secret key is required")
	}

	whoami, err := s.client.GetWhoAmI(ctx, clientID, secretKey)
	if err != nil {
		return LoginResult{}, err
	}
	if !whoami.Authenticated {
		return LoginResult{}, fmt.Errorf("authentication failed")
	}

	targetProfile := profileName
	if targetProfile == "" {
		targetProfile = strings.ToLower(strings.TrimSpace(whoami.Environment))
	}
	if targetProfile == "" {
		return LoginResult{}, fmt.Errorf("profile could not be determined from whoami response")
	}

	cfg, err := s.configStore.Load(ctx)
	if err != nil {
		return LoginResult{}, err
	}

	profile := cfg.Profiles[targetProfile]
	profile.AuthMethod = "hmac"
	profile.CredentialsConfigured = true
	cfg.Profiles[targetProfile] = profile
	cfg.ActiveProfile = targetProfile

	credentials := Credentials{
		ClientID:  clientID,
		SecretKey: secretKey,
	}
	if err := s.authStore.SaveCredentials(ctx, targetProfile, credentials); err != nil {
		return LoginResult{}, err
	}

	if err := s.configStore.Save(ctx, cfg); err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		Profile: targetProfile,
		WhoAmI:  whoami,
	}, nil
}

func (s *CredentialsService) Logout(ctx context.Context) error {
	cfg, err := s.configStore.Load(ctx)
	if err != nil {
		return err
	}

	profile := cfg.Profiles[cfg.ActiveProfile]
	profile.AuthMethod = ""
	profile.CredentialsConfigured = false
	cfg.Profiles[cfg.ActiveProfile] = profile

	if err := s.authStore.DeleteCredentials(ctx, cfg.ActiveProfile); err != nil {
		return err
	}

	return s.configStore.Save(ctx, cfg)
}
