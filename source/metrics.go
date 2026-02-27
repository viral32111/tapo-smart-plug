package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	EXPORTER_NAMESPACE = "tapo"
)

/************************************************************/

type PlugMetrics struct {
	plug    PlugConfig
	session *KLAPSession

	status prometheus.Gauge

	currentMilliwatts   prometheus.Gauge
	todayWattHours      prometheus.Gauge
	todayRuntimeMinutes prometheus.Gauge
	monthWattHours      prometheus.Gauge
	monthRuntimeMinutes prometheus.Gauge

	uptimeSeconds prometheus.Gauge
}

type MetricsServer struct {
	config   *Config
	registry *prometheus.Registry
}

/************************************************************/

func NewPlugMetrics(plug PlugConfig, session *KLAPSession, deviceInfo *DeviceInfo, energyUsage *EnergyUsage) *PlugMetrics {
	createGauge := func(name, help string) prometheus.Gauge {
		return prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: EXPORTER_NAMESPACE,

			Name: name,
			Help: help,

			ConstLabels: prometheus.Labels{
				"id": deviceInfo.ID,

				"type":  deviceInfo.Type,
				"model": deviceInfo.Model,
				"name": func() string {
					if deviceInfo.Name != nil {
						return *deviceInfo.Name
					}

					return ""
				}(),
				"slot": func() string {
					if deviceInfo.Slot != nil {
						return strconv.Itoa(*deviceInfo.Slot)
					}

					return ""
				}(),

				"firmware": deviceInfo.FirmwareDescription,
				"hardware": deviceInfo.HardwareDescription,

				"ssid":     deviceInfo.WirelessSSID,
				"region":   deviceInfo.Region,
				"language": deviceInfo.Language,

				"ip_address":  deviceInfo.IPAddress,
				"mac_address": deviceInfo.MACAddress,
			},
		})
	}

	return &PlugMetrics{
		plug:    plug,
		session: session,

		status: createGauge("status", "1 if the smart plug is switched on, 0 if switched off"),

		currentMilliwatts:   createGauge("current_milliwatts", "Current power consumption in milliwatts"),
		todayWattHours:      createGauge("today_watt_hours", "Energy consumed today in watt-hours"),
		todayRuntimeMinutes: createGauge("today_runtime_minutes", "Runtime today in minutes"),
		monthWattHours:      createGauge("month_watt_hours", "Energy consumed this month in watt-hours"),
		monthRuntimeMinutes: createGauge("month_runtime_minutes", "Runtime this month in minutes"),

		uptimeSeconds: createGauge("uptime_seconds", "Number of seconds the smart plug has been switched on for"),
	}
}

func (metrics *PlugMetrics) register(registry *prometheus.Registry) {
	registry.MustRegister(
		metrics.status,
		metrics.currentMilliwatts,
		metrics.todayWattHours,
		metrics.todayRuntimeMinutes,
		metrics.monthWattHours,
		metrics.monthRuntimeMinutes,
		metrics.uptimeSeconds,
	)
}

func (metrics *PlugMetrics) update(deviceInfo *DeviceInfo, energyUsage *EnergyUsage) {
	if deviceInfo.SwitchStatus != nil {
		if *deviceInfo.SwitchStatus {
			metrics.status.Set(1)
		} else {
			metrics.status.Set(0)
		}
	}

	metrics.currentMilliwatts.Set(float64(energyUsage.CurrentEnergy))
	metrics.todayWattHours.Set(float64(energyUsage.TodayEnergy))
	metrics.todayRuntimeMinutes.Set(float64(energyUsage.TodayRuntime))
	metrics.monthWattHours.Set(float64(energyUsage.MonthEnergy))
	metrics.monthRuntimeMinutes.Set(float64(energyUsage.MonthRuntime))

	if deviceInfo.UptimeSeconds != nil {
		metrics.uptimeSeconds.Set(float64(*deviceInfo.UptimeSeconds))
	}

	log.Printf("Updated metrics for '%s' (%s)", metrics.plug.Name, deviceInfo.IPAddress)
}

/************************************************************/

func NewMetricsServer(config *Config) *MetricsServer {
	registry := prometheus.NewRegistry() // Avoid Go's default global registry

	return &MetricsServer{
		config:   config,
		registry: registry,
	}
}

func (server *MetricsServer) Start() error {
	mux := http.NewServeMux()

	// Serve the metrics page
	mux.Handle(server.config.Metrics.Path, promhttp.HandlerFor(server.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))

	// Redirect everything else to the metrics page
	mux.HandleFunc("/", func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		http.Redirect(responseWriter, httpRequest, server.config.Metrics.Path, http.StatusTemporaryRedirect)
	})

	log.Printf("Listening on http://%s%s", server.config.Metrics.Listen, server.config.Metrics.Path)

	return http.ListenAndServe(server.config.Metrics.Listen, mux)
}
