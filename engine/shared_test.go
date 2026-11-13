package engine

import (
	"gobox/rules1e"
	"slices"
	"testing"
)

// TestSharedPurseAndItems (BUG-CLASSES 76): money and loose items are the
// party's. A member joining brings its coins into the purse and its
// unreadied items into the party inventory; its readied gear stays on it
// (the sheet's slots); no member holds coins.
func TestSharedPurseAndItems(t *testing.T) {
	e := quickStarted(t)
	id, ok := e.LoadCharacter(e.state().Party[0].Name(), false)
	if !ok {
		t.Fatal("no saved character")
	}
	c := e.member(id)
	c.Coins = [7]uint16{0, 0, 0, 37} // gold
	loose := e.state().RandomItem(rules1e.ItemDagger)
	loose.Readied = 0
	c.Items = append(c.Items, loose)
	gold, pool := e.state().Purse.Gold, len(e.state().Pool)
	e.state().AddToParty(c)
	if e.state().Purse.Gold != gold+37 || c.Coins != [7]uint16{} {
		t.Errorf("purse %d (was %d), the member keeps %v", e.state().Purse.Gold, gold, c.Coins)
	}
	if slices.Contains(c.Items, loose) || len(e.state().Pool) != pool+1 {
		t.Errorf("the loose dagger: on the member %v, pool %d (was %d)", slices.Contains(c.Items, loose), len(e.state().Pool), pool)
	}
	for _, m := range e.state().Party {
		if m.Coins != [7]uint16{} {
			t.Errorf("%s holds coins %v", m.Name(), m.Coins)
		}
		for _, it := range m.Items {
			if it.Readied == 0 {
				t.Errorf("%s holds an unreadied %s", m.Name(), it.Name())
			}
		}
	}
}
