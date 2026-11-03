package ui

import (
	"image"
	"testing"
)

func TestAreaGeom(t *testing.T) {
	win, m := image.Rect(0, 0, 640, 400), image.Rect(0, 0, 320, 320)
	k, n, x, y := AreaGeom(win, m, 1, 8, image.Pt(0, 0))
	if k != 2 || n != 16 || x != 0 || y != 0 {
		t.Fatal(k, n, x, y)
	}
	_, _, x, y = AreaGeom(win, m, 1, 8, image.Pt(39, 39)) // far corner: kept in the map
	if x != 0 || y != 240 {
		t.Fatal(x, y)
	}
}
