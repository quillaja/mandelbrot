package mandelbrot

import (
	"encoding/gob"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math/cmplx"
	"os"
	"runtime"
	"sync"
)

// DEPRECATED
// PrintToConsole displays the mandelbrot set as text on the console.
// func PrintToConsole(coords Set) {
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

// Action is a function which takes a complex number and does iterations
// to determine if the point is in a set (eg Mandelbrot set) or not. It also
// returns the number of iterations.
type Action func(complex128) (bool, int)

// Set is a list of *Job
type Set []*Job

// Job contains information about a specific point in the mandelbrot set.
type Job struct {
	N          complex128 // the complex number in question
	In         bool       // rough classification
	Iterations int        // number of iterations before becoming unbound
	Index      int        // for indexing/sorting in slice
	X, Y       int        // for making jpgs
}

// Initialize sets up a MandelSet according to the configuration specified.
func (coords *Set) Initialize(cfg Config) {
	left, right := cfg.CenterReal-(cfg.PlotWidth/2), cfg.CenterReal+(cfg.PlotWidth/2)
	top, bottom := cfg.CenterImag+(cfg.PlotHeight/2), cfg.CenterImag-(cfg.PlotHeight/2)
	yStep := (top - bottom) / float64(cfg.YRes)
	xStep := (right - left) / float64(cfg.XRes)

	// Initialize coords
	for i, h, y := 0, 0, top; h < cfg.YRes; h, y = h+1, y-yStep {
		for w, x := 0, left; w < cfg.XRes; w, x = w+1, x+xStep {
			*coords = append(*coords, &Job{
				N:          complex(x, y),
				In:         false,
				Iterations: 0,
				Index:      i,
				X:          w,
				Y:          h})
			i++
		}

	}
}

// Calculate performs `action` on all the coordinates in a MandelSet.
func (coords Set) Calculate(action Action, progress *float64) {
	// concurrent implementation of actually computing mandelbrot set
	//
	// buffered input channel to hold values, 1 for each worker so none have
	// to block while waiting for jobs

	workers := runtime.NumCPU()
	in := make(chan *Job, workers)

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
	total := float64(len(coords))
	for i, j := range coords {
		in <- j // will block when buffered channel is full
		*progress = float64(i) / total
	}

	close(in) // close channel to stop workers
	wg.Wait() // wait for all workers to finish (join)
}

// WriteData writes a MandelSet to filename as a gob (go object serialization).
func WriteData(coords Set, filename string) {
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

// ReadData reads a Set from filename.
func ReadData(filename string) (coords Set) {
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
		z = F(z, c, 2)
		if cmplx.Abs(z) > 2.0 {
			// went to infinity
			return false, i
		}
	}

	// did not go to "infinity"
	return true, iterations
}

// F is the general form of the recurrence function `F = z^exp + c` .
func F(z, c, exp complex128) complex128 {
	return cmplx.Pow(z, exp) + c
}

//CreatePicture draws an image.RGBA image.Image from the points created above.
func CreatePicture(coords Set, ramp []color.RGBA, width, height int, setColor color.RGBA) image.Image {

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	workers := runtime.NumCPU()
	in := make(chan *Job, workers)
	wg := sync.WaitGroup{}
	wg.Add(workers)

	// start workers
	for w := 0; w < workers; w++ {
		go func(id int) {
			for c := range in {
				if c.In {
					img.SetRGBA(c.X, c.Y, setColor)
				} else {
					img.SetRGBA(c.X, c.Y, ramp[c.Iterations%len(ramp)])
				}
			}
			wg.Done()

		}(w)
	}

	// send jobs to workers
	for _, c := range coords {
		in <- c
	}

	// join workers
	close(in)
	wg.Wait()

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
