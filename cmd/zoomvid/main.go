package main

import (
	"fmt"
	m "mandelbrot"
	"mandelbrot/cmd"
	"math"
	"time"
)

func main() {
	cfg, verbose := cmd.Startup()
	if verbose {
		//fuck you
	}

	// alter plot_width,plot_height, iterations, image_file in order,
	// producing a series of images which 'zoom' into the configured point.
	// do this non-concurrently to save CPU for mandelbrot calcs and to prevent
	// excessive use of memory (keeping all the Sets in memory).

	ramp := m.MakeRamp(m.ReadStops(cfg.RampFile))
	setColor := m.HexToRGBA(cfg.SetColor)

	const zoomFactor = 1.1
	const iterFactor = 2

	finalZoomLevel := cfg.PlotWidth
	for i := 0; cfg.PlotWidth >= finalZoomLevel; i++ {
		start := time.Now()

		cfg.PlotWidth = 4.0 / math.Pow(zoomFactor, float64(i))
		cfg.PlotHeight = cfg.PlotWidth
		cfg.ImageFile = fmt.Sprintf("zoom/%010d.jpg", i)
		if i > 0 && i%20 == 0 {
			//increase iterations
			cfg.Iterations = cfg.Iterations * iterFactor
		}
		action := func(n complex128) (bool, int) {
			return m.IsMemberMandelbrot(n, cfg.Iterations)
		}

		setProgress := 0.0
		if verbose {
			fmt.Printf("Frame %d\n", i)
			fmt.Printf("Iterations: %d\n", cfg.Iterations)
			fmt.Printf(" Plot width: %0.8e. Set progress:\n", cfg.PlotWidth)
			showProgress(&setProgress)
		}

		coords := m.Set{}
		coords.Initialize(cfg)
		coords.Calculate(action, &setProgress)

		setProgress = 100.0 // end setProgress
		took := time.Since(start).Seconds()
		if verbose {
			fmt.Printf("\n Took %0.1f seconds.\n\n", took)
		}

		img := m.CreatePicture(coords, ramp, cfg.XRes, cfg.YRes, setColor)
		m.OutputToJPG(img, cfg.ImageFile)

	}
}

func showProgress(progress *float64) {
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		for *progress < 100.0 {
			select {
			case <-ticker.C:
				// the ansi escape code here moves the cursor left 100 characters
				fmt.Printf(" \u001b[100D%0.1f%% complete.", *progress*100)
			default:
			}
		}
		ticker.Stop()
	}()
}
