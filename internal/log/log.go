package log

import (
	"log/slog"
	"os"
)

func NewLogger(level slog.Level, isLocalDevelopment bool) *slog.Logger {
	options := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if isLocalDevelopment {
		handler = slog.NewTextHandler(os.Stdout, options)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, options)
	}

	logger := slog.New(handler)
	// So internal packages can just use slog.Debug(), etc. without needing to pass a logger around
	slog.SetDefault(logger)

	return logger
}
