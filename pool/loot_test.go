package pool

import (
	"rules1e"
	"slices"
	"testing"

	"gobox/core"
	"gobox/game/play"
)

// TestLootContract: the treasure after a fight is play's Looting, no
// inventory of por's own up; Take takes one piece or all for the selected
// member; leaving with treasure left asks Yes/No (pending data): Yes goes
// back to the treasure, No leaves.
func TestLootContract(t *testing.T) {
	e := quickStarted(t)
	e.selectMember(0)
	e.state().Treasure = []*rules1e.Item{e.state().RandomItem(rules1e.ItemJavelin), e.state().RandomItem(rules1e.ItemGauntlets), e.state().RandomItem(rules1e.ItemArrow)}
	e.push(cont{kind: kNop})
	left := func() bool { return !slices.ContainsFunc(e.k, func(c cont) bool { return c.kind == kNop }) } // the cont is done: the treasure screen is left
	e.Start(e.lootMenu)
	seeLoot(t, e)
	items := func() int {
		l, ok := e.build().(play.Looting)
		if !ok {
			return -1
		}
		return len(l.Items)
	}
	if n := items(); n != 3 || e.snap.invOn {
		t.Fatalf("state %T, %d pieces, inventory up %v: want Looting, 3 pieces, none", e.build(), n, e.snap.invOn)
	}
	if playAllows(e.build(), play.LootShare{}) {
		t.Error("Share offered: the money is the party's already")
	}
	held := len(e.state().Pool) // taken loot goes to the party inventory
	if _, err := e.doNow(play.Take{Item: e.goods()[0]}); err != nil {
		t.Fatal(err)
	}
	if n := items(); n != 2 || len(e.state().Pool) != held+1 {
		t.Fatalf("after one: %d left, %d held, want 2 and %d", n, len(e.state().Pool), held+1)
	}
	if _, err := e.doNow(play.Take{Item: core.NoItem}); err == nil || items() != 2 { // BUG-CLASSES 31: nothing to take is a refusal
		t.Fatalf("the money (none) took items or was accepted: %v", err)
	}
	if _, err := e.doNow(play.Leave{}); err != nil || left() {
		t.Fatalf("leave with treasure left: %v, left %v", err, left())
	}
	q := commonOf(e.build()).Ask
	if q == nil || !q.YesNo || q.Text != "There is still treasure left.  Do you want to go back and claim your treasure?" {
		t.Fatalf("ask %+v: want ovr06's Yes/No", q)
	}
	if _, err := e.doNow(play.Reply{N: 1}); err != nil || left() || items() != 2 { // Yes: back to the treasure
		t.Fatalf("Yes: %v, left %v, %d pieces", err, left(), items())
	}
	if _, err := e.doNow(play.Take{Item: core.NoItem, All: true}); err != nil || len(e.state().Treasure) != 0 {
		t.Fatalf("all: %v, %d left", err, len(e.state().Treasure))
	}
	if _, err := e.doNow(play.Leave{}); err != nil || !left() || items() >= 0 {
		t.Fatalf("leave with nothing left: %v, left %v, %T", err, left(), e.build())
	}
}
