package cli

import "github.com/bitnob/bitnob-cli/internal/config"

func applicationConfigWithExtraProfile() config.Config {
	cfg := config.DefaultConfig()
	cfg.Profiles["live"] = config.Profile{
		AuthMethod: "hmac",
	}
	return cfg
}
