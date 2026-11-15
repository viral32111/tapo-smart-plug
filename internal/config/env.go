package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

const (
	ENVIRONMENT_VARIABLE_PREFIX = "TAPO_"
)

func parseEnvString(key, fallback string) string {
	if value := os.Getenv(ENVIRONMENT_VARIABLE_PREFIX + key); value != "" {
		return value
	}

	return fallback
}

func parseEnvNumber(key string, fallback int) (int, error) {
	if value := os.Getenv(ENVIRONMENT_VARIABLE_PREFIX + key); value != "" {
		number, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("Failed to parse %q as integer: %w", value, err)
		}

		return number, nil
	}

	return fallback, nil
}

func parseEnvBool(key string, fallback bool) (bool, error) {
	if value := os.Getenv(ENVIRONMENT_VARIABLE_PREFIX + key); value != "" {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return false, fmt.Errorf("Failed to parse %q as boolean: %w", value, err)
		}

		return b, nil
	}

	return fallback, nil
}

func parseEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	if value := os.Getenv(ENVIRONMENT_VARIABLE_PREFIX + key); value != "" {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return 0, fmt.Errorf("Failed to parse %q as duration: %w", value, err)
		}

		return duration, nil
	}

	return fallback, nil
}

func parseEnvLogLevel(key string, fallback slog.Level) (slog.Level, error) {
	if value := os.Getenv(ENVIRONMENT_VARIABLE_PREFIX + key); value != "" {
		return parseLogLevel(value, fallback)
	}

	return fallback, nil
}
