package classic

import (
	"image"
	"testing"

	"gobox/game"
)

// TestSquareArt: a square with its own Art draws it (a tile slot of its
// own), whatever the named tiles; two squares with one picture share it.
func TestSquareArt(t *testing.T) {
	art := game.Picture{File: "WILDCOM", Block: 1, Frame: 3}
	c := &game.Combat{Tiles: []string{"", "x"}, Squares: []game.Ground{{At: image.Pt(1, 1), Art: art}, {At: image.Pt(2, 1), Art: art}, {At: image.Pt(3, 1), Tile: 1}}}
	b := BattleOf(c, nil, nil)
	g := b.Ground[1]
	if g[1].Tile != 2 || g[2].Tile != 2 || g[3].Tile != 1 || b.TilePix[2].File != "WILDCOM" || b.TilePix[2].Frame != 3 {
		t.Fatalf("tiles %d %d %d, pix %+v", g[1].Tile, g[2].Tile, g[3].Tile, b.TilePix[2])
	}
}
