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
