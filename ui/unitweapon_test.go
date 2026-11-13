package ui

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestUnitWeapon: a member's unit names its readied weapon (the panel shows
// it in place of Dmg, as the originals); a foe's keeps Dmg.
func TestUnitWeapon(t *testing.T) {
	s := shown{Party: []game.Member{{Items: []game.Item{{Name: "Dagger", Kind: "weapon"}, {Name: "Long Sword", Kind: "weapon", Readied: true}}}},
		Combat: &play.Combat{Units: []game.Unit{{ID: 1, Member: 0}, {ID: 2, Member: -1, Foe: true, Dmg: "1d6"}}}}
	b := boardIn(s.Combat, s.Party)
	if b.Units[0].Weapon != "Long Sword" || b.Units[1].Weapon != "" {
		t.Fatalf("%q %q", b.Units[0].Weapon, b.Units[1].Weapon)
	}
	s.Party[0].Items = nil // SOTSB: the turn's Weapon is the source
	s.Combat.Turn = &play.Turn{Unit: 1, Weapon: &game.Weapon{Name: "Broad Sword"}}
	if w := boardIn(s.Combat, s.Party).Units[0].Weapon; w != "Broad Sword" {
		t.Fatalf("turn weapon %q", w)
	}
}
