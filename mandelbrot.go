package mandelbrot

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"mandelbrot/big"
	"mandelbrot/gob"
	stdbig "math/big"
	"math/cmplx"
	"os"
	"runtime"
	"sync"
)

const precision = 1024

// Action is a function which takes a complex number and does iterations
// to determine if the point is in a set (eg Mandelbrot set) or not. It also
// returns the number of iterations.
type Action func(Job)

// type BigAction func(big.Complex) (bool, int)

// Set is a list of *Job
type Set []Job

type Job interface {
	// SetN(complex128)
	// PerformAction(Action)
	RunMandelbrot(int)
	GetImageInfo() (bool, int, int, int)
}

// Job contains information about a specific point in the mandelbrot set.
type C128Job struct {
	N          complex128 // the complex number in question
	In         bool       // rough classification
	Iterations int        // number of iterations before becoming unbound
	Index      int        // for indexing/sorting in slice
	X, Y       int        // for making jpgs
}

// func (j C128Job) SetN(c complex128) {
// 	j.N = c
// }

func NewC128Job(n complex128, index, x, y int) *C128Job {
	return &C128Job{
		N:          n,
		In:         false,
		Iterations: 0,
		Index:      index,
		X:          x,
		Y:          y}
}

// func (j C128Job) PerformAction(action Action) {
// 	action(j)
// }

func (j *C128Job) RunMandelbrot(iterations int) {
	j.In, j.Iterations = IsMemberMandelbrot(j.N, iterations)
}

func (j *C128Job) GetImageInfo() (bool, int, int, int) {
	return j.In, j.Iterations, j.X, j.Y
}

type BigJob struct {
	N          *big.Complex
	In         bool
	Iterations int
	Index      int
	X, Y       int
}

// func (j BigJob) SetN(c complex128) {
// 	j.N = big.NewComplex(real(c), imag(c), precision)
// }

func NewBigJob(n complex128, index, x, y int) *BigJob {
	return &BigJob{
		N:          big.NewComplex(real(n), imag(n), precision),
		In:         false,
		Iterations: 0,
		Index:      index,
		X:          x,
		Y:          y}
}

// func (j BigJob) PerformAction(action Action) {
// 	action(j)
// }

// This version runs about 30% faster than my "V1" version below.
// 3.3 mins vs 2.1 mins
// Math from
// https://randomascii.wordpress.com/2011/08/13/faster-fractals-through-algebra/
func (j *BigJob) RunMandelbrot(iterations int) {
	z := new(big.Complex).Copy(j.N)
	for i := 0; i < iterations; i++ {
		realsq := new(stdbig.Float).Mul(&z.R, &z.R)
		imagsq := new(stdbig.Float).Mul(&z.I, &z.I)
		rsq, _ := realsq.Float64()
		isq, _ := imagsq.Float64()
		if rsq+isq > 4.0 {
			j.In = false
			j.Iterations = i
			return
		}
		z.I.Mul(&z.R, &z.I)
		z.I.Add(&z.I, &z.I)
		z.I.Add(&z.I, &j.N.I)
		z.R.Add(new(stdbig.Float).Sub(realsq, imagsq), &j.N.R)
	}

	j.In = true
	j.Iterations = iterations
}

func (j *BigJob) RunMandelbrotV1(iterations int) {

	z := new(big.Complex).Copy(j.N)
	for i := 0; i < iterations; i++ {
		left := new(big.Complex).Pow2(z)
		z.Add(left, j.N)
		abs, _ := z.AbsSq().Float64()
		if abs > 4.0 {
			// went to infinity
			j.In = false
			j.Iterations = i
			return
		}
	}

	// did not go to infinity
	j.In = true
	j.Iterations = iterations
}

func (j *BigJob) GetImageInfo() (bool, int, int, int) {
	return j.In, j.Iterations, j.X, j.Y
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
			// var j Job
			// if useBig {
			// 	j = NewBigJob(complex(x, y), i, w, h)
			// } else {
			// 	j = NewC128Job(complex(x, y), i, w, h)
			// }

			*coords = append(*coords, NewC128Job(complex(x, y), i, w, h))
			i++
		}

	}
}

func (coords *Set) InitializeBig(cfg Config) {
	halfwidth := new(stdbig.Float).SetPrec(precision).SetFloat64(cfg.PlotWidth)
	halfwidth.Quo(halfwidth, stdbig.NewFloat(2))
	halfheight := new(stdbig.Float).SetPrec(precision).SetFloat64(cfg.PlotHeight)
	halfheight.Quo(halfheight, stdbig.NewFloat(2))
	centerReal := new(stdbig.Float).SetPrec(precision)
	centerReal.Parse("-1.77810334274064037110522326038852639499207961414628307584575173232969154440", 0)
	centerImag := new(stdbig.Float).SetPrec(precision)
	centerImag.Parse("0.00767394242121339392672671947893471774958985018535019684946671264012302378", 0)

	left := new(stdbig.Float).Sub(centerReal, halfwidth)
	right := new(stdbig.Float).Add(centerReal, halfwidth)
	top := new(stdbig.Float).Add(centerImag, halfheight)
	bottom := new(stdbig.Float).Sub(centerImag, halfheight)

	yStep := new(stdbig.Float).Sub(top, bottom)
	yStep.Quo(yStep, stdbig.NewFloat(float64(cfg.YRes)))
	xStep := new(stdbig.Float).Sub(right, left)
	xStep.Quo(xStep, stdbig.NewFloat(float64(cfg.XRes)))

	for i, h, y := 0, 0, new(stdbig.Float).Copy(top); h < cfg.YRes; h++ {
		for w, x := 0, new(stdbig.Float).Copy(left); w < cfg.XRes; w++ {

			j := BigJob{}
			j.In = false
			j.Iterations = 0
			j.Index = i
			j.X = w
			j.Y = h
			j.N = new(big.Complex)
			j.N.R.Copy(x)
			j.N.I.Copy(y)

			*coords = append(*coords, &j)

			i++
			x.Add(x, xStep)
		}
		y.Sub(y, yStep)
	}

}

// CalculateProgress performs `action` on all the coordinates in a Set.
// The progress can be obtained by providing the address of a float64 in which
// [0,1] will be written.
func (coords Set) CalculateProgress(iterations int, progress *float64) {
	// concurrent implementation of actually computing mandelbrot set
	//
	// buffered input channel to hold values, 1 for each worker so none have
	// to block while waiting for jobs
	workers := runtime.NumCPU()
	in := make(chan Job, workers)

	// start workers
	wg := sync.WaitGroup{}
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func(id int) {
			defer wg.Done()
			for j := range in {
				j.RunMandelbrot(iterations)
			}
			// fmt.Printf("worker %d stopped.\n", id)
		}(w)
	}

	// send jobs to workers
	total := float64(len(coords))
	for i, j := range coords {
		in <- j // will block when buffered channel is full
		if progress != nil {
			*progress = float64(i+1) / total
		}
	}

	close(in) // close channel to stop workers
	wg.Wait() // wait for all workers to finish (join)
}

// Calculate performs `action` on all the coordinates in a Set.
func (coords Set) Calculate(iterations int) {
	coords.CalculateProgress(iterations, nil)
}

// WriteData writes a Set to filename as a gob (go object serialization).
func WriteData(coords Set, filename string) {
	gob.Register(&BigJob{})
	gob.Register(&C128Job{})
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
	gob.Register(&BigJob{})
	gob.Register(&C128Job{})
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
	in := make(chan Job, workers)
	wg := sync.WaitGroup{}
	wg.Add(workers)

	// start workers
	for w := 0; w < workers; w++ {
		go func(id int) {
			for c := range in {

				isIn, iterations, x, y := c.GetImageInfo()

				if isIn {
					img.SetRGBA(x, y, setColor)
				} else {
					img.SetRGBA(x, y, ramp[iterations%len(ramp)])
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

// OutputToJPG writes an image.Image to the given output filename
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
