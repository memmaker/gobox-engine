package engine

import (
	"strings"
	"testing"
)

// TestOutdoorPicture (BUG-CLASSES 4 recheck): a script's picture on the
// overland (In3D 0, no full-picture flag) is published, not left black.
func TestOutdoorPicture(t *testing.T) {
	e := testEngine(t)
	e.Explore = true
	e.DevAction = e.DevScene("overland")
	e.Init()
	e.Start(e.Play)
	if e.state().Area.In3D != 0 {
		t.Fatal("not outdoors")
	}
	e.es.picShown = 1 // as script.OpPicture with block 1, the Portrait 0xFF branch
	e.LoadAnimPic("PIC", 1, false)
	e.DrawAnimFirst()
	e.Advance()
	var got []Picture
	for _, p := range e.pagePictures() {
		if p.File != "" && p.Block == 1 && strings.HasPrefix(p.File, "PIC") {
			got = append(got, p)
		}
	}
	if len(got) != 1 {
		t.Fatalf("outdoor picture: %d layers of PIC block 1 (want 1); pictures %d", len(got), len(e.pagePictures()))
	}
}
