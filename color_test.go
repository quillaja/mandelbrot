package mbrot

import (
	"image/color"
	"reflect"
	"testing"
)

func TestMakeRamp(t *testing.T) {
	type args struct {
		stops []Stop
		// maxIteration int
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
	}{
		{
			name: "16 b-w greyscale",
			args: args{
				[]Stop{
					Stop{0, "000000"},
					Stop{16, "FFFFFF"}},
				// 16},
			},
			wantLen: 17},
		{
			name: "32 w-b-w greyscale",
			args: args{
				[]Stop{
					Stop{0, "FFFFFF"},
					Stop{16, "000000"},
					Stop{32, "FFFFFF"}},
				// 32},
			},
			wantLen: 33},
		{
			name: "128 r-g-b",
			args: args{
				[]Stop{
					Stop{0, "FF0000"},
					Stop{64, "00FF00"},
					Stop{128, "0000FF"}},
				// 128},
			},
			wantLen: 129},
		{
			name: "r-b 32 color",
			args: args{
				[]Stop{
					Stop{0, "FF0000"},
					Stop{32, "0000FF"}},
				// 64},
			},
			wantLen: 33},
		{
			name: "2 b-w",
			args: args{
				[]Stop{
					Stop{0, "000000"},
					Stop{1, "FFFFFF"}},
				// 20},
			},
			wantLen: 2},
		{
			name: "256 grayscale w-b-w",
			args: args{
				[]Stop{
					Stop{0, "FFFFFF"},
					Stop{128, "000000"},
					Stop{256, "FFFFFF"}},
				// 20},
			},
			wantLen: 257},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRamp := MakeRamp(tt.args.stops /*, tt.args.maxIteration*/)
			t.Log(gotRamp)

			// test length
			if len(gotRamp) != tt.wantLen {
				t.Errorf("MakeRamp() = len(%v), want len(%v)", len(gotRamp), tt.wantLen)
			}

			// test that stops are in the correct positions (indices) in the ramp
			for _, s := range tt.args.stops {
				if s.RGBA() != gotRamp[s.Position] {
					t.Errorf("MakeRamp(): stop %v != gotRamp[%d] (%v)", s, s.Position, gotRamp[s.Position])
				}
			}
		})
	}
}

func TestHexToRGBA(t *testing.T) {
	type args struct {
		hexColor string
	}
	tests := []struct {
		name string
		args args
		want color.RGBA
	}{
		{"black", args{"000000"}, color.RGBA{0, 0, 0, 255}},
		{"white", args{"FFFFFF"}, color.RGBA{255, 255, 255, 255}},
		{"blue", args{"0000FF"}, color.RGBA{0, 0, 255, 255}},
		{"50% gray", args{"808080"}, color.RGBA{128, 128, 128, 255}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HexToRGBA(tt.args.hexColor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HexToRGBA() = %v, want %v", got, tt.want)
			}
		})
	}
}
