package ega

import (
	"image/color"
	"testing"
)

func TestPalette(t *testing.T) {
	if RGBA(15) != (color.RGBA{255, 255, 255, 255}) || RGBA(6).G != 82 || AmigaRGBA(AmigaBase).A != 255 {
		t.Fatal("palette")
	}
}
