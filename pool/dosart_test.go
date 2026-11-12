package pool

import (
	"os"
	"testing"
)

// TestDOSArt: an Amiga picture carries its DOS twin, and DOSArt (F12)
// shows the twin.
func TestDOSArt(t *testing.T) {
	if _, err := os.Stat("../../por-go/data/amiga/PIC.DAX"); err != nil {
		t.Skip("no Amiga import")
	}
	e := testEngine(t)
	e.LoadAnimPic("PIC", 1, false)
	if len(e.pic.frames) == 0 || e.pic.frames[0].Image.Alt == nil {
		t.Fatal("PIC 1 has no DOS twin")
	}
	im := e.pic.frames[0].Image
	if e.art(im, 0) != im {
		t.Fatal("the Amiga art shows the twin")
	}
	DOSArt.Store(true)
	t.Cleanup(func() { DOSArt.Store(false) })
	if e.art(im, 0) != im.Alt {
		t.Fatal("DOSArt does not show the twin")
	}
}
