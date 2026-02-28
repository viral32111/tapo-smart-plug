package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/viral32111/tapo-smart-plug/internal/config"
	"github.com/viral32111/tapo-smart-plug/internal/flags"
	"github.com/viral32111/tapo-smart-plug/internal/klap"
	"github.com/viral32111/tapo-smart-plug/internal/log"
	"github.com/viral32111/tapo-smart-plug/internal/metrics"
	"github.com/viral32111/tapo-smart-plug/internal/protocol/commands"
)

const (
	VERSION = "0.1.0"
)

func main() {
	// Intended for local development, silently fails if no .env file is found
	godotenv.Load()

	flags := flags.Parse()

	// --version / -v
	if flags.PrintVersion {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	// --log-level / -l
	logLevel, err := log.LevelFromString(flags.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing log level: %s\n", err)
		os.Exit(1)
	}

	logger := log.NewLogger(logLevel)

	// --config / -c
	config, err := config.Load(flags.ConfigurationFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration file: %s\n", err)
		os.Exit(1)
	}

	metricsServer := metrics.NewServer(config)

	var metricsList []*metrics.Metrics
	for _, plugConfig := range config.Plugs {
		/* if !plugConfig.IsEnabled() {
			continue
		} */

		emailAddress := plugConfig.DetermineEmail(config.Account)
		password := plugConfig.DeterminePassword(config.Account)

		klapSession, err := klap.NewSession(plugConfig.IPAddress, emailAddress, password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error communicating with plug '%s' (%s): %s\n", plugConfig.Name, plugConfig.IPAddress, err)
			continue
		}

		deviceInfo, err := commands.GetDeviceInfo(klapSession)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching device info from '%s' (%s): %s\n", plugConfig.Name, plugConfig.IPAddress, err)
			continue
		}

		// Only add child plugs for extension cords
		if deviceInfo.Model == "P304M" {
			childList, err := commands.GetChildDeviceList(klapSession)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error fetching child device list from '%s' (%s): %s\n", plugConfig.Name, plugConfig.IPAddress, err)
				continue
			}

			for _, child := range childList.Devices {
				childDeviceInfo, err := commands.GetChildDeviceInfo(klapSession, child.ID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error fetching device info for child '%s' of '%s' (%s): %s\n", child.ID, plugConfig.Name, plugConfig.IPAddress, err)
					continue
				}

				// Fixes for missing fields that should use parent device's values
				childDeviceInfo.FirmwareDescription = deviceInfo.FirmwareDescription
				childDeviceInfo.HardwareDescription = deviceInfo.HardwareDescription
				childDeviceInfo.IPAddress = deviceInfo.IPAddress
				childDeviceInfo.MACAddress = child.MACAddress
				childDeviceInfo.WirelessSSID = deviceInfo.WirelessSSID
				childDeviceInfo.WirelessSignalLevel = deviceInfo.WirelessSignalLevel
				childDeviceInfo.Language = deviceInfo.Language

				metrics := metrics.NewMetrics(plugConfig, klapSession, childDeviceInfo, emailAddress, password)
				metrics.Register(metricsServer.Registry)

				metricsList = append(metricsList, metrics)
				logger.Info("Registered child metrics", "name", metrics.Name, "ip_address", metrics.IPAddress, "id", childDeviceInfo.ID, "type", childDeviceInfo.Type, "model", childDeviceInfo.Model, "slot", childDeviceInfo.Slot)
			}
		} else {
			metrics := metrics.NewMetrics(plugConfig, klapSession, deviceInfo, emailAddress, password)
			metrics.Register(metricsServer.Registry)

			metricsList = append(metricsList, metrics)
			logger.Info("Registered metrics", "name", metrics.Name, "ip_address", metrics.IPAddress, "id", deviceInfo.ID, "type", deviceInfo.Type, "model", deviceInfo.Model)
		}
	}

	// Update metrics in the background
	for _, _metrics := range metricsList {
		go metrics.Poll(_metrics, config.Metrics.Interval)
		logger.Debug("Started background metrics polling", "name", _metrics.Name, "ip_address", _metrics.IPAddress, "interval", config.Metrics.Interval.String())
	}

	if err := metricsServer.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting metrics server: %s\n", err)
		os.Exit(1)
	}
}
