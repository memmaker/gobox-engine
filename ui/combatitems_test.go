package ui

import (
	"testing"

	"gobox/game"
)

// TestCombatItemsList: the fight's Items are the acting member's items the
// rules allow UseItem for, nothing else.
func TestCombatItemsList(t *testing.T) {
	wand, sword := game.Item{Handle: 1, Name: "Wand"}, game.Item{Handle: 2, Name: "Sword"}
	f := game.Fighting{Common: game.Common{Party: []game.Member{{Name: "A", Items: []game.Item{wand, sword}}},
		Allowed: game.Allowed{Plain: []game.Action{game.UseItem{Member: 0, Item: wand.Handle}}}},
		Field: game.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, Member: 0}}}}
	var r frontState
	c := r.combat(f)
	if len(c.Items) != 1 || c.Items[0].Handle != wand.Handle {
		t.Fatalf("items %+v, want the wand only", c.Items)
	}
}

// TestNoItemsIcon: with nothing usable (no UseItem allowed) the bar has
// no Items.
func TestNoItemsIcon(t *testing.T) {
	f := game.Fighting{Common: game.Common{Party: []game.Member{{Name: "A", Items: []game.Item{{Handle: 2, Name: "Sword"}}}},
		Allowed: game.Allowed{Plain: []game.Action{game.Guard{}}}},
		Field: game.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, Member: 0}}}}
	var r frontState
	for _, o := range r.combat(f).Turn.Options {
		if _, ok := o.Do.(game.OpenItems); ok {
			t.Fatalf("Items on the bar: %+v", o)
		}
	}
}
