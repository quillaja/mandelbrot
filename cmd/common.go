package cmd

import (
	"flag"
	"fmt"
	mbrot "mandelbrot"
	"os"
)

// VPrint prints str if verbose is true.
func VPrint(verbose bool, str string) {
	if verbose {
		fmt.Print(str)
	}
}

// Startup performs common startup tasks for commands, such as parsing
// command line arguments and loading the program configuration file.
func Startup() (mbrot.Config, bool) {
	var configFile string
	var writeDefault bool
	var verbose bool

	flag.StringVar(&configFile, "config", "", "The file with configuration data (in json format).")
	flag.BoolVar(&writeDefault, "default", false, "Set this flag to output a default config file to 'default.json'. The program will then exit.")
	flag.BoolVar(&verbose, "v", false, "Show verbose output when set.")
	flag.Parse()

	if writeDefault {
		mbrot.WriteDefault()
		os.Exit(0)
	}

	if configFile == "" {
		panic("Config file not specified.")
	}

	cfg := mbrot.ReadConfig(configFile)
	// TODO: check config for errors

	return cfg, verbose
}
