package metrics

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/viral32111/tapo-smart-plug/internal/protocol/commands"
)

func Poll(metrics *Metrics, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if err := Update(metrics); err != nil {
			slog.Warn("Error while updating metrics", "name", metrics.Name, "ip_address", metrics.IPAddress, "error", err)
		}
	}
}

func Update(metrics *Metrics) error {
	// P304M child devices
	if metrics.ChildDeviceID != "" {
		deviceInfo, err := commands.GetChildDeviceInfo(metrics.KLAPSession, metrics.ChildDeviceID)
		if err != nil {
			return fmt.Errorf("Failed to fetch device info from child '%s' of '%s' (%s): %w", metrics.ChildDeviceID, metrics.Name, metrics.IPAddress, err)
		}

		energyUsage, err := commands.GetChildEnergyUsage(metrics.KLAPSession, metrics.ChildDeviceID)
		if err != nil {
			return fmt.Errorf("Failed to fetch energy usage from child '%s' of '%s' (%s): %w", metrics.ChildDeviceID, metrics.Name, metrics.IPAddress, err)
		}

		metrics.Update(deviceInfo, energyUsage)
		slog.Info("Updated child metrics", "name", metrics.Name, "ip_address", metrics.IPAddress, "id", deviceInfo.ID, "type", deviceInfo.Type, "model", deviceInfo.Model, "slot", deviceInfo.Slot)
	} else {
		deviceInfo, err := commands.GetDeviceInfo(metrics.KLAPSession)
		if err != nil {
			return fmt.Errorf("Failed to fetch device info from '%s' (%s): %w", metrics.Name, metrics.IPAddress, err)
		}

		energyUsage, err := commands.GetEnergyUsage(metrics.KLAPSession)
		if err != nil {
			return fmt.Errorf("Failed to fetch energy usage from '%s' (%s): %w", metrics.Name, metrics.IPAddress, err)
		}

		metrics.Update(deviceInfo, energyUsage)
		slog.Info("Updated metrics", "name", metrics.Name, "ip_address", metrics.IPAddress, "id", deviceInfo.ID, "type", deviceInfo.Type, "model", deviceInfo.Model)
	}

	return nil
}
