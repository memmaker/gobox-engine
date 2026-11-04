package present

import (
	"gobox/front"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// A key or click during the title picture's hold or fade ends it and is
// swallowed; once the menu is in, input goes through. Without a picture
// there is nothing to skip.
func TestMainSkip(t *testing.T) {
	f := filepath.Join(t.TempDir(), "title.png")
	w, _ := os.Create(f)
	png.Encode(w, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	w.Close()
	bd := &front.Picture{File: f}
	p := &Presenter{main: true}
	if p.mm.skip(false, nil) {
		t.Fatal("no picture: no phase to skip")
	}
	if p.mm.fade() >= 1 {
		t.Fatal("fresh menu should be fading")
	}
	if !p.mm.skip(false, bd) || p.mm.fade() < 1 {
		t.Fatal("a key during the fade should be taken and end it")
	}
	if p.mm.skip(false, bd) {
		t.Fatal("a second key must not be swallowed")
	}
	p.mm.at = time.Now()
	if !p.mm.skip(false, bd) || p.mm.fade() < 1 {
		t.Fatal("a click during the hold should end it and be taken")
	}
}
