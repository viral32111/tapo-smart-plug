package log

import (
	"fmt"
	"log/slog"
	"os"
)

func LevelFromString(level string) (slog.Level, error) {
	switch level {

	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil

	default:
		return slog.LevelInfo, fmt.Errorf("Unknown log level '%s'", level)

	}
}

func NewLogger(level slog.Level) *slog.Logger {
	options := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewJSONHandler(os.Stdout, options)
	// handler := slog.NewTextHandler(os.Stdout, options)

	logger := slog.New(handler)

	// So internal packages can just use slog.Debug(), etc. without needing to pass a logger around
	slog.SetDefault(logger)

	return logger
}
