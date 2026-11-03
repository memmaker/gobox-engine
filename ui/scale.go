package ui

import (
	"image"
	"math"
)

// Scale is K, the UI's one glyph cell size over 8: the largest whole
// factor a minW x minH page fits b at (at least 1). Every glyph and ribbon
// tile is 8*K window pixels; a larger window gives more cells, not bigger
// ones.
func Scale(b image.Rectangle, minW, minH int) float64 {
	k, _, _ := FullPageCells(b, minW, minH)
	return float64(k)
}

// FitScale is unit (K) times size s (a whole size step: 1x-4x) when
// ok takes it, else the largest smaller whole step of unit ok takes, else
// (a box too big for the window even at K) the largest whole scale under
// unit ok takes, its pixels square, else halves of 1 down to an eighth.
func FitScale(unit, s float64, ok func(k float64) bool) float64 {
	s = max(1, math.Round(s)) // a size saved before as a fraction: 1
	for n := s; n >= 1; n-- {
		if ok(unit * n) {
			return unit * n
		}
	}
	for k := math.Ceil(unit) - 1; k >= 1; k-- {
		if ok(k) {
			return k
		}
	}
	k := 1.0
	for k > 0.125 && !ok(k) { // a window too small for it even at 1: halves, as far as it fits
		k /= 2
	}
	return k
}
