package main

import (
	"fmt"
	m "mandelbrot"
	"mandelbrot/cmd"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TODO:
// Done - (see makeOutputDir) parameterize the output directory,
// Done - (see makeOutputDir) create the directory if it doesn't exist,
// parameterize zoom factor and iterFactor (etc)
// parameterize starting 'zoom' (width of 4.0)
// ?? stop after N frames instead of going until reaching Config.PlotWidth?
// Done - aspect ratio ... used yres/xres to alter plotheight

func main() {
	cfg, verbose := cmd.Startup()

	path := makeOutputDir(cfg.ImageFile)

	// alter plot_width,plot_height, iterations, image_file in order,
	// producing a series of images which 'zoom' into the configured point.
	// do this non-concurrently to save CPU for mandelbrot calcs and to prevent
	// excessive use of memory (keeping all the Sets in memory).

	ramp := m.MakeRamp(m.ReadStops(cfg.RampFile))
	setColor := m.HexToRGBA(cfg.SetColor)

	const startWidth = 4.0
	const zoomFactor = 1.1
	const iterFactor = 25

	totalFrames := totalFrames(zoomFactor, cfg.PlotWidth)
	totalTime := 0.0

	origPlotWidth := cfg.PlotWidth
	origIterations := cfg.Iterations
	for i := 0; cfg.PlotWidth >= origPlotWidth; i++ {
		start := time.Now()

		// setup parameters for this frame
		cfg.PlotWidth = startWidth / math.Pow(zoomFactor, float64(i))
		cfg.PlotHeight = cfg.PlotWidth * (float64(cfg.YRes) / float64(cfg.XRes))
		// double the number of iterations every iterFactor (25) frames
		cfg.Iterations = origIterations * 1 << uint(i/iterFactor)
		cfg.ImageFile = filepath.Join(path, fmt.Sprintf("%010d.jpg", i))
		action := func(n complex128) (bool, int) {
			return m.IsMemberMandelbrot(n, cfg.Iterations)
		}

		// show status
		var setProgress float64
		if verbose {
			fmt.Printf("Frame %d of %d\n", i+1, totalFrames)
			fmt.Printf(" Iterations: %d\n", cfg.Iterations)
			fmt.Printf(" Plot width: %0.8e\n", cfg.PlotWidth)
			showProgress(&setProgress)
		}

		// do work
		coords := m.Set{}
		coords.Initialize(cfg)
		coords.CalculateProgress(action, &setProgress)

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
		for *progress < 1.0 {
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
func totalFrames(zoomFactor, finalWidth float64) (frames int) {
	for curZoom := 4.0; curZoom >= finalWidth; frames++ {
		curZoom = 4.0 / math.Pow(zoomFactor, float64(frames))
	}
	return
}

// Create a directory for the program output from the filename
// provided (eg in Config.ImageFile). Creates directories if they don't exist.
func makeOutputDir(base string) string {
	dir, file := filepath.Split(base)
	file = strings.Split(file, ".")[0]
	path := filepath.Join(dir, file+"_zoom")
	err := os.MkdirAll(path, 0755)
	if err != nil {
		panic(err)
	}
	return path
}
