package main

import (
	"flag"
	"fmt"
	mbrot "mandelbrot"
	"runtime"
	"sync"
	"time"
)

func main() {

	var configFile string
	var writeDefault bool
	var verbose bool
	var left, right, top, bottom float64

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

	left, right = cfg.CenterReal-(cfg.PlotWidth/2), cfg.CenterReal+(cfg.PlotWidth/2)
	top, bottom = cfg.CenterImag+(cfg.PlotHeight/2), cfg.CenterImag-(cfg.PlotHeight/2)

	// print configuration settings
	if verbose {
		fmt.Println(cfg)
		fmt.Println("Processing...")
	}

	start := time.Now() // to show processing time when finished
	coords := mbrot.MandelSet{}
	yStep := (top - bottom) / float64(cfg.YRes)
	xStep := (right - left) / float64(cfg.XRes)

	for i, h, y := 0, 0, top; h < cfg.YRes; h, y = h+1, y-yStep {
		for w, x := 0, left; w < cfg.XRes; w, x = w+1, x+xStep {
			coords = append(coords, &mbrot.MandelJob{
				N:          complex(x, y),
				In:         false,
				Iterations: 0,
				Index:      i,
				X:          w,
				Y:          h})
			i++
		}

	}

	// concurrent implementation of actually computing mandelbrot set
	//
	// buffered input channel to hold values, 1 for each worker so none have
	// to block while waiting for jobs
	workers := runtime.NumCPU()
	in := make(chan *mbrot.MandelJob, workers)

	// choose mandelbrot set or julia set
	var action func(complex128) (bool, int)
	if cfg.DoJulia() {
		action = func(n complex128) (bool, int) {
			return mbrot.IsMemberJulia(n, cfg.GetJulia(), cfg.Iterations)
		}
	} else {
		action = func(n complex128) (bool, int) {
			return mbrot.IsMemberMandelbrot(n, cfg.Iterations)
		}
	}

	// start workers
	wg := sync.WaitGroup{}
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func(id int) {
			defer wg.Done()
			for j := range in {
				j.In, j.Iterations = action(j.N)
			}
			// fmt.Printf("worker %d stopped.\n", id)
		}(w)
	}

	// send jobs to workers
	for _, j := range coords {
		in <- j // will block when buffered channel is full
	}

	close(in) // close channel to stop workers

	// wait for all workers to finish (join)
	wg.Wait()

	// output data
	if verbose {
		fmt.Printf("Writing data to %s.\n", cfg.DataFile)
	}

	mbrot.WriteData(coords, cfg.DataFile)

	if verbose {
		fmt.Println("Took", time.Since(start).Seconds(), "seconds")
	}

}
