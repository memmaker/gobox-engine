package pool

import (
	"gobox/rules1e"
	"testing"
)

// The party inventory: unreadied items of members go into one pool with a
// slot count; stacks join; a readied bag of holding adds slots; a saved
// game keeps it with the first member's items.
func TestPartyPool(t *testing.T) {
	e := &Engine{}
	e.initVM()
	a, b := &rules1e.Character{}, &rules1e.Character{}
	e.state().Party = []*rules1e.Character{a, b}
	sword := &rules1e.Item{Type: rules1e.ItemLongSword, Readied: 1}
	a.Items = []*rules1e.Item{sword, {Type: rules1e.ItemDagger}}
	e.state().ApplyItems(a)
	if len(a.Items) != 1 || len(e.state().Pool) != 1 {
		t.Fatalf("stow: own %d pool %d", len(a.Items), len(e.state().Pool))
	}
	e.state().AddItemCopy(b, &rules1e.Item{Type: rules1e.ItemArrow, Count: 10})
	e.state().AddItemCopy(a, &rules1e.Item{Type: rules1e.ItemArrow, Count: 5})
	if len(e.state().Pool) != 2 || e.state().Pool[1].Count != 15 || len(b.Items) != 0 {
		t.Fatalf("arrows: pool %d", len(e.state().Pool))
	}
	for len(e.state().Pool) < rules1e.PoolSlots {
		e.state().AddItemCopy(b, &rules1e.Item{Type: rules1e.ItemDagger})
	}
	if !e.state().Overloaded(&rules1e.Item{Type: rules1e.ItemDagger}, a) || e.state().Overloaded(&rules1e.Item{Type: rules1e.ItemArrow, Count: 1}, a) {
		t.Error("full pool: a new item must not fit, arrows join their stack")
	}
	a.Items = append(a.Items, &rules1e.Item{Readied: 1, Words: [3]rules1e.WordID{0xBA}}) // bag of holding
	if e.state().PoolCap() != rules1e.PoolSlots+rules1e.PoolBag || e.state().Overloaded(&rules1e.Item{Type: rules1e.ItemDagger}, a) {
		t.Error("bag of holding")
	}
	if len(e.state().SavedItems(a, "CHRDATA1")) != len(a.Items)+len(e.state().Pool) || len(e.state().SavedItems(b, "CHRDATA2")) != 0 || len(e.state().SavedItems(a, "")) != len(a.Items) {
		t.Error("saved items")
	}
	d := e.state().Pool[2]
	e.state().RemoveItem(b, d)
	if len(e.state().Pool) != rules1e.PoolSlots-1 {
		t.Error("remove from pool")
	}
}

// A scroll in the party inventory is scribed by the member who chose to
// scribe it, not by whoever comes first in the party.
func TestPoolScribe(t *testing.T) {
	e := &Engine{}
	e.state().Items = make([]rules1e.ItemBase, rules1e.ItemMageScroll+1)
	e.initVM()
	e.state().Items[rules1e.ItemMageScroll].Slot = 0xB
	fighter, mage := &rules1e.Character{}, &rules1e.Character{}
	e.state().Party = []*rules1e.Character{fighter, mage}
	scroll := &rules1e.Item{Type: rules1e.ItemMageScroll, Spells: [3]rules1e.ItemPower{15 | rules1e.ScrollScribing}} // Magic Missile
	e.state().Pool = []*rules1e.Item{scroll}
	e.state().MarkScribe(mage, scroll, 0)
	if e.state().StudyTime(fighter) != 0 || e.state().StudyTime(mage) == 0 {
		t.Error("study time goes to the scriber only")
	}
	check := true
	if e.state().ScribeOne(fighter, &check) != 0 || e.state().ScribeOne(mage, &check) == 0 {
		t.Error("only the scriber scribes")
	}
}

// The spiritual hammer needs room in the party inventory, not in the
// member's own (readied-only) list.
func TestHammerPoolFull(t *testing.T) {
	e := &Engine{}
	e.initVM()
	a := &rules1e.Character{}
	e.state().Party = []*rules1e.Character{a}
	for len(e.state().Pool) < rules1e.PoolSlots {
		e.state().Pool = append(e.state().Pool, &rules1e.Item{Type: rules1e.ItemDagger})
	}
	e.state().SpiritualHammerItem(a, &rules1e.Affect{}, 0)
	if len(e.state().Pool) != rules1e.PoolSlots {
		t.Errorf("pool %d, hammer added to a full pool", len(e.state().Pool))
	}
}
