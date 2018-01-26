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

	var progress float64
	// progress output if verbose mode is on
	if verbose {
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			for progress < 100.0 {
				select {
				case <-ticker.C:
					// the ansi escape code here moves the cursor left 100 characters
					cmd.VPrint(verbose, fmt.Sprintf("\u001b[100D%0.1f%% complete.", progress*100))
				default:
				}
			}
			ticker.Stop()
		}()
	}

	// the data for the set
	coords := mbrot.Set{}
	coords.Initialize(cfg)              // set up
	coords.Calculate(action, &progress) // do the work

	progress = 100.0 // will stop the gofunc

	// output data
	cmd.VPrint(verbose, fmt.Sprintf("\nWriting data to %s.\n", cfg.DataFile))

	mbrot.WriteData(coords, cfg.DataFile)

	cmd.VPrint(verbose, fmt.Sprintf("Took %0.4f seconds.\n", time.Since(start).Seconds()))

}
