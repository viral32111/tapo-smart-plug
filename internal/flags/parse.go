package flags

import (
	"flag"
	"os"
)

const (
	DEFAULT_CONFIG_PATH = "/etc/tapo.toml"
)

type Result struct {
	Version bool
	Usage   bool

	ConfigPath string
}

func Parse() Result {
	var result Result

	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagSet.Usage = func() {
		PrintUsage(os.Stderr)
	}

	flagSet.BoolVar(&result.Version, "version", false, "Print version & exit.")
	flagSet.BoolVar(&result.Version, "v", false, "Print version & exit.")

	flagSet.BoolVar(&result.Usage, "usage", false, "Show this usage message & exit.")
	flagSet.BoolVar(&result.Usage, "u", false, "Show this usage message & exit.")

	flagSet.BoolVar(&result.Usage, "help", false, "Show this usage message & exit.")
	flagSet.BoolVar(&result.Usage, "h", false, "Show this usage message & exit.")

	flagSet.StringVar(&result.ConfigPath, "config", DEFAULT_CONFIG_PATH, "Path to the TOML configuration file.")
	flagSet.StringVar(&result.ConfigPath, "c", DEFAULT_CONFIG_PATH, "Path to the TOML configuration file.")

	flagSet.Parse(os.Args[1:])

	return result
}
