package mandelbrot

import (
	"encoding/hex"
	"encoding/json"
	"image/color"
	"io/ioutil"
)

// DEPRECATED
// MakeColorRamp produces a list of colors by interpolating
// color values between a series of stops.
//
// `stops` must contain at least 2 colors, and `numColors` must be greater
// than the number of stops.
// func MakeColorRamp(stops []color.RGBA, numColors int) []color.RGBA {
// 	if len(stops) < 2 {
// 		panic("Invalid arguments. len(stops) must be >= 2.")
// 	}
// 	if len(stops) > numColors {
// 		panic("Invalid arguments. More stops than number of desired colors.")
// 	}
// 	if len(stops) == numColors {
// 		return stops // why did the idiot even call the func?
// 	}

// 	ramp := []color.RGBA{}

// 	// num colors to add between each stop, and any remainders to distribute
// 	betweenStops := (numColors - len(stops)) / (len(stops) - 1)
// 	remainder := (numColors - len(stops)) % (len(stops) - 1)

// 	// go through pairs of stops, and interpolate the ramp in between
// 	for i := 0; i < len(stops)-1; i++ {
// 		cur, next := stops[i], stops[i+1]

// 		// add current stop to list
// 		ramp = append(ramp, cur)

// 		// create interpolated colors bewtween current and next stop.
// 		// first, distribute some of the remainder, if any
// 		extra := 0
// 		if remainder > 0 {
// 			extra = 1
// 			remainder--
// 		}

// 		// calculate delta for each RGB component
// 		dR := (int(next.R) - int(cur.R)) / (1 + betweenStops + extra)
// 		dG := (int(next.G) - int(cur.G)) / (1 + betweenStops + extra)
// 		dB := (int(next.B) - int(cur.B)) / (1 + betweenStops + extra)

// 		// do interpolation
// 		for j := 1; j <= (betweenStops + extra); j++ {
// 			c := color.RGBA{
// 				R: uint8(int(cur.R) + (dR * j)),
// 				G: uint8(int(cur.G) + (dG * j)),
// 				B: uint8(int(cur.B) + (dB * j)),
// 				A: 255}
// 			ramp = append(ramp, c)
// 		}

// 		// if next is the last stop, it has to be added
// 		if i == len(stops)-2 {
// 			ramp = append(ramp, next)
// 		}
// 	}

// 	return ramp
// }

// Stop represents a color stop and its position within a color ramp.
type Stop struct {
	Position int    `json:"position,omitempty"`
	Color    string `json:"color,omitempty"`
}

// RGBA converts the `Stop` to `color.RGBA`.
func (s Stop) RGBA() color.RGBA {
	return HexToRGBA(s.Color)
}

// HexToRGBA converts a hex string in the form "RRGGBB" to a color.RGBA.
// The alpha component is always 255 (opaque).
func HexToRGBA(hexColor string) color.RGBA {
	components, err := hex.DecodeString(hexColor)
	if err != nil {
		panic(err)
	}
	return color.RGBA{
		R: components[0],
		G: components[1],
		B: components[2],
		A: 255}
}

// ReadStops reads and returns the list of `Stop` from the json file.
func ReadStops(filename string) []Stop {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var stops []Stop
	err = json.Unmarshal(data, &stops)
	if err != nil {
		panic(err)
	}

	return stops
}

// MakeRamp uses a list of `Stop` to create a color ramp.
func MakeRamp(stops []Stop /*maxIteration int*/) (ramp []color.RGBA) {
	for is := 0; is < len(stops)-1; is++ {
		cur, next := stops[is], stops[is+1]

		// calculate various parameters
		between := (next.Position - 1) - cur.Position
		nc, cc := next.RGBA(), cur.RGBA()
		dR := round((float64(nc.R) - float64(cc.R)) / float64(between+1))
		dG := round((float64(nc.G) - float64(cc.G)) / float64(between+1))
		dB := round((float64(nc.B) - float64(cc.B)) / float64(between+1))

		// do interpolation between stops
		for i := 0; i <= between; i++ {
			ramp = append(ramp,
				color.RGBA{
					R: uint8(int(cc.R) + i*dR),
					G: uint8(int(cc.G) + i*dG),
					B: uint8(int(cc.B) + i*dB),
					A: 255})
		}

		// add final stop color to finish ramp
		if is == len(stops)-2 {
			ramp = append(ramp, nc)
		}
	}

	// if maxIteration is greater than the length of the ramp, the ramp
	// must be repeated.
	// var origRamp []color.RGBA
	// origRamp = append(origRamp, ramp...)
	// for repeat := maxIteration / len(ramp); repeat > 0; repeat-- {
	// 	ramp = append(ramp, origRamp...)
	// }

	return
}

// utility function to round floats to ints, since golang is so
// omniscient to realize that we don't need this crap in the std libary
func round(val float64) int {
	if val < 0 {
		return int(val - 0.5)
	}
	return int(val + 0.5)
}
