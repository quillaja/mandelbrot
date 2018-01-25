package main

import (
	"fmt"
	mbrot "mandelbrot"
	"mandelbrot/cmd"
	"time"
)

func main() {

	cfg, verbose := cmd.Startup()

	start := time.Now() // to show processing time when finished

	// print configuration settings
	cmd.VPrint(verbose, cfg.String())
	cmd.VPrint(verbose, "\n----------\n")
	cmd.VPrint(verbose, "Reading data file...\n")

	// read data
	coords := mbrot.ReadData(cfg.DataFile)

	// output to jpg
	cmd.VPrint(verbose, fmt.Sprintf("Writing image to %s\n", cfg.ImageFile))

	ramp := mbrot.MakeRamp(mbrot.ReadStops(cfg.RampFile))
	mbrot.OutputToJPG(
		mbrot.CreatePicture(coords, ramp, cfg.XRes, cfg.YRes, mbrot.HexToRGBA(cfg.SetColor)),
		cfg.ImageFile)

	cmd.VPrint(verbose, fmt.Sprintf("Took %0.4f seconds.\n", time.Since(start).Seconds()))
}
