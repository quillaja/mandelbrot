package main

import (
	"fmt"
	m "mandelbrot"
	"mandelbrot/cmd"
	"math"
	"time"
)

// TODO:
// parameterize the output directory,
// make the program create the directory if it doesn't exist,
// parameterize zoom factor and iterFactor (etc)
// parameterize starting 'zoom' (width of 4.0)
// ?? stop after N frames instead of going until reaching Config.PlotWidth?
// ?? aspect ratio?

func main() {
	cfg, verbose := cmd.Startup()

	// alter plot_width,plot_height, iterations, image_file in order,
	// producing a series of images which 'zoom' into the configured point.
	// do this non-concurrently to save CPU for mandelbrot calcs and to prevent
	// excessive use of memory (keeping all the Sets in memory).

	ramp := m.MakeRamp(m.ReadStops(cfg.RampFile))
	setColor := m.HexToRGBA(cfg.SetColor)

	const startWidth = 4.0
	const zoomFactor = 1.1
	const iterFactor = 25

	totalFrames := totalFrames(zoomFactor, cfg.PlotWidth) - 1
	totalTime := 0.0

	finalZoomLevel := cfg.PlotWidth
	origIterations := cfg.Iterations
	for i := 0; cfg.PlotWidth >= finalZoomLevel; i++ {
		start := time.Now()

		// setup parameters for this frame
		cfg.PlotWidth = startWidth / math.Pow(zoomFactor, float64(i))
		cfg.PlotHeight = cfg.PlotWidth
		// double the number of iterations every iterFactor (25) frames
		cfg.Iterations = origIterations * 1 << uint(i/iterFactor)
		cfg.ImageFile = fmt.Sprintf("zoom/%010d.jpg", i)
		action := func(n complex128) (bool, int) {
			return m.IsMemberMandelbrot(n, cfg.Iterations)
		}

		// show status
		var setProgress float64
		if verbose {
			fmt.Printf("Frame %d of %d\n", i, totalFrames)
			fmt.Printf(" Iterations: %d\n", cfg.Iterations)
			fmt.Printf(" Plot width: %0.8e. Set progress:\n", cfg.PlotWidth)
			showProgress(&setProgress)
		}

		// do work
		coords := m.Set{}
		coords.Initialize(cfg)
		coords.Calculate(action, &setProgress)

		setProgress = 100.0 // stop progress output

		// output image
		img := m.CreatePicture(coords, ramp, cfg.XRes, cfg.YRes, setColor)
		m.OutputToJPG(img, cfg.ImageFile)

		took := time.Since(start).Seconds()
		totalTime += took
		if verbose {
			fmt.Printf("\n Took %0.1f seconds.\n\n", took)
		}

	}

	// overall stats
	if verbose {
		fmt.Println("-------------------------")
		fmt.Printf("Total time: %0.1f seconds.\n", totalTime)
		fmt.Printf("Average %0.1f seconds per frame.\n", totalTime/float64(totalFrames))
	}
}

func showProgress(progress *float64) {
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		for *progress < 100.0 {
			select {
			case <-ticker.C:
				// the ansi escape code here moves the cursor left 100 characters
				fmt.Printf("\u001b[100D %0.1f%% complete.", *progress*100)
			default:
			}
		}
		ticker.Stop()
	}()
}

// could just be done with a formula, probably
func totalFrames(zoomFactor, finalZoomLevel float64) (frames int) {
	for curZoom := 4.0; curZoom >= finalZoomLevel; frames++ {
		curZoom = 4.0 / math.Pow(zoomFactor, float64(frames))
	}
	return
}
