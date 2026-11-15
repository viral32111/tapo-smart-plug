package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/viral32111/tapo-smart-plug/internal/config"
)

type Server struct {
	Config          config.Config
	HTTPMultiplexer *http.ServeMux
}

func New(config config.Config) (*Server, error) {
	server := &Server{
		Config:          config,
		HTTPMultiplexer: http.NewServeMux(),
	}

	server.RegisterRoutes()
	slog.Debug("Registered HTTP routes")

	return server, nil
}

func (this *Server) Start(ctx context.Context, listenAddress string) error {
	slog.Debug("Starting HTTP server", "address", listenAddress)

	httpServer := &http.Server{
		Addr:    listenAddress,
		Handler: this.HTTPMultiplexer,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("Error stopping HTTP server", "error", err)
		}
	}()

	slog.Info("HTTP server listening", "address", listenAddress)
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("Unable to start HTTP server on '%s': %w", listenAddress, err)
	}

	slog.Info("HTTP server stopped")

	return nil
}

func (this *Server) RegisterRoutes() {
	this.HTTPMultiplexer.Handle("/", http.HandlerFunc(this.NotFoundHandler))
}

func (this *Server) NotFoundHandler(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(404)
	json.NewEncoder(response).Encode(map[string]any{
		"error": "Unknown route",
	})
}
