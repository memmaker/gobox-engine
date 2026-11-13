package ui

import (
	"image"
	"math"
)

// Scale is K as a float, for the GeoM scales.
func Scale(b image.Rectangle) float64 { return float64(K(b)) }

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
