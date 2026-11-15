package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"

	"github.com/viral32111/tapo-smart-plug/internal/config"
	"github.com/viral32111/tapo-smart-plug/internal/flags"
	"github.com/viral32111/tapo-smart-plug/internal/log"
	"github.com/viral32111/tapo-smart-plug/internal/metrics"
	"github.com/viral32111/tapo-smart-plug/internal/plug"
	"github.com/viral32111/tapo-smart-plug/internal/server"
)

// -ldflags="-X main.VERSION=X.X.X"
var VERSION = "0.0.0"

func main() {
	isLocalDevelopment := VERSION == "0.0.0"

	/***********************/

	flagResult := flags.Parse()

	if flagResult.Version {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	if flagResult.Usage {
		flags.PrintUsage(os.Stdout)
		os.Exit(0)
	}

	/***********************/

	// Intended for local development, silently fails if no .env file is found
	godotenv.Load()

	config, err := config.GetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to load configuration: %s\n", err)
		os.Exit(1)
	}

	// Loaded before the logger exists, as the configuration file can set the log level
	configFileLoaded, err := config.LoadFile(flagResult.ConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to load configuration file: %s\n", err)
		os.Exit(1)
	}

	/***********************/

	logger := log.NewLogger(config.LogLevel, isLocalDevelopment)

	logger.Info("Hello World", "version", VERSION)

	if configFileLoaded {
		logger.Info("Loaded configuration file", "config_file_path", flagResult.ConfigPath)
	} else {
		logger.Warn("Skipped loading non-existent configuration file", "config_file_path", flagResult.ConfigPath)
	}

	config.LogLoaded()

	/***********************/

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	group, ctx := errgroup.WithContext(ctx)

	/***********************/

	httpServer, err := server.New(config)
	if err != nil {
		logger.Error("Unable to create HTTP server", "error", err)
		os.Exit(1)
	}

	prometheusRegistry := metrics.NewRegistry(config.MetricsGoBuiltin)

	// Serve metrics on the configured path
	httpServer.HTTPMultiplexer.Handle(config.MetricsPath, metrics.Handler(prometheusRegistry))

	// Start HTTP server in the background
	group.Go(func() error {
		listenAddress := net.JoinHostPort(config.HTTPAddress, strconv.Itoa(config.HTTPPort))
		if err := httpServer.Start(ctx, listenAddress); err != nil {
			return fmt.Errorf("HTTP server error: %w", err)
		}

		return nil
	})

	/***********************/

	var metricsList []*metrics.Metrics
	for _, plugConfig := range config.Plugs {
		if ctx.Err() != nil {
			break
		}

		devices, session, err := plug.Setup(plugConfig, *plugConfig.Email, *plugConfig.Password)
		if err != nil {
			logger.Error("Error initializing plug", "name", plugConfig.Name, "ip_address", plugConfig.IPAddress, "error", err)
			continue
		}

		for _, device := range devices {
			deviceMetrics := metrics.NewMetrics(plugConfig, session, device, *plugConfig.Email, *plugConfig.Password, *plugConfig.Enabled)
			deviceMetrics.Register(prometheusRegistry)

			metricsList = append(metricsList, deviceMetrics)
			logger.Info("Registered metrics for plug", "name", deviceMetrics.Name, "ip_address", deviceMetrics.IPAddress, "id", device.ID, "type", device.Type, "model", device.Model, "slot", device.Slot)
		}
	}

	// Update metrics in the background, one job per enabled plug
	for _, deviceMetrics := range metricsList {
		if !deviceMetrics.EnablePolling {
			logger.Info("Skipping disabled plug", "name", deviceMetrics.Name, "ip_address", deviceMetrics.IPAddress)
			continue
		}

		group.Go(func() error {
			metrics.Poll(ctx, deviceMetrics, config.Metrics.Interval)

			return nil
		})
		logger.Debug("Started background metrics polling", "name", deviceMetrics.Name, "ip_address", deviceMetrics.IPAddress, "interval", config.Metrics.Interval.String())
	}

	/***********************/

	if err := group.Wait(); err != nil {
		logger.Error("Error in execution group", "error", err)
		os.Exit(1)
	}
}
