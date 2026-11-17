package ui

import (
	"testing"

	"gobox/asset"
	"gobox/game"
)

// TestTilePictureSeriesFirst: a combat tile the series has a sprite for draws that sprite, not the original's pixels
// (BR1's remastered floor and walls fell back to the DOS tiles). Detects the class: a remastered asset found by name
// that a picture the game also has shadows.
func TestTilePictureSeriesFirst(t *testing.T) {
	defer func(s string) { Series = s }(Series)
	Series = "buck-rogers"
	const name = "hitek_tile_034"
	want := asset.SpriteImage(Series, "combat", name)
	if want == nil {
		t.Fatal("no series sprite " + name)
	}
	if got := tilePicture(name, game.Picture{Pixels: []uint8{1}, W: 1}); got.Img != want {
		t.Errorf("%s draws the game's pixels, not the series sprite", name)
	}
	if got := tilePicture("combat/tile/3", game.Picture{Pixels: []uint8{1}, W: 1}); got.Pix == nil {
		t.Error("a tile with no series sprite lost its own pixels")
	}
}
