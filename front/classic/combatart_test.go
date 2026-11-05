package classic

import (
	"testing"

	"gobox/game"
)

// BattleOf keeps a unit's DAX reference and resolves a tile name missing
// from the game's art against the series' combat sprites.
func TestBattleOfRefs(t *testing.T) {
	Series = "buck-rogers"
	defer func() { Series = "pool" }()
	s := game.State{Combat: &game.Combat{
		Tiles: []string{"", "hitek_tile_000"},
		Units: []game.Unit{{Art: game.Picture{File: "CPIC1", Block: 23, Clear: -1}}},
	}}
	b := BattleOf(s, nil)
	if b.TilePix[1].Img == nil {
		t.Error("tile not read from the series sprites")
	}
	if p := b.IconPix[0]; p.File != "CPIC1" || p.Block != 23 {
		t.Errorf("icon ref lost: %+v", p)
	}
}
