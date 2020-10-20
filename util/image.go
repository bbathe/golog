package util

import (
	"image"
	"image/color"
)

func GenerateStatusImage(clr color.Color) image.Image {
	var w, h int = 8, 16

	// create image with a filled rectangle
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			img.Set(x, y, clr)
		}
	}

	return img
}
