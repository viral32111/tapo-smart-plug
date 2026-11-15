package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Creating a new registry discards Go built-in metrics (runtime, process, etc.), so they are only added when requested.
func NewRegistry(goBuiltin bool) *prometheus.Registry {
	registry := prometheus.NewRegistry()

	if goBuiltin {
		registry.MustRegister(collectors.NewGoCollector())
		registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}

	return registry
}

func Handler(registry *prometheus.Registry) http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		Registry: registry,

		EnableOpenMetrics: true,

		DisableCompression: false,

		MaxRequestsInFlight: 25,
		Timeout:             10 * time.Second,
	})
}
