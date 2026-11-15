package engine

import (
	"gobox/game"
	"gobox/rules1e"
	"slices"
	"testing"
)

// TestDerivedFollowReady (BUG-CLASSES 61): the member's published AC is
// recomputed the moment armour is readied or unreadied.
func TestDerivedFollowReady(t *testing.T) {
	e := partyTurnScene(t)
	c := e.member(e.cs.actor)
	m := e.state().PartyIndex(c)
	if h := c.Slots[rules1e.SlotArmor]; h != nil {
		e.state().ReadyItem(h)
	}
	e.state().ApplyItems(c)
	w := e.state().AddItemCopy(c, e.state().RandomItem(rules1e.ItemPlateMail))
	page := func() game.Member { return commonOf(e.build()).Party[m] }
	before := page()
	if _, err := e.doNow(game.Ready{Member: m, Item: e.itemHandle(w)}); err != nil || w.Readied == 0 {
		t.Fatalf("Ready: %v", err)
	}
	on := page()
	if on.AC == before.AC {
		t.Errorf("after Ready AC %d, before %d", on.AC, before.AC)
	}
	if _, err := e.doNow(game.Unready{Member: m, Item: e.itemHandle(w)}); err != nil || w.Readied != 0 {
		t.Fatalf("Unready: %v", err)
	}
	if off := page(); off.AC != before.AC {
		t.Errorf("after Unready AC %d, want %d", off.AC, before.AC)
	}
}

// TestSpearEitherSlot: a weapon thrown or used in melee (a spear) readies
// into the free weapon hand, and beside a weapon in hand into the Missiles
// slot, taking no hands; unreadied it comes off again.
func TestSpearEitherSlot(t *testing.T) {
	e := partyTurnScene(t)
	c := e.member(e.cs.actor)
	m := e.state().PartyIndex(c)
	for _, s := range []int{rules1e.SlotWeapon, rules1e.SlotShield, rules1e.SlotMissiles, rules1e.SlotArrows, rules1e.SlotBolts} {
		if h := c.Slots[s]; h != nil {
			e.state().ReadyItem(h)
		}
	}
	e.state().ApplyItems(c)
	spear := e.state().AddItemCopy(c, e.state().RandomItem(rules1e.ItemSpear))
	sword := e.state().AddItemCopy(c, e.state().RandomItem(rules1e.ItemLongSword))
	ready := func(it *rules1e.Item) {
		t.Helper()
		if _, err := e.doNow(game.Ready{Member: m, Item: e.itemHandle(it)}); err != nil {
			t.Fatalf("Ready: %v", err)
		}
	}
	ready(spear)
	if c.Slots[rules1e.SlotWeapon] != spear {
		t.Fatalf("the spear is not in the free weapon hand: readied %d", spear.Readied)
	}
	if _, err := e.doNow(game.Unready{Member: m, Item: e.itemHandle(spear)}); err != nil || spear.Readied != 0 {
		t.Fatalf("Unready: %v, readied %d", err, spear.Readied)
	}
	ready(sword)
	hands := c.Hands
	ready(spear)
	if c.Slots[rules1e.SlotWeapon] != sword || c.Slots[rules1e.SlotMissiles] != spear || c.Hands != hands {
		t.Fatalf("weapon %v, missiles %v, hands %d (were %d)", c.Slots[rules1e.SlotWeapon], c.Slots[rules1e.SlotMissiles], c.Hands, hands)
	}
	if s := commonOf(e.build()).Party[m].Items; !slices.ContainsFunc(s, func(it game.Item) bool { return it.Handle == e.itemHandle(spear) && it.Slot == "missiles" }) {
		t.Errorf("the spear's Slot is not missiles: %+v", s)
	}
	if _, err := e.doNow(game.Unready{Member: m, Item: e.itemHandle(spear)}); err != nil || spear.Readied != 0 || c.Slots[rules1e.SlotMissiles] != nil {
		t.Fatalf("Unready from the Missiles slot: %v, readied %d", err, spear.Readied)
	}
}
