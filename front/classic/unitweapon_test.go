package classic

import (
	"testing"

	"gobox/game"
)

// TestUnitWeapon: a member's unit names its readied weapon (the panel shows
// it in place of Dmg, as the originals); a foe's keeps Dmg.
func TestUnitWeapon(t *testing.T) {
	s := game.State{Party: []game.Member{{Items: []game.Item{{Name: "Dagger", Kind: "weapon"}, {Name: "Long Sword", Kind: "weapon", Readied: true}}}},
		Combat: &game.Combat{Units: []game.Unit{{ID: 1, Member: 0}, {ID: 2, Member: -1, Foe: true, Dmg: "1d6"}}}}
	b := BattleOf(s, nil)
	if b.Units[0].Weapon != "Long Sword" || b.Units[1].Weapon != "" {
		t.Fatalf("%q %q", b.Units[0].Weapon, b.Units[1].Weapon)
	}
	s.Party[0].Items = nil // SOTSB: the turn's Weapon is the source
	s.Combat.Turn = &game.Turn{Unit: 1, Weapon: &game.Weapon{Name: "Broad Sword"}}
	if w := BattleOf(s, nil).Units[0].Weapon; w != "Broad Sword" {
		t.Fatalf("turn weapon %q", w)
	}
}
