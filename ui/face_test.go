package ui

import (
	"image"
	"image/color"
	"testing"
)

// TestFaceCentredSquare: a card's face is the head strip's middle square at
// one scale; a stray lit pixel at the strip's edge (a staff's tip) no longer
// widens the crop and shrinks the face.
func TestFaceCentredSquare(t *testing.T) {
	head := image.NewRGBA(image.Rect(0, 0, 88, 40))
	for y := 0; y < 40; y++ {
		for x := 24; x < 64; x++ {
			head.Set(x, y, color.RGBA{200, 80, 80, 255})
		}
	}
	head.Set(1, 30, color.RGBA{255, 0, 255, 255}) // the stray pixel
	p := &Presenter{faces: map[image.Image]image.Image{}}
	f := p.face(head)
	n := portraitSize
	for _, at := range []image.Point{{0, 0}, {n - 1, n - 1}, {n / 2, n / 2}} {
		if r, _, _, _ := f.At(at.X, at.Y).RGBA(); r>>8 != 200 {
			t.Errorf("face at %v: %v, want the face filling the square", at, f.At(at.X, at.Y))
		}
	}
}
