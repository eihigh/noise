package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/eihigh/noise"
)

func main() {
	img := image.NewGray(image.Rect(0, 0, 256, 256))
	period := 50.0
	f := noise.Fractal2(noise.Perlin2(88), 6, 0.5, 2, noise.Ridge)
	var vMin = 1.0
	var vMax = 0.0

	for y := range 256 {
		for x := range 256 {
			v := f(float64(x)/period, float64(y)/period)
			v = noise.Bias(0.1)(noise.SmoothStep(0, 0.95)(v))
			vMin = min(vMin, v)
			vMax = max(vMax, v)
			img.SetGray(x, y, color.Gray{uint8(v * 255)})
		}
	}
	fmt.Println("vMin:", vMin, "vMax:", vMax)

	w, err := os.Create("perlin_noise.png")
	if err != nil {
		panic(err)
	}
	defer w.Close()
	if err = png.Encode(w, img); err != nil {
		panic(err)
	}
}
