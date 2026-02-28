package config

import (
	"time"
)

const (
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
