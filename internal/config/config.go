package config

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/BurntSushi/toml"
)

const (
	DEFAULT_LOG_LEVEL = slog.LevelInfo

	DEFAULT_HTTP_ADDRESS = "127.0.0.1"
	DEFAULT_HTTP_PORT    = 5000

	DEFAULT_METRICS_PATH       = "/metrics"
	DEFAULT_METRICS_GO_BUILTIN = false

	DEFAULT_METRICS_INTERVAL = 10 * time.Second

	DEFAULT_PLUG_PROTOCOL = "klap"
)

type Config struct {
	LogLevel slog.Level `toml:"-"`

	HTTPAddress string `toml:"-"`
	HTTPPort    int    `toml:"-"`

	MetricsPath      string `toml:"-"`
	MetricsGoBuiltin bool   `toml:"-"`

	Log     LogConfig     `toml:"log"`
	Metrics MetricsConfig `toml:"metrics"`
	Account AccountConfig `toml:"account"`
	Plugs   []PlugConfig  `toml:"plugs"`
}

type LogConfig struct {
	Level string `toml:"level"`
}

type MetricsConfig struct {
	Listen    string        `toml:"listen"`
	Path      string        `toml:"path"`
	GoBuiltin bool          `toml:"go_builtin"`
	Interval  time.Duration `toml:"interval"`
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

func GetConfig() (Config, error) {
	logLevel, err := parseEnvLogLevel("LOG_LEVEL", DEFAULT_LOG_LEVEL)
	if err != nil {
		return Config{}, fmt.Errorf("Unable to parse environment variable '%s%s': %w", ENVIRONMENT_VARIABLE_PREFIX, "LOG_LEVEL", err)
	}

	httpPort, err := parseEnvNumber("HTTP_PORT", DEFAULT_HTTP_PORT)
	if err != nil {
		return Config{}, fmt.Errorf("Unable to parse environment variable '%s%s': %w", ENVIRONMENT_VARIABLE_PREFIX, "HTTP_PORT", err)
	}

	if err := validatePort(httpPort); err != nil {
		return Config{}, fmt.Errorf("Invalid value for environment variable '%s%s': %w", ENVIRONMENT_VARIABLE_PREFIX, "HTTP_PORT", err)
	}

	metricsGoBuiltin, err := parseEnvBool("METRICS_GO_BUILTIN", DEFAULT_METRICS_GO_BUILTIN)
	if err != nil {
		return Config{}, fmt.Errorf("Unable to parse environment variable '%s%s': %w", ENVIRONMENT_VARIABLE_PREFIX, "METRICS_GO_BUILTIN", err)
	}

	metricsPath := parseEnvString("METRICS_PATH", DEFAULT_METRICS_PATH)

	if err := validateMetricsPath(metricsPath); err != nil {
		return Config{}, fmt.Errorf("Invalid value for environment variable '%s%s': %w", ENVIRONMENT_VARIABLE_PREFIX, "METRICS_PATH", err)
	}

	metricsInterval, err := parseEnvDuration("METRICS_INTERVAL", DEFAULT_METRICS_INTERVAL)
	if err != nil {
		return Config{}, fmt.Errorf("Unable to parse environment variable '%s%s': %w", ENVIRONMENT_VARIABLE_PREFIX, "METRICS_INTERVAL", err)
	}

	if metricsInterval <= 0 {
		return Config{}, fmt.Errorf("Invalid value for environment variable '%s%s': Interval %s must be greater than zero", ENVIRONMENT_VARIABLE_PREFIX, "METRICS_INTERVAL", metricsInterval)
	}

	config := Config{
		LogLevel: logLevel,

		HTTPAddress: parseEnvString("HTTP_ADDRESS", DEFAULT_HTTP_ADDRESS),
		HTTPPort:    httpPort,

		MetricsPath:      metricsPath,
		MetricsGoBuiltin: metricsGoBuiltin,

		Metrics: MetricsConfig{
			Interval: metricsInterval,
		},
		Account: AccountConfig{
			Email:    parseEnvString("ACCOUNT_EMAIL", ""),
			Password: parseEnvString("ACCOUNT_PASSWORD", ""),
		},
	}

	return config, nil
}

func (config *Config) LoadFile(filePath string) (bool, error) {
	fileLoaded := false

	// Only attempt to load the config file if it actually exists
	if _, err := os.Stat(filePath); err == nil {
		fileConfig := Config{}

		metadata, err := toml.DecodeFile(filePath, &fileConfig)
		if err != nil {
			return false, fmt.Errorf("Error parsing configuration file %q: %w", filePath, err)
		}

		if metadata.IsDefined("log", "level") {
			logLevel, err := parseLogLevel(fileConfig.Log.Level, DEFAULT_LOG_LEVEL)
			if err != nil {
				return false, fmt.Errorf("Invalid property 'log.level' in configuration file %q: %w", filePath, err)
			}

			if !isEnvSet("LOG_LEVEL") {
				config.LogLevel = logLevel
			}
		}

		if metadata.IsDefined("metrics", "go_builtin") && !isEnvSet("METRICS_GO_BUILTIN") {
			config.MetricsGoBuiltin = fileConfig.Metrics.GoBuiltin
		}

		if metadata.IsDefined("metrics", "listen") {
			host, portString, err := net.SplitHostPort(fileConfig.Metrics.Listen)
			if err != nil {
				return false, fmt.Errorf("Invalid property 'metrics.listen' in configuration file %q: %w", filePath, err)
			}

			port, err := strconv.Atoi(portString)
			if err != nil {
				return false, fmt.Errorf("Invalid property 'metrics.listen' in configuration file %q: Failed to parse port %q as integer: %w", filePath, portString, err)
			}

			if err := validatePort(port); err != nil {
				return false, fmt.Errorf("Invalid property 'metrics.listen' in configuration file %q: %w", filePath, err)
			}

			if !isEnvSet("HTTP_ADDRESS") {
				config.HTTPAddress = host
			}

			if !isEnvSet("HTTP_PORT") {
				config.HTTPPort = port
			}
		}

		if metadata.IsDefined("metrics", "path") {
			if err := validateMetricsPath(fileConfig.Metrics.Path); err != nil {
				return false, fmt.Errorf("Invalid property 'metrics.path' in configuration file %q: %w", filePath, err)
			}

			if !isEnvSet("METRICS_PATH") {
				config.MetricsPath = fileConfig.Metrics.Path
			}
		}

		if metadata.IsDefined("metrics", "interval") && !isEnvSet("METRICS_INTERVAL") {
			config.Metrics.Interval = fileConfig.Metrics.Interval
		}

		if !isEnvSet("ACCOUNT_EMAIL") {
			config.Account.Email = fileConfig.Account.Email
		}

		if !isEnvSet("ACCOUNT_PASSWORD") {
			config.Account.Password = fileConfig.Account.Password
		}

		config.Plugs = fileConfig.Plugs

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

		fileLoaded = true
	}

	if err := config.validate(); err != nil {
		return false, fmt.Errorf("Invalid configuration: %w", err)
	}

	return fileLoaded, nil
}

func (config *Config) LogLoaded() {
	slog.Debug("Loaded configuration",
		"log_level", config.LogLevel,

		"http_address", config.HTTPAddress,
		"http_port", config.HTTPPort,

		"metrics_path", config.MetricsPath,
		"metrics_go_builtin", config.MetricsGoBuiltin,
		"metrics_interval", config.Metrics.Interval,

		"plugs", len(config.Plugs),
	)
}

func (config *Config) validate() error {
	if config.Account.Email == "" {
		return fmt.Errorf("Property 'account.email' or environment variable '%sACCOUNT_EMAIL' is required", ENVIRONMENT_VARIABLE_PREFIX)
	}

	if config.Account.Password == "" {
		return fmt.Errorf("Property 'account.password' or environment variable '%sACCOUNT_PASSWORD' is required", ENVIRONMENT_VARIABLE_PREFIX)
	}

	if config.Metrics.Interval <= 0 {
		return fmt.Errorf("Property 'metrics.interval' or environment variable '%sMETRICS_INTERVAL' must be a positive duration (e.g. '10s')", ENVIRONMENT_VARIABLE_PREFIX)
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

func isEnvSet(key string) bool {
	return os.Getenv(ENVIRONMENT_VARIABLE_PREFIX+key) != ""
}

func validatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("Port %d must be between 1 and 65535", port)
	}

	return nil
}

func validateMetricsPath(path string) error {
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("Path %q must start with a slash", path)
	}

	if len(path) > 1 && strings.HasSuffix(path, "/") {
		return fmt.Errorf("Path %q must not end with a slash", path)
	}

	if strings.ContainsFunc(path, unicode.IsSpace) {
		return fmt.Errorf("Path %q must not contain whitespace", path)
	}

	return nil
}

func parseLogLevel(value string, fallback slog.Level) (slog.Level, error) {
	switch value {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return fallback, fmt.Errorf("Failed to parse %q as log level: Must be either debug, info, warn, error", value)
	}
}
