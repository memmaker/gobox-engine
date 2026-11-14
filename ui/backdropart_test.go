package ui

import (
	"bytes"
	"image"
	"testing"

	"gobox/asset"
	"gobox/game"
	"gobox/rules"
)

// TestBackdropArt: a menu's backdrop (still) holds the State as captured;
// F12 draws its picture in the new art all the same, from the DOS twin the
// picture carries (game.Picture.Alt), placed by its own cell offset.
func TestBackdropArt(t *testing.T) {
	art, am := daxArt.art, daxArt.amiga
	defer func() { daxArt.art = art; switchArt(am) }()
	daxArt.art = asset.OpenArt("../data/sotsb/art")
	if !hasAmiga() {
		t.Skip("no sotsb Amiga art")
	}
	dos := &game.Picture{W: 4, H: 1, Pixels: []byte{5, 6, 7, 8}, Clear: -1, CellX: 2, CellY: 1}
	amiga := game.Picture{File: "PIC", Block: 1, W: 2, H: 1, Pixels: []byte{1, 2}, Clear: -1, Alt: dos}
	bd := still{st: game.Scene{Common: game.Common{Picture: &amiga}}, ver: 1}
	drawn := func() Picture { return viewPictures(game.CommonOf(bd.State()), false)[0] }
	v := innerWindow.Pixels().Min
	switchArt(true)
	if f := drawn(); !bytes.Equal(f.Pix, amiga.Pixels) || f.At != image.Rect(0, 0, 2, 1).Add(v) {
		t.Fatalf("Amiga shown: %v at %v", f.Pix, f.At)
	}
	switchArt(false)
	f := drawn()
	if !bytes.Equal(f.Pix, dos.Pixels) || f.At != image.Rect(16, 8, 20, 9).Add(v) || f.File != "PIC" {
		t.Fatalf("DOS shown: %v at %v (%q)", f.Pix, f.At, f.File)
	}
}

// TestViewPictures: the presenter places the game's picture itself: a
// whole-screen one at the page's top left at its size, its Mark on its
// own cells.
func TestViewPictures(t *testing.T) {
	big := game.Picture{File: "BIGPIC1", Block: 1, W: 304, H: 120}
	c := game.Common{Picture: &big, Mark: &game.PictureMark{Picture: game.Picture{File: "CURSOR", Block: 13, W: 8, H: 8}, At: rules.MapPos{X: 3, Y: 2}}}
	ps := viewPictures(c, true)
	if len(ps) != 2 || ps[0].At != image.Rect(0, 0, 304, 120) || ps[1].At != image.Rect(24, 16, 32, 24) {
		t.Fatalf("pictures %+v", ps)
	}
}
