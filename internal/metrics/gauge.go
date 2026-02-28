package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Gauges struct {
	Status        prometheus.Gauge
	UptimeSeconds prometheus.Gauge

	// Energy monitoring
	CurrentMilliwatts   prometheus.Gauge
	TodayWattHours      prometheus.Gauge
	TodayRuntimeMinutes prometheus.Gauge
	MonthWattHours      prometheus.Gauge
	MonthRuntimeMinutes prometheus.Gauge
}

func newGauge(name string, help string, labels prometheus.Labels) prometheus.Gauge {
	return prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: EXPORTER_NAMESPACE,

		Name: name,
		Help: help,

		ConstLabels: labels,
	})
}
