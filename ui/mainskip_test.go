package ui

import (
	"image"
	"testing"
	"time"
)

// A key or click during the title picture's hold or fade ends it and is
// swallowed; once the menu is in, input goes through. Without a picture
// there is nothing to skip.
func TestMainSkip(t *testing.T) {
	p := &Presenter{main: true}
	if p.mm.skip(false) {
		t.Fatal("no picture: no phase to skip")
	}
	p.mm.title = image.NewRGBA(image.Rect(0, 0, 1, 1))
	if p.mm.fade() >= 1 {
		t.Fatal("fresh menu should be fading")
	}
	if !p.mm.skip(false) || p.mm.fade() < 1 {
		t.Fatal("a key during the fade should be taken and end it")
	}
	if p.mm.skip(false) {
		t.Fatal("a second key must not be swallowed")
	}
	p.mm.at = time.Now()
	if !p.mm.skip(false) || p.mm.fade() < 1 {
		t.Fatal("a click during the hold should end it and be taken")
	}
}
