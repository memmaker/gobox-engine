package ui

import (
	"testing"

	"gobox/game"
	"gobox/rules"
)

// TestGroundOverlay: a square's Overlay (POD's clouds and bodies) is a tile
// of its own on the board's square, the tile under it kept; one picture,
// one slot.
func TestGroundOverlay(t *testing.T) {
	cloud := game.Picture{W: 1, H: 1, Pixels: []byte{3}, Clear: -1}
	s := shown{Combat: &game.Combat{Squares: []game.Ground{
		{At: rules.BoardPos{X: 2, Y: 3}, Tile: 5, Overlay: cloud}, {At: rules.BoardPos{X: 4, Y: 3}, Tile: 6, Overlay: cloud}, {At: rules.BoardPos{X: 5, Y: 3}, Tile: 6}}}}
	b := boardIn(s.Combat, s.Party)
	a, c := b.Ground[3][2], b.Ground[3][4]
	if a.Tile != 5 || c.Tile != 6 || a.Overlay == 0 || a.Overlay != c.Overlay || b.Ground[3][5].Overlay != 0 {
		t.Fatalf("squares %+v %+v %+v", a, c, b.Ground[3][5])
	}
	if b.Tiles[a.Overlay-1] == nil {
		t.Error("the overlay's picture is not in its slot")
	}
}

// TestGroundOverlayManyTiles: a game naming 0x100 tiles still gets its
// overlay a slot (POD's bodies showed bare floor).
func TestGroundOverlayManyTiles(t *testing.T) {
	names := make([]string, 0x100)
	body := game.Picture{File: "COMSPR", Block: 21, Clear: -1}
	s := shown{Combat: &game.Combat{Tiles: names, Squares: []game.Ground{{At: rules.BoardPos{X: 2, Y: 3}, Tile: 31, Overlay: body}}}}
	b := boardIn(s.Combat, s.Party)
	o := b.Ground[3][2].Overlay
	if o == 0 || len(b.Tiles) < o {
		t.Fatalf("overlay %d, slots %d", o, len(b.Tiles))
	}
}
