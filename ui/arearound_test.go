package ui

import (
	"image"
	"testing"
)

// The preview burst is the square of radius r round the square, clipped to the map.
func TestAreaRound(t *testing.T) {
	if n := len(areaRound(image.Pt(10, 10), 2)); n != 25 {
		t.Errorf("radius 2: %d squares, want 25", n)
	}
	if n := len(areaRound(image.Pt(0, 0), 1)); n != 4 {
		t.Errorf("corner radius 1: %d squares, want 4", n)
	}
}
