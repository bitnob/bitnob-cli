package profile

import (
	"context"
	"fmt"

	"github.com/bitnob/bitnob-cli/internal/config"
)

type Profile struct {
	Name       string
	AuthMethod string
}

func Active(cfg config.Config) (Profile, error) {
	profile, ok := cfg.Profiles[cfg.ActiveProfile]
	if !ok {
		return Profile{}, fmt.Errorf("active profile %q not found", cfg.ActiveProfile)
	}

	return Profile{
		Name:       cfg.ActiveProfile,
		AuthMethod: profile.AuthMethod,
	}, nil
}

type Service struct {
	configStore *config.Store
}

func NewService(configStore *config.Store) *Service {
	return &Service{configStore: configStore}
}

func (s *Service) Switch(ctx context.Context, name string) (Profile, error) {
	cfg, err := s.configStore.Load(ctx)
	if err != nil {
		return Profile{}, err
	}

	if _, ok := cfg.Profiles[name]; !ok {
		return Profile{}, fmt.Errorf("profile %q not found", name)
	}

	cfg.ActiveProfile = name
	if err := s.configStore.Save(ctx, cfg); err != nil {
		return Profile{}, err
	}

	selected := cfg.Profiles[name]

	return Profile{
		Name:       name,
		AuthMethod: selected.AuthMethod,
	}, nil
}
