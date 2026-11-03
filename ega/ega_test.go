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

func TestColourNames(t *testing.T) {
	if Black != 0 || Red != 4 || Green != 2 || LightRed != 12 || Yellow != 14 || White != 15 {
		t.Fatal("colour numbers")
	}
}
