package ui

import (
	"image"
	"testing"
)

func TestScale(t *testing.T) {
	if k := Scale(image.Rect(0, 0, 1280, 800)); k != 3 {
		t.Fatal(k)
	}
	if k := Scale(image.Rect(0, 0, 100, 100)); k != 1 {
		t.Fatal(k)
	}
	if k := FitScale(3, 2, func(k float64) bool { return k <= 4 }); k != 3 {
		t.Fatal(k) // 6 too big; 3 fits
	}
	if k := FitScale(1, 1, func(k float64) bool { return k <= 0.5 }); k != 0.5 {
		t.Fatal(k)
	}
}
