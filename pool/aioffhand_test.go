package pool

import (
	"gobox/rules1e"
	"testing"
)

type offHandRules struct{ rules1e.ClassicRules }

func (offHandRules) OffHand(*rules1e.RuleState, *rules1e.Character, *rules1e.Item) bool { return true }

// TestAIKeepsOffHand: the computer's weapon pick must never toggle a readied
// item off: a weapon in hand and one in the off hand stay readied over turns.
func TestAIKeepsOffHand(t *testing.T) {
	e := quickStarted(t)
	e.state().SetRules(offHandRules{})
	defer e.state().SetRules(nil)
	c := e.member(e.selected())
	e.state().ApplyItems(c)
	for _, s := range []int{rules1e.SlotWeapon, rules1e.SlotShield, rules1e.SlotArrows, rules1e.SlotBolts} {
		if c.Slots[s] != nil {
			e.state().ReadyItem(c.Slots[s])
			e.state().ApplyItems(c)
		}
	}
	main := e.state().AddItemCopy(c, e.state().RandomItem(rules1e.ItemDagger))
	off := e.state().AddItemCopy(c, e.state().RandomItem(rules1e.ItemLongSword))
	off.Plus = 3
	main.Cursed, off.Cursed = 0, 0
	c.ClassMask = 0xFF
	c.Combat = &rules1e.CombatInfo{}
	e.state().ReadyItem(main)
	e.state().ApplyItems(c)
	e.state().ReadyItem(off)
	e.state().ApplyItems(c)
	if main.Readied == 0 || off.Readied == 0 {
		t.Fatalf("setup: main %d, off %d", main.Readied, off.Readied)
	}
	for i := 0; i < 4; i++ {
		e.state().AiPickWeapon(c)
		if main.Readied == 0 || off.Readied == 0 {
			t.Fatalf("turn %d: main %d, off %d", i, main.Readied, off.Readied)
		}
	}
}

// TestNoOffHandUnder1e: under the original's rules (no off hand) a weapon
// readied in the off hand comes off at the next ApplyItems (U9).
func TestNoOffHandUnder1e(t *testing.T) {
	e := quickStarted(t)
	c := e.member(e.selected())
	dagger := e.state().RandomItem(rules1e.ItemDagger)
	dagger.Readied = rules1e.ReadiedOffHand
	off := e.state().AddItemCopy(c, dagger)
	e.state().ApplyItems(c)
	if off.Readied != 0 || c.Slots[rules1e.SlotShield] == off {
		t.Fatalf("off hand under 1e: readied %d", off.Readied)
	}
}
