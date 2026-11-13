package asset

import (
	"image/color"
	"testing"
)

func TestPalette(t *testing.T) {
	if EGARGBA(15) != (color.RGBA{255, 255, 255, 255}) || EGARGBA(6).G != 82 || AmigaRGBA(AmigaBase).A != 255 {
		t.Fatal("palette")
	}
}

func TestColourNames(t *testing.T) {
	if EGABlack != 0 || EGARed != 4 || EGAGreen != 2 || EGALightRed != 12 || EGAYellow != 14 || EGAWhite != 15 {
		t.Fatal("colour numbers")
	}
}
