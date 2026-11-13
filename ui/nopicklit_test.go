package ui

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
	"gobox/gpu"
)

// TestNoPickNoHighlight: a NoPick question's menu lights no row (no
// selection ground).
func TestNoPickNoHighlight(t *testing.T) {
	dst := gpu.NewImage(640, 400)
	defer dst.Deallocate()
	full := func(byte) [8]byte { return [8]byte{255, 255, 255, 255, 255, 255, 255, 255} }
	blue := func(noPick bool) (n int) {
		dst.Clear()
		p := askedOn(&play.Ask{Kind: game.Choice, NoPick: noPick})
		p.asking(func() { p.drawMenu(dst, full, "Take", []string{"Yes", "No"}, []byte{'Y', 'N'}, nil, 0, 1, 320, 100) })
		px := make([]byte, 4*640*400)
		dst.ReadPixels(px)
		s := menuTheme.Sel
		for i := 0; i < len(px); i += 4 {
			if px[i] == s.R && px[i+1] == s.G && px[i+2] == s.B {
				n++
			}
		}
		return
	}
	if blue(false) == 0 {
		t.Fatal("no highlight drawn at all: test can't see it")
	}
	if n := blue(true); n != 0 {
		t.Fatalf("NoPick menu lit %d px", n)
	}
}
