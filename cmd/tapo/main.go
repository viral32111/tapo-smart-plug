package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/viral32111/tapo-smart-plug/internal/config"
	"github.com/viral32111/tapo-smart-plug/internal/flags"
	"github.com/viral32111/tapo-smart-plug/internal/log"
	"github.com/viral32111/tapo-smart-plug/internal/metrics"
	"github.com/viral32111/tapo-smart-plug/internal/plug"
)

const (
	VERSION = "0.1.2"
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
		devices, session, err := plug.Setup(plugConfig, *plugConfig.Email, *plugConfig.Password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing plug '%s' (%s): %s\n", plugConfig.Name, plugConfig.IPAddress, err)
			continue
		}

		for _, device := range devices {
			metrics := metrics.NewMetrics(plugConfig, session, device, *plugConfig.Email, *plugConfig.Password, *plugConfig.Enabled)
			metrics.Register(metricsServer.Registry)

			metricsList = append(metricsList, metrics)
			logger.Info("Registered metrics for plug", "name", metrics.Name, "ip_address", metrics.IPAddress, "id", device.ID, "type", device.Type, "model", device.Model, "slot", device.Slot)
		}
	}

	// Update metrics in the background
	for _, _metrics := range metricsList {
		if !_metrics.EnablePolling {
			logger.Info("Skipping disabled plug", "name", _metrics.Name, "ip_address", _metrics.IPAddress)
			continue
		}

		go metrics.Poll(_metrics, config.Metrics.Interval)
		logger.Debug("Started background metrics polling", "name", _metrics.Name, "ip_address", _metrics.IPAddress, "interval", config.Metrics.Interval.String())
	}

	if err := metricsServer.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting metrics server: %s\n", err)
		os.Exit(1)
	}
}
