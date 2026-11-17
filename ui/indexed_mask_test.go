package ui

import (
	"image/color"
	"testing"
)

// TestPalettedMask: a VGA picture's mask index is left out although its palette has an opaque colour there (BR1's combat
// icons drew their key colour as a black box). Detects the class: any paletted path that drops Picture.Mask.
func TestPalettedMask(t *testing.T) {
	pal := make([]color.RGBA, 256)
	for i := range pal {
		pal[i] = color.RGBA{uint8(i), 0, 0, 255}
	}
	m := uint8(16)
	img := Picture{Pix: []uint8{16, 96}, PixW: 2, Palette: pal, Mask: &m}.Converted(nil).Img
	if _, _, _, a := img.At(0, 0).RGBA(); a != 0 {
		t.Errorf("mask pixel alpha %d, want 0", a)
	}
	if got := img.At(1, 0).(color.RGBA); got != pal[96] {
		t.Errorf("colour 96 is %v, want %v", got, pal[96])
	}
}
