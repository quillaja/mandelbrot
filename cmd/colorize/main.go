package main

import (
	"flag"
	"fmt"
	mbrot "mandelbrot"
	"time"
)

func main() {

	var configFile string
	var writeDefault bool
	var verbose bool

	flag.StringVar(&configFile, "config", "", "The file with configuration data (in json format).")
	flag.BoolVar(&writeDefault, "default", false, "Set this flag to output a default config file to 'default.json'. The program will then exit.")
	flag.BoolVar(&verbose, "v", false, "Show verbose output when set.")
	flag.Parse()

	if writeDefault {
		mbrot.WriteDefault()
		return
	}

	if configFile == "" {
		panic("Config file not specified.")
	}

	cfg := mbrot.ReadConfig(configFile)
	// TODO: check config for errors

	start := time.Now() // to show processing time when finished

	// print configuration settings
	if verbose {
		fmt.Println(cfg)
		fmt.Println("Processing...")
	}

	// read data
	coords := mbrot.ReadData(cfg.DataFile)

	// output to jpg
	if verbose {
		fmt.Printf("Writing image to %s\n", cfg.ImageFile)
	}
	ramp := mbrot.MakeRamp(mbrot.ReadStops(cfg.RampFile))
	mbrot.OutputToJPG(
		mbrot.CreatePicture(coords, ramp, cfg.XRes, cfg.YRes, mbrot.HexToRGBA(cfg.SetColor)),
		cfg.ImageFile)

	if verbose {
		fmt.Println("Took", time.Since(start).Seconds(), "seconds")
	}
}
