package main

import (
	"flag"
	"fmt"
	m "mandelbrot"
	"mandelbrot/cmd"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {

	// const startWidth = 4.0
	// const zoomFactor = 1.05 // how 'fast' the zoom happens. must be >1.
	// const iterFactor = 0.02 // double the number of iterations every iterFactor^-1 frames
	var startWidth, zoomFactor, iterFactor float64
	flag.Float64Var(&startWidth, "width", 4.0, "Starting width of the plot.")
	flag.Float64Var(&zoomFactor, "zoom", 1.1, "How 'fast' the zoom happens. 1.05 or 1.1 is generally appropriate. Must be >1.")
	flag.Float64Var(&iterFactor, "iter", 0.02, "Rate of increase of the iterations. Equals 1/<DoubleEveryNFrames> (eg 0.02 = 1/25).")
	showInfo := flag.Bool("info", false, "When set, display info only and do no computation.")
	cfg, verbose := cmd.Startup()

	// params for image generation and saving
	path := makeOutputDir(cfg.ImageFile)
	ramp := m.MakeRamp(m.ReadStops(cfg.RampFile))
	setColor := m.HexToRGBA(cfg.SetColor)

	totalFrames := totalFrames(zoomFactor, cfg.PlotWidth)
	totalTime := 0.0

	if *showInfo {
		fmt.Printf("Config info:\n------------\n%s\n----------\n", cfg)
		fmt.Printf("Start width:\t%0.5e\nZoom factor:\t%0.2f\nIter. factor:\t%0.2f\n", startWidth, zoomFactor, iterFactor)
		fmt.Printf("%d frames will be created in '%s'.\n", totalFrames, path)
		return
	}

	// alter plot_width,plot_height, iterations, image_file in order,
	// producing a series of images which 'zoom' into the configured point.
	// do this non-concurrently to save CPU for mandelbrot calcs and to prevent
	// excessive use of memory (keeping all the Sets in memory).

	origPlotWidth := cfg.PlotWidth
	origIterations := cfg.Iterations
	for i := 0; cfg.PlotWidth >= origPlotWidth; i++ {
		start := time.Now()

		// setup parameters for this frame
		cfg.PlotWidth = startWidth * math.Pow(zoomFactor, float64(-i))
		cfg.PlotHeight = cfg.PlotWidth * (float64(cfg.YRes) / float64(cfg.XRes))
		cfg.Iterations = origIterations * 1 << uint(float64(i)*iterFactor)
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

// displays textual progress every 500ms
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
