package classic

import (
	"testing"

	"gobox/game"
)

func TestLoot(t *testing.T) {
	s := game.State{Mode: game.ModeLoot, Selected: 1, Loot: &game.Loot{Items: []game.Item{{Name: "Gauntlets"}}},
		Party: []game.Member{{}, {Spells: []game.SpellRef{{Name: detectMagic, Memorized: true}}}}}
	if l := Loot(s); l == nil || !l.Take || !l.Detect {
		t.Errorf("loot %+v", l)
	}
	s.Loot.Items = nil
	if l := Loot(s); l == nil || l.Take || l.Detect {
		t.Errorf("no treasure: %+v", l)
	}
}
