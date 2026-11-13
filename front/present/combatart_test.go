package present

import (
	"testing"

	"gobox/front"
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
	b := BattleOf(s.Combat, s.Party, nil)
	if b.TilePix[1].Img == nil {
		t.Error("tile not read from the series sprites")
	}
	if b.FloorImg != nil {
		t.Error("a floor image with no FloorTile")
	}
	if p := b.IconPix[0]; p.File != "CPIC1" || p.Block != 23 {
		t.Errorf("icon ref lost: %+v", p)
	}
}

// A Gone unit is not on the board; a Dead one stays.
func TestBattleOfGone(t *testing.T) {
	s := shown{Combat: &play.Combat{Units: []game.Unit{{ID: 1}, {ID: 2, Gone: true}, {ID: 3, Dead: true}}}}
	b := BattleOf(s.Combat, s.Party, nil)
	if len(b.Units) != 2 || len(b.IconPix) != 2 || b.Units[1].ID != 3 {
		t.Errorf("units %+v", b.Units)
	}
	if o := Battle(b, &play.Combat{Units: []game.Unit{{ID: 1}, {ID: 2}, {ID: 3, Gone: true}}}, nil); len(o.Units) != 1 || len(o.IconPix) != 1 {
		t.Errorf("Battle kept a gone unit: %+v", o.Units)
	}
}

// FloorTile names a series combat sprite, which the board tiles as its floor.
func TestBattleOfFloorTile(t *testing.T) {
	Series = "buck-rogers"
	defer func() { Series = "pool" }()
	if b := BattleOf(&play.Combat{FloorTile: "hitek_tile_034"}, nil, nil); b.FloorImg == nil {
		t.Error("floor tile not read from the series sprites")
	}
}

// TestBattleMemberArt: a party unit with no icon to compose (COTAB's NPC
// members) is drawn from its Art, as a foe is, not as an empty disc.
func TestBattleMemberArt(t *testing.T) {
	s := shown{Party: []game.Member{{Name: "NPC"}}, Combat: &play.Combat{
		Units: []game.Unit{{ID: 1, Member: 0, Art: game.Picture{File: "CPIC1", Block: 85, Clear: -1}}},
	}}
	b := Battle(&front.Battle{Units: []front.Unit{{ID: 1}}, IconPix: []front.Picture{{}}}, s.Combat, s.Party)
	if p := b.IconPix[0]; p.File != "CPIC1" || p.Block != 85 {
		t.Fatalf("member drawn from %+v, want its Art", p)
	}
}
