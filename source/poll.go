package main

import (
	"fmt"
	"log"
	"time"
)

func pollPlugMetrics(metrics *PlugMetrics, config *Config) {
	ticker := time.NewTicker(config.Metrics.Interval)
	defer ticker.Stop()

	for range ticker.C {
		if err := updatePlugMetrics(metrics, config); err != nil {
			log.Printf("Error polling plug '%s' (%s): %s", metrics.plug.Name, metrics.plug.IPAddress, err)
		}
	}
}

func updatePlugMetrics(metrics *PlugMetrics, config *Config) error {
	deviceInfo, err := GetDeviceInfo(metrics.session)
	if err != nil {
		log.Printf("Session error for '%s': %s", metrics.plug.Name, err)
		session, err := NewKLAPSession(
			metrics.plug.IPAddress,
			metrics.plug.DetermineEmail(config.Account),
			metrics.plug.DeterminePassword(config.Account),
		)

		if err != nil {
			return fmt.Errorf("Failed to reconnect: %w", err)
		}

		metrics.session = session

		deviceInfo, err = GetDeviceInfo(session)
		if err != nil {
			return fmt.Errorf("Failed to fetch device info after reconnect: %w", err)
		}
	}

	if deviceInfo.Model == "P110" || deviceInfo.Model == "P110M" {
		energyUsage, err := GetEnergyUsage(metrics.session)
		if err != nil {
			return fmt.Errorf("Failed to fetch energy usage: %w", err)
		}

		metrics.update(deviceInfo, energyUsage)
	} else if deviceInfo.Model == "P304M" {
		// childDeviceList, err := GetChildDeviceList(metrics.session)
		// if err != nil {
		// 	return fmt.Errorf("Failed to fetch child device list: %w", err)
		// }

		// for _, childDevice := range childDeviceList.Devices {
		// 	childDeviceInfo, err := ControlChild_GetDeviceInfo(metrics.session, childDevice.ID)
		// 	if err != nil {
		// 		log.Printf("Failed to fetch device info for child device '%s' of plug '%s': %s", childDevice.ID, metrics.plug.Name, err)
		// 		continue
		// 	}

		// 	childEnergyUsage, err := ControlChild_GetEnergyUsage(metrics.session, childDevice.ID)
		// 	if err != nil {
		// 		log.Printf("Failed to fetch energy usage for child device '%s' of plug '%s': %s", childDevice.ID, metrics.plug.Name, err)
		// 		continue
		// 	}

		// 	// Fixes for missing fields that should use parent device's values
		// 	childDeviceInfo.FirmwareDescription = deviceInfo.FirmwareDescription
		// 	childDeviceInfo.HardwareDescription = deviceInfo.HardwareDescription
		// 	childDeviceInfo.IPAddress = deviceInfo.IPAddress
		// 	childDeviceInfo.MACAddress = childDevice.MACAddress
		// 	childDeviceInfo.WirelessSSID = deviceInfo.WirelessSSID
		// 	childDeviceInfo.WirelessSignalLevel = deviceInfo.WirelessSignalLevel
		// 	childDeviceInfo.Language = deviceInfo.Language

		// 	metrics.update(childDeviceInfo, childEnergyUsage)
		// }
	}

	return nil
}
