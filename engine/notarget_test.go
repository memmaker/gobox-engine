package engine

import (
	"gobox/rules1e"
	"slices"
	"testing"

	"gobox/game"
)

// TestNoTargetAsks (BUG-CLASSES 67): a UsePower cast or a UseItem sent with
// Target -1 (gobox's first, before any pick) goes through (the caster's or
// the party's spell) or is refused with the pick it needs (Refusal.Ask), a
// spell's then a member's (castAsk: "Cast Spell on whom"), never a bare error.
func TestNoTargetAsks(t *testing.T) {
	check := func(name string, e *Engine, a game.Action) {
		t.Helper()
		for range 4 {
			_, err := e.doNow(a)
			if r, ok := err.(*game.Refusal); ok && r.Ask != nil { // a pick gobox puts up
				switch r.Ask.Kind {
				case game.PickMember:
					return
				case game.PickSpell:
					a = r.Ask.Options[len(r.Ask.Options)-1].Do // a spell, as gobox's pick makes the call again
					continue
				}
			}
			if err != nil {
				t.Errorf("%s: %v, want it taken or a pick", name, err)
				return
			}
			pq := e.ext.pq
			if pq == nil {
				return
			}
			a = game.Reply{N: len(pq.calls) - 1} // a yes/no's Yes
		}
	}
	e := quickStarted(t)
	e.state().BeginFlag = 4
	ci := slices.IndexFunc(e.state().Party, func(c *rules1e.Character) bool { return c.Levels[rules1e.ClsCleric] > 0 })
	e.state().Party[ci].Memorized[0] = rules1e.SpCureLightWounds
	check("Cure Light Wounds", e, game.UsePower{Member: ci, Kind: game.PowerCast, Spell: rules1e.SpCureLightWounds, Target: -1})
	for _, typ := range []rules1e.ItemType{rules1e.ItemPotion, rules1e.ItemWand, rules1e.ItemClericScroll, rules1e.ItemMageScroll} {
		e := quickStarted(t)
		e.Start(e.GameLoop) // adventuring: the items are the party's to use
		playOn(e, 50, func(s game.State) bool { x, ok := s.(game.Exploring); return ok && x.Ask == nil })
		c := e.state().Party[0]
		it := e.state().AddItemCopy(c, e.state().RandomItem(typ))
		if typ == rules1e.ItemClericScroll {
			it.Spells = [3]rules1e.ItemPower{rules1e.ItemPower(rules1e.SpCureLightWounds) | rules1e.ScrollScribing}
		}
		e.state().ReadyItem(it)
		e.state().ApplyItems(c)
		it.Hidden = 0
		n := e.itemHandle(it)
		if !playAllows(e.build(), game.UseItem{Member: 0, Item: n}) {
			t.Fatalf("item %x: UseItem not allowed in %T (%d items)", typ, e.build(), len(commonOf(e.build()).Party[0].Items))
		}
		check(it.Name(), e, game.UseItem{Member: 0, Item: n, Target: -1})
	}
}
