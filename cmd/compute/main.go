package main

import (
	"fmt"
	mbrot "mandelbrot"
	"mandelbrot/cmd"
	"time"
)

func main() {

	// parse command line, load config, etc
	cfg, verbose := cmd.Startup()

	// print configuration settings
	cmd.VPrint(verbose, cfg.String())
	cmd.VPrint(verbose, "\n----------\n")
	cmd.VPrint(verbose, "Processing...\n")

	start := time.Now() // to show processing time when finished

	// choose mandelbrot set or julia set
	var action mbrot.Action
	if cfg.DoJulia() {
		cmd.VPrint(verbose, "Calculating the Julia set.\n")
		action = func(n complex128) (bool, int) {
			return mbrot.IsMemberJulia(n, cfg.GetJulia(), cfg.Iterations)
		}
	} else {
		cmd.VPrint(verbose, "Calculating the Mandelbrot set.\n")
		action = func(n complex128) (bool, int) {
			return mbrot.IsMemberMandelbrot(n, cfg.Iterations)
		}
	}

	// the data for the set
	coords := mbrot.Set{}
	coords.Initialize(cfg)   // set up
	coords.Calculate(action) // do the work

	// output data
	cmd.VPrint(verbose, fmt.Sprintf("Writing data to %s.\n", cfg.DataFile))

	mbrot.WriteData(coords, cfg.DataFile)

	cmd.VPrint(verbose, fmt.Sprintf("Took %0.4f seconds.\n", time.Since(start).Seconds()))

}
