package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	VERSION = "0.1.0"
)

type Plug struct {
	config  PlugConfig
	metrics *PlugMetrics
}

func main() {
	godotenv.Load()

	configPath := parseFlags()
	log.Printf("Using '%s' as configuration file path", configPath)

	config, err := loadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while loading configuration file: %s\n", err)
		os.Exit(1)
	}

	metricsServer := NewMetricsServer(config)

	var metricsList []*PlugMetrics
	for _, plugConfig := range config.Plugs {
		if !plugConfig.IsEnabled() {
			continue
		}

		session, err := NewKLAPSession(
			plugConfig.IPAddress,
			plugConfig.DetermineEmail(config.Account),
			plugConfig.DeterminePassword(config.Account),
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error communicating with plug '%s' (%s): %s\n", plugConfig.Name, plugConfig.IPAddress, err)
			continue
		}

		deviceInfo, err := GetDeviceInfo(session)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting device info from plug '%s' (%s): %s\n", plugConfig.Name, plugConfig.IPAddress, err)
			continue
		}

		// Extension cord
		if deviceInfo.Model == "P110" || deviceInfo.Model == "P110M" {
			energyUsage, err := GetEnergyUsage(session)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting energy usage from plug '%s' (%s): %s\n", plugConfig.Name, plugConfig.IPAddress, err)
			}

			metrics := NewPlugMetrics(plugConfig, session, deviceInfo, energyUsage)
			metrics.register(metricsServer.registry)

			metricsList = append(metricsList, metrics)
			log.Printf("Registered metrics for '%s' (%s) @ %s", plugConfig.Name, deviceInfo.Model, deviceInfo.IPAddress)

			// Initial metrics population before starting the server
			metrics.update(deviceInfo, energyUsage)

		} else if deviceInfo.Model == "P304M" {
			childDeviceList, err := GetChildDeviceList(session)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting child device list from plug '%s' (%s): %s\n", plugConfig.Name, plugConfig.IPAddress, err)
				continue
			}

			for _, childDevice := range childDeviceList.Devices {
				childDeviceInfo, err := ControlChild_GetDeviceInfo(session, childDevice.ID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting device info for child device '%s' of plug '%s' (%s): %s\n", childDevice.ID, plugConfig.Name, plugConfig.IPAddress, err)
					continue
				}

				childEnergyUsage, err := ControlChild_GetEnergyUsage(session, childDevice.ID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting energy usage for child device '%s' of plug '%s' (%s): %s\n", childDevice.ID, plugConfig.Name, plugConfig.IPAddress, err)
					continue
				}

				// Fixes for missing fields that should use parent device's values
				childDeviceInfo.FirmwareDescription = deviceInfo.FirmwareDescription
				childDeviceInfo.HardwareDescription = deviceInfo.HardwareDescription
				childDeviceInfo.IPAddress = deviceInfo.IPAddress
				childDeviceInfo.MACAddress = childDevice.MACAddress
				childDeviceInfo.WirelessSSID = deviceInfo.WirelessSSID
				childDeviceInfo.WirelessSignalLevel = deviceInfo.WirelessSignalLevel
				childDeviceInfo.Language = deviceInfo.Language

				metrics := NewPlugMetrics(plugConfig, session, childDeviceInfo, childEnergyUsage)
				metrics.register(metricsServer.registry)

				metricsList = append(metricsList, metrics)

				// Initial metrics population before starting the server
				metrics.update(childDeviceInfo, childEnergyUsage)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Unsupported device model '%s' for plug '%s' (%s)\n", deviceInfo.Model, plugConfig.Name, plugConfig.IPAddress)
		}

	}

	// Update metrics in the background
	for _, metrics := range metricsList {
		go pollPlugMetrics(metrics, config)
		log.Printf("Will update metrics for '%s' (%s) every %s", metrics.plug.Name, metrics.plug.IPAddress, config.Metrics.Interval)
	}

	if err := metricsServer.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error while starting metrics server: %s\n", err)
		os.Exit(1)
	}
}
