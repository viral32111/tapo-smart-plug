package flags

import (
	"flag"
)

const (
	DEFAULT_CONFIG_PATH = "/etc/tapo.toml"
	DEFAULT_LOG_LEVEL   = "info"
)

type Result struct {
	PrintVersion          bool
	ConfigurationFilePath string
	LogLevel              string
}

func Parse() Result {
	flag.Usage = printUsage

	var printVersion bool
	flag.BoolVar(&printVersion, "version", false, "")
	flag.BoolVar(&printVersion, "v", false, "")

	var configurationFilePath string
	flag.StringVar(&configurationFilePath, "config", DEFAULT_CONFIG_PATH, "")
	flag.StringVar(&configurationFilePath, "c", DEFAULT_CONFIG_PATH, "")

	var logLevel string
	flag.StringVar(&logLevel, "log-level", DEFAULT_LOG_LEVEL, "")
	flag.StringVar(&logLevel, "l", DEFAULT_LOG_LEVEL, "")

	flag.Parse()

	return Result{
		PrintVersion:          printVersion,
		ConfigurationFilePath: configurationFilePath,
		LogLevel:              logLevel,
	}
}
