package asset

import "testing"

// TestCutLogo: no cut set, or a picture without the lettering, is an error;
// a yellow pixel above Sub is cut to a 1x1 lettering.
func TestCutLogo(t *testing.T) {
	pix := make([]uint8, 320*200)
	lc := LogoCut{Sub: 108, Bottom: 160, Keep: []uint8{EGAYellow}, SubInk: EGALightMagenta}
	if _, err := CutLogo(pix, 320, LogoCut{}); err == nil {
		t.Error("the zero cut cut a logo")
	}
	if _, err := CutLogo(pix, 320, lc); err == nil {
		t.Error("an empty picture cut a logo")
	}
	pix[5*320+7] = EGAYellow
	m, err := CutLogo(pix, 320, lc)
	if err != nil || m.Bounds().Dx() != 1 || m.Bounds().Dy() != 1 {
		t.Errorf("cut %v, %v", m, err)
	}
}
