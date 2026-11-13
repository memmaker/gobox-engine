//go:build devviews

package engine

import (
	"testing"

	"gobox/rules"
	"gobox/rules1e"
)

// TestUseItemOutside: drinking a potion in camp goes through Resolve(core.UseItem)
// and the potion is used up.
func TestUseItemOutside(t *testing.T) {
	e := quickStarted(t)
	e.CampMenu()
	e.fixture(rules.FixtureUsable, e.selected())
	e.OpenSheet(-1)
	e.OpenItems(-1, false)
	n := len(e.state().Carried(e.member(e.selected())))
	for i := 0; ; i++ {
		items := e.listedItems(e.selected())
		if i >= len(items) {
			t.Fatal("no potion")
		}
		h := items[i]
		if !e.state().Usable(e.heldOf(h)) {
			continue
		}
		if err := e.UseItem(h, 0, 0, true); err != nil {
			t.Fatal(err)
		}
		break
	}
	if m := len(e.state().Carried(e.member(e.selected()))); m != n-1 {
		t.Fatalf("items %d -> %d", n, m)
	}
}

// TestUsingItemPure: asking Allowed(core.UsingItem) about a covered scroll
// a cleric can read changes nothing; only the use uncovers it.
func TestUsingItemPure(t *testing.T) {
	e := quickStarted(t)
	e.CampMenu()
	s := e.state()
	c := e.member(e.selected())
	c.Class, c.Levels[rules1e.ClsCleric] = rules1e.ClsCleric, 3
	it := s.RandomItem(rules1e.ItemClericScroll)
	it.Hidden = 1
	it.Readied = 1
	s.AddItemCopy(c, it)
	held := s.Carried(c)[len(s.Carried(c))-1]
	q := rules.UsingItem{Who: e.charID(c), Item: e.itemHandle(held)}
	for range 2 {
		if o := e.ruleset().Allowed(q).Offers; len(o) == 0 || o[0].Not != "" {
			t.Fatal("the cleric reads no spell off the scroll")
		}
	}
	if held.Hidden != 1 || s.UsedScroll != nil {
		t.Errorf("Allowed changed state: Hidden %d, UsedScroll %v", held.Hidden, s.UsedScroll)
	}
}
