package parity

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"por/rules"
)

// samePlace says the State's temple, shop and inventory are por's at name.
func samePlace(t *testing.T, name string, e *rules.Engine) {
	t.Helper()
	at(t, name, e)
}

// members shows each member's items in turn (the inventory up).
func members(t *testing.T, name string, e *rules.Engine) {
	t.Helper()
	for i := range e.Rules().State().Party {
		if err := e.ShowMember(i); err != nil {
			t.Fatalf("%s: member %d: %v", name, i, err)
		}
		settled(e, front.Action{})
		samePlace(t, name, e)
	}
}

// TestInventoryParity: the shop (goods, Id, its reveal, selling), the
// appraisal, the temple member by member, the inventory and the treasure
// over it, CAPTURE's members' items.
func TestInventoryParity(t *testing.T) {
	sh := engine(t, true, "shop")
	settled(sh, front.Action{})
	samePlace(t, "shop", sh)
	s := sh.Rules().State()
	if len(s.Shop.Unknown) == 0 {
		t.Fatal("shop: nothing to identify")
	}
	for _, u := range s.Shop.Unknown {
		if err := sh.Identify(u.Member, u.Item); err != nil {
			t.Fatalf("identify %+v: %v", u, err)
		}
		settled(sh, front.Action{})
		samePlace(t, "identified", sh)
	}
	if sh.Rules().State().Shop.Identified == nil {
		t.Error("no reveal")
	}
	if err := sh.Browse(-1); err != nil {
		t.Fatal(err)
	}
	settled(sh, front.Action{})
	samePlace(t, "browse", sh)
	members(t, "browse", sh)

	ap := engine(t, true, "appraise")
	settled(ap, front.Action{})
	samePlace(t, "appraise", ap)
	if err := ap.Appraise(false); err != nil {
		t.Fatal(err)
	}
	settled(ap, front.Action{})
	samePlace(t, "appraised", ap)

	tp := engine(t, true, "temple")
	settled(tp, front.Action{})
	for i := range tp.Rules().State().Party {
		tp.Rules().Apply(game.Select{N: i})
		settled(tp, front.Action{})
		samePlace(t, "temple", tp)
	}

	inv := engine(t, true, "inventory")
	settled(inv, front.Action{})
	if err := inv.OpenItems(0, true); err != nil {
		t.Fatal(err)
	}
	settled(inv, front.Action{})
	if inv.Rules().State().Items == game.ItemsNone {
		t.Fatal("no inventory")
	}
	members(t, "inventory", inv)

	l := engine(t, true, "loot")
	settled(l, front.Action{})
	if s := l.Rules().State(); s.Loot == nil || s.Items != game.ItemsNone { // gobox's loot screen: the game opens no inventory of its own
		t.Fatalf("treasure %v items %v", s.Loot != nil, s.Items)
	}
	samePlace(t, "loot", l)

	c := capture(t) // no one carries anything: the empty inventory, from the sheet
	if err := c.OpenSheet(0); err != nil {
		t.Fatal(err)
	}
	settled(c, front.Action{})
	if err := c.OpenItems(0, false); err != nil {
		t.Fatal(err)
	}
	settled(c, front.Action{})
	members(t, "capture", c)
}
