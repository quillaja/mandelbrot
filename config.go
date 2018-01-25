package mandelbrot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Config is configuration info for the program, loaded from file.
type Config struct {
	CenterReal float64 `json:"center_real"`
	CenterImag float64 `json:"center_imag"`
	PlotWidth  float64 `json:"plot_width"`
	PlotHeight float64 `json:"plot_height"`
	XRes       int     `json:"x_res"`
	YRes       int     `json:"y_res"`
	Iterations int     `json:"iterations"`
	RampFile   string  `json:"ramp_file"`
	DataFile   string  `json:"data_file"`
	ImageFile  string  `json:"image_file"`
	SetColor   string  `json:"set_color"`
	JuliaReal  float64 `json:"julia_real"`
	JuliaImag  float64 `json:"julia_imag"`
}

// DoJulia is a convenince function to determine if the program should
// make a Julia set or not. If Julia[Real/Imag] is 0.0,0.0, then
// this returns false.
func (c Config) DoJulia() bool {
	return c.JuliaReal != 0.0 && c.JuliaImag != 0.0
}

// GetJulia is a convenience function to get the Julia point as a complex128.
func (c Config) GetJulia() complex128 {
	return complex(c.JuliaReal, c.JuliaImag)
}

func (c Config) String() string {
	f := "Plot center:\t%0.5f, %0.5f\nPlot W, H:\t%0.5f, %0.5f\nImage size:\t%dx%d\nIterations:\t%d\nJulia c =\t%0.5f + %0.5fi\nRamp file:\t%s\nData file:\t%s\nImage file:\t%s"
	return fmt.Sprintf(f, c.CenterReal, c.CenterImag, c.PlotWidth, c.PlotHeight, c.XRes, c.YRes, c.Iterations, c.JuliaReal, c.JuliaImag, c.RampFile, c.DataFile, c.ImageFile)
}

// NewConfig gets a Config with reasonable default values.
func NewConfig() Config {
	return Config{
		0.0, 0.0,
		4.0, 4.0,
		1000, 1000,
		512,
		"ramp.json",
		"default.gob",
		"output.jpg",
		"000000",
		0.0, 0.0}
}

// WriteConfig saves a config to file.
func WriteConfig(c Config, filename string) {

	data, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(filename, data, 664)
	if err != nil {
		panic(err)
	}
}

// ReadConfig loads a config file.
func ReadConfig(filename string) (c Config) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &c)
	if err != nil {
		panic(err)
	}
	return
}

// WriteDefault is WriteConfig(NewConfig(), "default.json").
func WriteDefault() {
	WriteConfig(NewConfig(), "default.json")
}
