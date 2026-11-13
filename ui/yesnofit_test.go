package ui

import (
	"image"
	"testing"
)

// TestYesNoUnderPicture: a yes/no with a picture up sits under it.
func TestYesNoUnderPicture(t *testing.T) {
	b := image.Rect(0, 0, 640, 400)
	_, top := yesNoFit(b, 250, true, 2, "Go in?", []string{"Yes", "No"}, false)
	if top < 250 {
		t.Fatalf("top %d: over the picture (bottom 250)", top)
	}
	if _, top := yesNoFit(b, 0, true, 2, "Go in?", []string{"Yes", "No"}, false); top >= 250 || top <= 0 {
		t.Fatalf("no picture: top %d, want centred", top)
	}
}
