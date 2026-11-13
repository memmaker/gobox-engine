package engine

import (
	"testing"

	"gobox/asset"
)

// The combat board is rebuilt every frame: its ground tiles convert once
// (a per-frame picPix of all 256 was a quarter of the combat frame's CPU).
func TestTilePixOnce(t *testing.T) {
	var e Engine
	im := &asset.Image{Height: 2, Width: 1, Count: 1, BPP: 16, Data: make([]byte, 16)}
	a := e.tilePix(im, 0, 0)
	if b := e.tilePix(im, 0, 0); &a.Pix[0] != &b.Pix[0] {
		t.Error("tile converted again")
	}
	if e.groundOf(im) != e.groundOf(im) || len(e.grounds) != 1 {
		t.Error("ground counted again")
	}
}

// TestTilePixRelaid: tiles laid in place over a converted frame are converted again.
func TestTilePixRelaid(t *testing.T) {
	var e Engine
	im := &asset.Image{Height: 2, Width: 1, Count: 1, BPP: 16, Data: make([]byte, 16)}
	a := e.tilePix(im, 0, 0)
	im.Data[0] = 0x11
	e.tilesChanged()
	if b := e.tilePix(im, 0, 0); &a.Pix[0] == &b.Pix[0] || len(e.grounds) != 0 {
		t.Error("stale tile after the frames changed")
	}
}
