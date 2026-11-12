package pool

import (
	"rules1e"
	"slices"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestNoTargetAsks (BUG-CLASSES 67): a UsePower cast or a UseItem sent with
// Target -1 (gobox's first, before any pick) goes through (the caster's or
// the party's spell) or is asks the pick it needs (pending), a spell's then a
// member's (castAsk: "Cast Spell on whom"), never a bare error.
func TestNoTargetAsks(t *testing.T) {
	check := func(name string, e *Engine, a play.Action) {
		t.Helper()
		if _, err := e.Do(a); err != nil {
			t.Errorf("%s: %v, want it taken or a pick", name, err)
			return
		}
		for range 3 {
			pq := e.ext.pq
			if pq == nil || pq.of == askMember {
				return
			}
			if _, err := e.Do(play.Reply{N: len(pq.calls) - 1}); err != nil { // a spell; a yes/no's Yes
				t.Errorf("%s: reply: %v", name, err)
				return
			}
		}
	}
	e := quickStarted(t)
	e.state().BeginFlag = 4
	ci := slices.IndexFunc(e.state().Party, func(c *rules1e.Character) bool { return c.Levels[rules1e.ClsCleric] > 0 })
	e.state().Party[ci].Memorized[0] = rules1e.SpCureLightWounds
	check("Cure Light Wounds", e, play.UsePower{Member: ci, Kind: game.PowerCast, Spell: rules1e.SpCureLightWounds, Target: -1})
	for _, typ := range []rules1e.ItemType{rules1e.ItemPotion, rules1e.ItemWand, rules1e.ItemClericScroll, rules1e.ItemMageScroll} {
		e := quickStarted(t)
		e.Start(e.GameLoop) // adventuring: the items are the party's to use
		playOn(e, 50, func(s play.State) bool { x, ok := s.(play.Exploring); return ok && x.Ask == nil })
		c := e.state().Party[0]
		it := e.state().AddItemCopy(c, e.state().RandomItem(typ))
		if typ == rules1e.ItemClericScroll {
			it.Spells = [3]rules1e.SpellID{rules1e.SpCureLightWounds | 0x80}
		}
		e.state().ReadyItem(it)
		e.state().ApplyItems(c)
		it.Hidden = 0
		n := e.itemHandle(it)
		if !playAllows(e.State(), play.UseItem{Member: 0, Item: n}) {
			t.Fatalf("item %x: UseItem not allowed in %T (%d items)", typ, e.State(), len(commonOf(e.State()).Party[0].Items))
		}
		check(it.Name(), e, play.UseItem{Member: 0, Item: n, Target: -1})
	}
}
