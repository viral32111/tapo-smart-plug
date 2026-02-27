package main

import (
	"flag"
	"fmt"
	"os"
)

func usage() {
	fmt.Fprintf(os.Stderr, `A Prometheus exporter for TP-Link Tapo smart plugs.
  https://github.com/viral32111/tapo-smart-plug

Usage:
  %s [flags]

Flags:
  -v, --version          Print version & exit.
  -h, --help             Show this help message & exit.
  -c, --config <path>    Set path to TOML config file (default: '%s').

Environment variables:
  TAPO_METRICS_LISTEN    The IP address & TCP port to listen on for Prometheus scrapes.
  TAPO_METRICS_PATH      The HTTP path to serve Prometheus metrics on.
  TAPO_ACCOUNT_EMAIL     The global TP-Link Tapo account email.
  TAPO_ACCOUNT_PASSWORD  The global TP-Link Tapo account password.

Environment variables will take precedence over their corresponding config file properties.

Copyright (C) 2026 viral32111, under the GNU AGPL v3.0 license.
`, os.Args[0], DEFAULT_CONFIG_PATH)
}

func parseFlags() (configPath string) {
	flag.Usage = usage

	var configurationFilePath string
	flag.StringVar(&configurationFilePath, "config", DEFAULT_CONFIG_PATH, "")
	flag.StringVar(&configurationFilePath, "c", DEFAULT_CONFIG_PATH, "")

	var printVersion bool
	flag.BoolVar(&printVersion, "version", false, "")
	flag.BoolVar(&printVersion, "v", false, "")

	flag.Parse()

	if printVersion {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	return configurationFilePath
}
