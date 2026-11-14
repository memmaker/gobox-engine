package ui

import (
	"image"
	"path/filepath"
	"testing"

	"gobox/asset"
	"gobox/game"
)

// TestAreaMapFollowsArt: F12 (switchArt) rebuilds the overland's square
// and marker pictures, though the game's art map stays the same (the
// twin is picked in the front: shownPic).
func TestAreaMapFollowsArt(t *testing.T) {
	defer func(a *asset.Art, am bool) { daxArt.art = a; switchArt(am) }(daxArt.art, daxArt.amiga)
	daxArt.art = asset.OpenArt(filepath.Join("..", "data", "por", "art"))
	if !hasAmiga() {
		t.Skip("por: no imported Amiga art")
	}
	pc := game.Picture{W: 1, H: 1, Pixels: []byte{1}, Alt: &game.Picture{W: 1, H: 1, Pixels: []byte{2}}}
	art := map[string]game.Picture{"terrain/0": pc, "marker/0": pc}
	var a areaPic
	a.resolve([]int{0}, image.Pt(1, 1), art)
	sq, mk := a.pics[0][0], a.marker("marker/0", art)
	switchArt(!daxArt.amiga)
	a.resolve([]int{0}, image.Pt(1, 1), art)
	if a.pics[0][0].At(0, 0) == sq.At(0, 0) {
		t.Error("square picture kept over F12")
	}
	if a.marker("marker/0", art).At(0, 0) == mk.At(0, 0) {
		t.Error("marker picture kept over F12")
	}
}
