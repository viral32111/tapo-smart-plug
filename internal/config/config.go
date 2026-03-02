package config

import (
	"time"
)

const (
	DEFAULT_METRICS_LISTEN   = "127.0.0.1:5000"
	DEFAULT_METRICS_PATH     = "/metrics"
	DEFAULT_METRICS_INTERVAL = 10 * time.Second

	DEFAULT_PLUG_PROTOCOL = "klap"
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

	Protocol *string `toml:"protocol"`
}
