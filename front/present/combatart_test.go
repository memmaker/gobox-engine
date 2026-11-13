package present

import (
	"image"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// BattleOf keeps a unit's DAX reference and resolves a tile name missing
// from the game's art against the series' combat sprites.
func TestBattleOfRefs(t *testing.T) {
	Series = "buck-rogers"
	defer func() { Series = "pool" }()
	s := shown{Combat: &play.Combat{
		Tiles: []string{"", "hitek_tile_000"},
		Units: []game.Unit{{Art: game.Picture{File: "CPIC1", Block: 23, Clear: -1}}},
	}}
	b := boardIn(s.Combat, s.Party)
	if b.Tiles[1] == nil {
		t.Error("tile not read from the series sprites")
	}
	if b.FloorTile != nil {
		t.Error("a floor image with no FloorTile")
	}
	if len(b.Units) != 1 {
		t.Error("the unit is not on the board")
	}
}

// A Gone unit is not on the board; a Dead one stays.
func TestBattleOfGone(t *testing.T) {
	s := shown{Combat: &play.Combat{Units: []game.Unit{{ID: 1}, {ID: 2, Gone: true}, {ID: 3, Dead: true}}}}
	b := boardIn(s.Combat, s.Party)
	if len(b.Units) != 2 || b.Units[1].ID != 3 {
		t.Errorf("units %+v", b.Units)
	}
}

// FloorTile names a series combat sprite, which the board tiles as its floor.
func TestBattleOfFloorTile(t *testing.T) {
	Series = "buck-rogers"
	defer func() { Series = "pool" }()
	if b := boardIn(&play.Combat{FloorTile: "hitek_tile_034"}, nil); b.FloorTile == nil {
		t.Error("floor tile not read from the series sprites")
	}
}

// TestBattleMemberArt: a party unit with no icon to compose (COTAB's NPC
// members) is drawn from its Art, as a foe is, not as an empty disc.
func TestBattleMemberArt(t *testing.T) {
	s := shown{Party: []game.Member{{Name: "NPC"}}, Combat: &play.Combat{
		Units: []game.Unit{{ID: 1, Member: 0, Art: game.Picture{W: 2, H: 2, Pixels: []byte{1, 2, 3, 4}, Clear: -1}}},
	}}
	b := boardIn(s.Combat, s.Party)
	if sz := b.Units[0].IconSize; sz != image.Pt(2, 2) {
		t.Fatalf("member drawn at %v, want its Art's size", sz)
	}
}
