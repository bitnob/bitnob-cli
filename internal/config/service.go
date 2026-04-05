package config

import (
	"context"
	"fmt"
	"sort"
)

type Summary struct {
	Path          string             `json:"path"`
	ActiveProfile string             `json:"active_profile"`
	Profiles      map[string]Profile `json:"profiles"`
}

type Value struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) Summary(ctx context.Context) (Summary, error) {
	cfg, err := s.store.Load(ctx)
	if err != nil {
		return Summary{}, err
	}

	return Summary{
		Path:          s.store.Path(),
		ActiveProfile: cfg.ActiveProfile,
		Profiles:      cfg.Profiles,
	}, nil
}

func (s *Service) Get(ctx context.Context, key string) (Value, error) {
	cfg, err := s.store.Load(ctx)
	if err != nil {
		return Value{}, err
	}

	value, err := getValue(cfg, key)
	if err != nil {
		return Value{}, err
	}

	return Value{
		Key:   key,
		Value: value,
	}, nil
}

func (s *Service) Set(ctx context.Context, key string, value string) (Value, error) {
	cfg, err := s.store.Load(ctx)
	if err != nil {
		return Value{}, err
	}

	updated, err := setValue(cfg, key, value)
	if err != nil {
		return Value{}, err
	}

	if err := s.store.Save(ctx, updated); err != nil {
		return Value{}, err
	}

	resolved, err := getValue(updated, key)
	if err != nil {
		return Value{}, err
	}

	return Value{
		Key:   key,
		Value: resolved,
	}, nil
}

func (s *Service) Unset(ctx context.Context, key string) (Value, error) {
	cfg, err := s.store.Load(ctx)
	if err != nil {
		return Value{}, err
	}

	updated, err := unsetValue(cfg, key)
	if err != nil {
		return Value{}, err
	}

	if err := s.store.Save(ctx, updated); err != nil {
		return Value{}, err
	}

	resolved, err := getValue(updated, key)
	if err != nil {
		return Value{}, err
	}

	return Value{
		Key:   key,
		Value: resolved,
	}, nil
}

func ProfileNames(cfg Config) []string {
	names := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func getValue(cfg Config, key string) (string, error) {
	switch key {
	case "profile.active":
		return cfg.ActiveProfile, nil
	}

	return "", fmt.Errorf("unsupported config key %q", key)
}

func setValue(cfg Config, key string, value string) (Config, error) {
	switch key {
	case "profile.active":
		if _, ok := cfg.Profiles[value]; !ok {
			return Config{}, fmt.Errorf("profile %q not found", value)
		}
		cfg.ActiveProfile = value
		cfg.ApplyDefaults()
		return cfg, nil
	}

	return Config{}, fmt.Errorf("unsupported config key %q", key)
}

func unsetValue(cfg Config, key string) (Config, error) {
	switch key {
	case "profile.active":
		cfg.ActiveProfile = defaultActiveProfile
		cfg.ApplyDefaults()
		return cfg, nil
	}

	return Config{}, fmt.Errorf("unsupported config key %q", key)
}
