package gpu

import "testing"

// TestSizedRemakesGone: a canvas deallocated by another holder is made
// again by Sized, and writing to a deallocated image does nothing.
func TestSizedRemakesGone(t *testing.T) {
	var img *Image
	Sized(&img, 8, 8)
	old := img
	old.Deallocate()
	old.WritePixels(make([]byte, 8*8*4)) // no panic
	if !Sized(&img, 8, 8) || img == old {
		t.Fatal("a deallocated canvas is kept")
	}
	img.WritePixels(make([]byte, 8*8*4))
	img.Deallocate()
}
