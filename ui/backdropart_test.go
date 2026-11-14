package ui

import (
	"bytes"
	"image"
	"testing"

	"gobox/asset"
	"gobox/game"
)

// TestBackdropArt: a menu's backdrop (still) holds the State as captured;
// F12 draws its picture in the new art all the same, from the DOS twin the
// picture carries (game.Picture.Alt), placed by its own offset.
func TestBackdropArt(t *testing.T) {
	art, am := daxArt.art, daxArt.amiga
	defer func() { daxArt.art = art; switchArt(am) }()
	daxArt.art = asset.OpenArt("../data/sotsb/art")
	if !hasAmiga() {
		t.Skip("no sotsb Amiga art")
	}
	dos := &game.Picture{W: 4, H: 1, Pixels: []byte{5, 6, 7, 8}, Clear: -1, Off: image.Pt(-1, 0)}
	amiga := game.Picture{File: "PIC", Block: 1, W: 2, H: 1, Pixels: []byte{1, 2}, Clear: -1, Alt: dos}
	at := image.Rect(10, 20, 12, 21)
	bd := still{st: game.Scene{Layers: []game.Drawn{{Picture: amiga, At: at}}}, ver: 1}
	drawn := func() Picture { return drawnLayer(bd.State().(game.Scene).Layers[0]) }
	switchArt(true)
	if f := drawn(); !bytes.Equal(f.Pix, amiga.Pixels) || f.At != at {
		t.Fatalf("Amiga shown: %v at %v", f.Pix, f.At)
	}
	switchArt(false)
	f := drawn()
	if !bytes.Equal(f.Pix, dos.Pixels) || f.At != image.Rect(9, 20, 13, 21) || f.File != "PIC" {
		t.Fatalf("DOS shown: %v at %v (%q)", f.Pix, f.At, f.File)
	}
}
