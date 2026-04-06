package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Backend  BackendConfig  `toml:"backend"`
	Defaults DefaultsConfig `toml:"defaults"`
}

type BackendConfig struct {
	Type   string       `toml:"type"`
	Teable TeableConfig `toml:"teable"`
}

type TeableConfig struct {
	URL    string                       `toml:"url"`
	Token  string                       `toml:"token"`
	Tables map[string]string            `toml:"tables"`
	Fields map[string]map[string]string `toml:"fields"`
}

type DefaultsConfig struct {
	Account  string `toml:"account"`
	Currency string `toml:"currency"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = os.Getenv("KOSA_CONFIG")
	}
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot determine home directory: %w", err)
		}
		path = filepath.Join(home, ".config", "kosa", "config.toml")
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("loading config from %s: %w", path, err)
	}

	if cfg.Defaults.Currency == "" {
		cfg.Defaults.Currency = "EUR"
	}

	return &cfg, nil
}
