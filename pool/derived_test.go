package pool

import (
	"gobox/game"
	"gobox/game/play"
	"rules1e"
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
	if _, err := e.doNow(play.Ready{Member: m, Item: e.itemHandle(w)}); err != nil || w.Readied == 0 {
		t.Fatalf("Ready: %v", err)
	}
	on := page()
	if on.AC == before.AC {
		t.Errorf("after Ready AC %d, before %d", on.AC, before.AC)
	}
	if _, err := e.doNow(play.Ready{Member: m, Item: e.itemHandle(w)}); err != nil || w.Readied != 0 { // a fight's Ready toggles
		t.Fatalf("Unready: %v", err)
	}
	if off := page(); off.AC != before.AC {
		t.Errorf("after Unready AC %d, want %d", off.AC, before.AC)
	}
}
