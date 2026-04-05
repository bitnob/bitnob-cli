package config

import (
	"os"
	"path/filepath"
)

func DefaultPath() string {
	if base, err := os.UserConfigDir(); err == nil {
		return filepath.Join(base, "bitnob", "config.json")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "bitnob-config.json"
	}

	return filepath.Join(home, ".bitnob", "config.json")
}
