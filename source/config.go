package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	DEFAULT_CONFIG_PATH = "/etc/tapo.toml"

	DEFAULT_METRICS_LISTEN   = "127.0.0.1:5000"
	DEFAULT_METRICS_PATH     = "/metrics"
	DEFAULT_METRICS_INTERVAL = 10 * time.Second
)

/************************************************************/

type Config struct {
	Metrics MetricsConfig `toml:"metrics"`
	Account AccountConfig `toml:"account"`
	Plugs   []PlugConfig  `toml:"plugs"`
}

type MetricsConfig struct {
	Listen   string        `toml:"listen"`
	Path     string        `toml:"path"`
	Interval time.Duration `toml:"interval"`
}

type AccountConfig struct {
	Email    string `toml:"email"`
	Password string `toml:"password"`
}

type PlugConfig struct {
	Enabled *bool `toml:"enabled"`

	Name      string `toml:"name"`
	IPAddress string `toml:"ip_address"`

	Email    *string `toml:"email"`
	Password *string `toml:"password"`
}

/************************************************************/

func (plug *PlugConfig) IsEnabled() bool {
	if plug.Enabled == nil {
		return true
	}

	return *plug.Enabled
}

func (plug *PlugConfig) DetermineEmail(global AccountConfig) string {
	if plug.Email != nil {
		return *plug.Email
	}

	return global.Email
}

func (plug *PlugConfig) DeterminePassword(account AccountConfig) string {
	if plug.Password != nil {
		return *plug.Password
	}

	return account.Password
}

func updateConfigWithEnvironmentVariables(cfg *Config) {
	if value := os.Getenv("TAPO_METRICS_LISTEN"); value != "" {
		cfg.Metrics.Listen = value
	}

	if value := os.Getenv("TAPO_METRICS_PATH"); value != "" {
		cfg.Metrics.Path = value
	}

	if value := os.Getenv("TAPO_METRICS_INTERVAL"); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			cfg.Metrics.Interval = duration
		}
	}

	if value := os.Getenv("TAPO_ACCOUNT_EMAIL"); value != "" {
		cfg.Account.Email = value
	}

	if value := os.Getenv("TAPO_ACCOUNT_PASSWORD"); value != "" {
		cfg.Account.Password = value
	}

	log.Printf("Processed environment variable configuration overrides")
}

func validateConfig(cfg *Config) error {
	if cfg.Account.Email == "" {
		return fmt.Errorf("Property 'account.email' or environment variable 'TAPO_ACCOUNT_EMAIL' is required")
	}

	if cfg.Account.Password == "" {
		return fmt.Errorf("Property 'account.password' or environment variable 'TAPO_ACCOUNT_PASSWORD' is required")
	}

	if cfg.Metrics.Listen == "" {
		return fmt.Errorf("Property 'metrics.listen' or environment variable 'TAPO_METRICS_LISTEN' is required")
	}

	if cfg.Metrics.Path == "" {
		return fmt.Errorf("Property 'metrics.path' or environment variable 'TAPO_METRICS_PATH' is required")
	}

	if cfg.Metrics.Interval <= 0 {
		return fmt.Errorf("Property 'metrics.interval' or environment variable 'TAPO_METRICS_INTERVAL' must be a positive duration (e.g. '10s')")
	}

	for index, plug := range cfg.Plugs {
		if plug.Name == "" {
			return fmt.Errorf("Property 'plugs[%d].name' is required", index)
		}

		if plug.IPAddress == "" {
			return fmt.Errorf("Property 'plugs[%d].ip_address' is required", index)
		}
	}

	log.Printf("Configuration file is valid")

	return nil
}

func loadConfig(filePath string) (*Config, error) {
	config := &Config{
		Metrics: MetricsConfig{
			Listen:   DEFAULT_METRICS_LISTEN,
			Path:     DEFAULT_METRICS_PATH,
			Interval: DEFAULT_METRICS_INTERVAL,
		},
	}

	if _, err := os.Stat(filePath); err == nil {
		if _, err := toml.DecodeFile(filePath, config); err != nil {
			return nil, fmt.Errorf("Error parsing config file '%q': %w", filePath, err)
		}

		log.Printf("Loaded configuration from file '%s'", filePath)
	}

	updateConfigWithEnvironmentVariables(config)

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("Invalid configuration file: %w", err)
	}

	return config, nil
}
