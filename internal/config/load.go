package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	ENVIRONMENT_VARIABLE_PREFIX = "TAPO_"
)

func updateWithEnvironmentVariables(config *Config) {
	if value := os.Getenv(ENVIRONMENT_VARIABLE_PREFIX + "METRICS_LISTEN"); value != "" {
		config.Metrics.Listen = value
	}

	if value := os.Getenv(ENVIRONMENT_VARIABLE_PREFIX + "METRICS_PATH"); value != "" {
		config.Metrics.Path = value
	}

	if value := os.Getenv(ENVIRONMENT_VARIABLE_PREFIX + "METRICS_INTERVAL"); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			config.Metrics.Interval = duration
		}
	}

	if value := os.Getenv(ENVIRONMENT_VARIABLE_PREFIX + "ACCOUNT_EMAIL"); value != "" {
		config.Account.Email = value
	}

	if value := os.Getenv(ENVIRONMENT_VARIABLE_PREFIX + "ACCOUNT_PASSWORD"); value != "" {
		config.Account.Password = value
	}
}

func validateConfig(config *Config) error {
	if config.Account.Email == "" {
		return fmt.Errorf("Property 'account.email' or environment variable '%s_ACCOUNT_EMAIL' is required", ENVIRONMENT_VARIABLE_PREFIX)
	}

	if config.Account.Password == "" {
		return fmt.Errorf("Property 'account.password' or environment variable '%s_ACCOUNT_PASSWORD' is required", ENVIRONMENT_VARIABLE_PREFIX)
	}

	if config.Metrics.Listen == "" {
		return fmt.Errorf("Property 'metrics.listen' or environment variable '%s_METRICS_LISTEN' is required", ENVIRONMENT_VARIABLE_PREFIX)
	}

	if config.Metrics.Path == "" {
		return fmt.Errorf("Property 'metrics.path' or environment variable '%s_METRICS_PATH' is required", ENVIRONMENT_VARIABLE_PREFIX)
	}

	if config.Metrics.Interval <= 0 {
		return fmt.Errorf("Property 'metrics.interval' or environment variable '%s_METRICS_INTERVAL' must be a positive duration (e.g. '10s')", ENVIRONMENT_VARIABLE_PREFIX)
	}

	for index, plugConfig := range config.Plugs {
		if plugConfig.Name == "" {
			return fmt.Errorf("Property 'plugs[%d].name' is required", index)
		}

		if plugConfig.IPAddress == "" {
			return fmt.Errorf("Property 'plugs[%d].ip_address' is required", index)
		}
	}

	return nil
}

func Load(filePath string) (*Config, error) {
	config := &Config{
		Metrics: MetricsConfig{
			Listen:   DEFAULT_METRICS_LISTEN,
			Path:     DEFAULT_METRICS_PATH,
			Interval: DEFAULT_METRICS_INTERVAL,
		},
	}

	// Only attempt to load the config file if it actually exists
	if _, err := os.Stat(filePath); err == nil {
		if _, err := toml.DecodeFile(filePath, config); err != nil {
			return nil, fmt.Errorf("Error parsing configuration file '%q': %w", filePath, err)
		}

		// Apply defaults to each plug
		for index, plugConfig := range config.Plugs {
			if plugConfig.Enabled == nil {
				config.Plugs[index].Enabled = new(bool)
				*config.Plugs[index].Enabled = true
			}

			if plugConfig.Email == nil {
				config.Plugs[index].Email = new(string)
				*config.Plugs[index].Email = config.Account.Email
			}

			if plugConfig.Password == nil {
				config.Plugs[index].Password = new(string)
				*config.Plugs[index].Password = config.Account.Password
			}

			if plugConfig.Protocol == nil {
				config.Plugs[index].Protocol = new(string)
				*config.Plugs[index].Protocol = DEFAULT_PLUG_PROTOCOL
			}
		}

		slog.Info("Loaded configuration file", "config_file_path", filePath)
	} else {
		slog.Warn("Skipped loading non-existent configuration file", "config_file_path", filePath)
	}

	updateWithEnvironmentVariables(config)
	slog.Debug("Updated configuration with environment variables overrides")

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("Invalid configuration file: %w", err)
	}
	slog.Debug("Configuration file is valid")

	return config, nil
}
