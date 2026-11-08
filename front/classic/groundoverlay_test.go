package classic

import (
	"image"
	"testing"

	"gobox/game"
)

// TestGroundOverlay: a square's Overlay (POD's clouds and bodies) is a tile
// of its own on the board's square, the tile under it kept; one picture,
// one slot.
func TestGroundOverlay(t *testing.T) {
	cloud := game.Picture{W: 1, H: 1, Pixels: []byte{3}, Clear: -1}
	s := game.State{Combat: &game.Combat{Squares: []game.Ground{
		{At: image.Pt(2, 3), Tile: 5, Overlay: cloud}, {At: image.Pt(4, 3), Tile: 6, Overlay: cloud}, {At: image.Pt(5, 3), Tile: 6}}}}
	b := BattleOf(s.Combat, s.Party, nil)
	a, c := b.Ground[3][2], b.Ground[3][4]
	if a.Tile != 5 || c.Tile != 6 || a.Overlay == 0 || a.Overlay != c.Overlay || b.Ground[3][5].Overlay != 0 {
		t.Fatalf("squares %+v %+v %+v", a, c, b.Ground[3][5])
	}
	if b.TilePix[a.Overlay-1].Pix == nil {
		t.Error("the overlay's picture is not in its slot")
	}
}

// TestGroundOverlayManyTiles: a game naming 0x100 tiles still gets its
// overlay a slot (POD's bodies showed bare floor).
func TestGroundOverlayManyTiles(t *testing.T) {
	names := make([]string, 0x100)
	body := game.Picture{File: "COMSPR", Block: 21, Clear: -1}
	s := game.State{Combat: &game.Combat{Tiles: names, Squares: []game.Ground{{At: image.Pt(2, 3), Tile: 31, Overlay: body}}}}
	b := BattleOf(s.Combat, s.Party, nil)
	o := b.Ground[3][2].Overlay
	if o == 0 || b.TilePix[o-1].File != "COMSPR" {
		t.Fatalf("overlay %d, slots %d", o, len(b.TilePix))
	}
}
