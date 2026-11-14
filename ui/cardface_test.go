package ui

import (
	"image"
	"image/color"
	"testing"
)

// TestCardFaceWholePixels: a pool head (88x40) shows on its card at one
// card pixel per head pixel, its middle 48 columns on the bottom edge, never
// stretched by a fraction (40 to 48 doubled every fifth row and column).
func TestCardFaceWholePixels(t *testing.T) {
	const w, h = 88, 40
	head := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			head.SetRGBA(x, y, color.RGBA{uint8(x), uint8(y), 1, 255}) // each pixel its own
		}
	}
	p := &Presenter{faces: map[image.Image]image.Image{}}
	f := p.face(head)
	x0, y0 := (w-portraitSize)/2, portraitSize-h // the head pixel at the face's (0, 0) column, the row the head starts
	for y := y0; y < portraitSize; y++ {
		for x := range portraitSize {
			if got, want := color.RGBAModel.Convert(f.At(x, y)), head.At(x0+x, y-y0); got != want {
				t.Fatalf("face (%d, %d) is %v, want head (%d, %d) %v", x, y, got, x0+x, y-y0, want)
			}
		}
	}
}
