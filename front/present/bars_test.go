package present

import (
	"image"
	"testing"
)

// The embedded bar art holds every piece the cards cut from it.
func TestBarArt(t *testing.T) {
	b := barArt().Bounds()
	for _, r := range []image.Rectangle{barLeft, barMid, barRight, greenFill, redFill, goldFill} {
		if !r.In(b) {
			t.Fatalf("%v outside the bar art %v", r, b)
		}
	}
}
