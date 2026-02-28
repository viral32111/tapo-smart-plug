package metrics

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/viral32111/tapo-smart-plug/internal/config"
)

type MetricsServer struct {
	Registry *prometheus.Registry

	ListenAddress string
	HttpPath      string
	PollInterval  time.Duration
}

func NewServer(config *config.Config) *MetricsServer {
	// Avoid pollution from Go's default metrics
	registry := prometheus.NewRegistry()

	return &MetricsServer{
		Registry: registry,

		ListenAddress: config.Metrics.Listen,
		HttpPath:      config.Metrics.Path,
		PollInterval:  config.Metrics.Interval,
	}
}

func (server *MetricsServer) Start() error {
	serve := http.NewServeMux()

	// Serve the metrics page
	serve.Handle(server.HttpPath, promhttp.HandlerFor(server.Registry, promhttp.HandlerOpts{
		Registry: server.Registry,

		EnableOpenMetrics:                   true,
		EnableOpenMetricsTextCreatedSamples: true,

		DisableCompression: false,

		MaxRequestsInFlight: 25,
		Timeout:             10 * time.Second,

		ProcessStartTime: time.Now(),
	}))

	// Redirect everything else to the metrics page
	serve.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		http.Redirect(response, request, server.HttpPath, http.StatusTemporaryRedirect)
	})

	slog.Info("Metrics server listening", "address", server.ListenAddress, "path", server.HttpPath)

	return http.ListenAndServe(server.ListenAddress, serve)
}
