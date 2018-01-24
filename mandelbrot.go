package mandelbrot

import (
	"encoding/gob"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math/cmplx"
	"os"
)

// func main() {

// 	var configFile string
// 	var writeDefault bool
// 	var verbose bool
// 	var left, right, top, bottom float64

// 	flag.StringVar(&configFile, "config", "", "The file with configuration data (in json format).")
// 	flag.BoolVar(&writeDefault, "default", false, "Set this flag to output a default config file to 'default.json'. The program will then exit.")
// 	flag.BoolVar(&verbose, "v", false, "Show verbose output when set.")
// 	flag.Parse()

// 	if writeDefault {
// 		WriteDefault()
// 		return
// 	}

// 	if configFile == "" {
// 		panic("Config file not specified.")
// 	}

// 	cfg := ReadConfig(configFile)
// 	// TODO: check config for errors

// 	left, right = cfg.CenterReal-(cfg.PlotWidth/2), cfg.CenterReal+(cfg.PlotWidth/2)
// 	top, bottom = cfg.CenterImag+(cfg.PlotHeight/2), cfg.CenterImag-(cfg.PlotHeight/2)

// 	// print configuration settings
// 	if verbose {
// 		fmt.Println(cfg)
// 		fmt.Println("Processing...")
// 	}

// 	start := time.Now() // to show processing time when finished
// 	coords := MandelSet{}
// 	yStep := (top - bottom) / float64(cfg.YRes)
// 	xStep := (right - left) / float64(cfg.XRes)

// 	for i, h, y := 0, 0, top; h < cfg.YRes; h, y = h+1, y-yStep {
// 		for w, x := 0, left; w < cfg.XRes; w, x = w+1, x+xStep {
// 			coords = append(coords, &MandelJob{
// 				N:          complex(x, y),
// 				In:         false,
// 				Iterations: 0,
// 				Index:      i,
// 				X:          w,
// 				Y:          h})
// 			i++
// 		}

// 	}

// 	// concurrent implementation of actually computing mandelbrot set
// 	//
// 	// buffered input channel to hold values, 1 for each worker so none have
// 	// to block while waiting for jobs
// 	workers := runtime.NumCPU()
// 	in := make(chan *MandelJob, workers)

// 	// choose mandelbrot set or julia set
// 	var action func(complex128) (bool, int)
// 	if cfg.DoJulia() {
// 		action = func(n complex128) (bool, int) {
// 			return IsMemberJulia(n, cfg.GetJulia(), cfg.Iterations)
// 		}
// 	} else {
// 		action = func(n complex128) (bool, int) {
// 			return IsMemberMandelbrot(n, cfg.Iterations)
// 		}
// 	}

// 	// start workers
// 	wg := sync.WaitGroup{}
// 	wg.Add(workers)
// 	for w := 0; w < workers; w++ {
// 		go func(id int) {
// 			defer wg.Done()
// 			for j := range in {
// 				j.In, j.Iterations = action(j.N)
// 			}
// 			// fmt.Printf("worker %d stopped.\n", id)
// 		}(w)
// 	}

// 	// send jobs to workers
// 	for _, j := range coords {
// 		in <- j // will block when buffered channel is full
// 	}

// 	close(in) // close channel to stop workers

// 	// wait for all workers to finish (join)
// 	wg.Wait()

// 	// output data
// 	if verbose {
// 		fmt.Printf("Writing data to %s.\n", cfg.DataFile)
// 	}
// 	WriteData(coords, cfg.DataFile)

// 	// output to jpg
// 	if verbose {
// 		fmt.Printf("Writing image to %s\n", cfg.ImageFile)
// 	}
// 	ramp := MakeRamp(ReadStops(cfg.RampFile))
// 	OutputToJPG(
// 		CreatePicture(coords, ramp, cfg.XRes, cfg.YRes, HexToRGBA(cfg.SetColor)),
// 		cfg.ImageFile)

// 	if verbose {
// 		fmt.Println("Took", time.Since(start).Seconds(), "seconds")
// 	}

// }

// DEPRECATED
// PrintToConsole displays the mandelbrot set as text on the console.
// func PrintToConsole(coords MandelSet) {
// 	for i := 0; i < len(coords); i++ {
// 		if coords[i].in {
// 			fmt.Print("*")
// 		} else {
// 			fmt.Print(" ")
// 		}

// 		if (i+1)%width == 0 {
// 			fmt.Print("\n")
// 		}
// 	}
// }

// MandelSet is a list of *MandelJob
type MandelSet []*MandelJob

// MandelJob contains information about a specific point in the mandelbrot set.
type MandelJob struct {
	N          complex128 // the complex number in question
	In         bool       // rough classification
	Iterations int        // number of iterations before becoming unbound
	Index      int        // for indexing/sorting in slice
	X, Y       int        // for making jpgs
}

// WriteData writes a MandelSet to filename as a gob (go object serialization).
func WriteData(coords MandelSet, filename string) {
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		panic(err)
	}

	enc := gob.NewEncoder(file)
	err = enc.Encode(coords)
	if err != nil {
		panic(err)
	}
}

// ReadData reads a MandelSet from filename.
func ReadData(filename string) (coords MandelSet) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		panic(err)
	}

	dec := gob.NewDecoder(file)
	err = dec.Decode(&coords)
	if err != nil {
		panic(err)
	}

	return
}

// IsMemberMandelbrot runs the recurrent formula for the Mandelbrot set
// for the given number of iterations, and returns if the complex number
// `c` is in the set or not, as well as how many iterations it took to
// become 'infinity'.
func IsMemberMandelbrot(c complex128, iterations int) (bool, int) {
	return IsMemberJulia(c, c, iterations)
	// for i, z := 0, c; i < iterations; i++ {
	// 	z = cmplx.Pow(z, 2) + c
	// 	if cmplx.Abs(z) > 2 {
	// 		// went to infinity
	// 		return false, i
	// 	}
	// }

	// // did not go to "infinity"
	// return true, iterations
}

// IsMemberJulia runs the recurrent formula for the Julia set for the given
// number of iterations, and returns if the complex number `z` is in the set
// or not, as well as how many iterations it took to become 'infinity'.
func IsMemberJulia(z complex128, c complex128, iterations int) (bool, int) {

	for i, z := 0, z; i < iterations; i++ {
		z = cmplx.Pow(z, 2) + c
		if cmplx.Abs(z) > 2 {
			// went to infinity
			return false, i
		}
	}

	// did not go to "infinity"
	return true, iterations
}

//CreatePicture draws an image.RGBA image.Image from the points created above.
func CreatePicture(coords MandelSet, ramp []color.RGBA, width, height int, setColor color.RGBA) image.Image {

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for _, c := range coords {
		if c.In {
			img.SetRGBA(c.X, c.Y, setColor)
		} else {
			img.SetRGBA(c.X, c.Y, ramp[c.Iterations%len(ramp)])
		}
	}
	return img
}

// OutputToJPG write an image.Image to the given output filename
func OutputToJPG(img image.Image, outputFilename string) {
	file, err := os.Create(outputFilename)
	defer file.Close()
	if err != nil {
		panic(fmt.Errorf("error when opening '%s'", outputFilename))
	}
	err = jpeg.Encode(file, img, &jpeg.Options{Quality: 98})
	if err != nil {
		panic(err)
	}
}
