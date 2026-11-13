package engine

import (
	"gobox/rules"
	"strconv"
)

// appraiseValue = the Gems and Jewelry picks of ovr22:19B2: a gem or a
// jewel of the party's is valued (gold pieces), and its item made (the
// shared tail of the branches); the player then keeps or sells it (keepValued).
func (e *Engine) appraiseValue(jewel bool) {
	if e.valued != rules.NoItem {
		return
	}
	if _, err := e.ruleset().Resolve(nil, rules.Valuate{Jewel: jewel}, nil); err != nil {
		e.say("Nothing to appraise")
		return
	}
	a := e.ruleset().Check(rules.ValuedItem{}, nil)
	e.valued = rules.ItemHandle(a.Value)
	what := "Gem"
	if a.Yes {
		what = "Jewel"
	}
	e.say("The " + what + " is Valued at " + strconv.Itoa(int(a.High)) + " gp.")
}

// keepValued: the one appraised is kept as an item when there is room
// (extra: the party inventory; the original lost it with an empty item
// list), else sold.
func (e *Engine) keepValued(keep bool) {
	if e.valued == rules.NoItem {
		return
	}
	e.valued = rules.NoItem
	e.do(rules.KeepValued{Who: e.partyIDs()[e.sel], Keep: keep})
}

// sellValued: leaving sells the one appraised (Esc on "Sell Keep").
func (e *Engine) sellValued() { e.keepValued(false) }
