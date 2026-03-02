package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/viral32111/tapo-smart-plug/internal/config"
	"github.com/viral32111/tapo-smart-plug/internal/klap"
	"github.com/viral32111/tapo-smart-plug/internal/protocol/commands"
)

const (
	EXPORTER_NAMESPACE = "tapo"
)

type Metrics struct {
	Name      string
	IPAddress string

	EnablePolling bool
	EmailAddress  string
	Password      string

	ChildDeviceID string

	KLAPSession *klap.Session

	Gauges Gauges
}

func NewMetrics(plugConfig config.PlugConfig, klapSession *klap.Session, deviceInfo *commands.DeviceInfo, emailAddress string, password string, enablePolling bool) *Metrics {
	labels := prometheus.Labels{
		"id": deviceInfo.ID,

		"type":  deviceInfo.Type,
		"model": deviceInfo.Model,
		"name": func() string {
			if deviceInfo.Name != nil {
				return *deviceInfo.Name
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

		// Only for P304M child devices
		"slot": func() string {
			if deviceInfo.Slot != nil {
				return strconv.Itoa(*deviceInfo.Slot)
			}

			return ""
		}(),
	}

	return &Metrics{
		Name:      plugConfig.Name,
		IPAddress: plugConfig.IPAddress,

		EnablePolling: enablePolling,
		EmailAddress:  emailAddress,
		Password:      password,

		// Only for P304M child devices
		ChildDeviceID: func() string {
			if deviceInfo.Slot != nil {
				return deviceInfo.ID
			}

			return ""
		}(),

		KLAPSession: klapSession,

		Gauges: Gauges{
			Status:        newGauge("status", "1 if the smart plug is switched on, 0 if switched off", labels),
			UptimeSeconds: newGauge("uptime_seconds", "Number of seconds the smart plug has been switched on for", labels),

			CurrentMilliwatts:   newGauge("current_milliwatts", "Current power consumption in milliwatts", labels),
			TodayWattHours:      newGauge("today_watt_hours", "Energy consumed today in watt-hours", labels),
			TodayRuntimeMinutes: newGauge("today_runtime_minutes", "Runtime today in minutes", labels),
			MonthWattHours:      newGauge("month_watt_hours", "Energy consumed this month in watt-hours", labels),
			MonthRuntimeMinutes: newGauge("month_runtime_minutes", "Runtime this month in minutes", labels),
		},
	}
}

func (metrics *Metrics) Register(registry *prometheus.Registry) {
	collectors := []prometheus.Collector{
		metrics.Gauges.Status,
		metrics.Gauges.UptimeSeconds,
		metrics.Gauges.CurrentMilliwatts,
		metrics.Gauges.TodayWattHours,
		metrics.Gauges.TodayRuntimeMinutes,
		metrics.Gauges.MonthWattHours,
		metrics.Gauges.MonthRuntimeMinutes,
	}

	for _, collector := range collectors {
		registry.MustRegister(collector)
	}
}

func (metrics *Metrics) Update(deviceInfo *commands.DeviceInfo, energyUsage *commands.EnergyUsage) {
	if deviceInfo.SwitchStatus != nil {
		if *deviceInfo.SwitchStatus {
			metrics.Gauges.Status.Set(1)
		} else {
			metrics.Gauges.Status.Set(0)
		}
	}

	if deviceInfo.UptimeSeconds != nil {
		metrics.Gauges.UptimeSeconds.Set(float64(*deviceInfo.UptimeSeconds))
	}

	// Energy monitoring
	metrics.Gauges.CurrentMilliwatts.Set(float64(energyUsage.CurrentEnergy))
	metrics.Gauges.TodayWattHours.Set(float64(energyUsage.TodayEnergy))
	metrics.Gauges.TodayRuntimeMinutes.Set(float64(energyUsage.TodayRuntime))
	metrics.Gauges.MonthWattHours.Set(float64(energyUsage.MonthEnergy))
	metrics.Gauges.MonthRuntimeMinutes.Set(float64(energyUsage.MonthRuntime))
}
