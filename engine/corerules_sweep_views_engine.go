package engine

import (
	"fmt"
	"gobox/core"
)

// The screens' questions to the ruleset (core.InFight .. core.ItemCardFact).

// partySize is the members of the party.
func (e *Engine) partySize() int { return int(e.ruleset().Check(core.PartySize{}, nil).Value) }

// carryCount is what id can reach, the party inventory after its own.
func (e *Engine) carryCount(id core.CharID) int {
	return int(e.ruleset().Check(core.CarryCount{Who: id}, nil).Value)
}

// hallOpen: the training hall is up; learning: and waits for a spell pick.
func (e *Engine) hallOpen() (open, learning bool) {
	a := e.ruleset().Check(core.HallOpen{}, nil)
	return a.Yes, a.Value == 1
}

// fact is the ruleset's Fact f of id, a bug of the engine when it has none.
func (e *Engine) fact(id core.CharID, f core.Fact) core.Value {
	v, err := e.ruleset().Fact(id, f)
	if err != nil {
		panic(fmt.Sprintf("fact %T: %v", f, err))
	}
	return v
}

// listedItems is the items id lists, as the inventory orders them.
func (e *Engine) listedItems(id core.CharID) core.Handles {
	return e.fact(id, core.ListedFact{}).(core.Handles)
}

// goods are the shop's wares as listed.
func (e *Engine) goods() core.Handles {
	return e.fact(core.NoChar, core.ListedFact{Goods: true}).(core.Handles)
}

// itemCard is h as id sees it.
func (e *Engine) itemCard(id core.CharID, h core.ItemHandle) core.ItemCard {
	return e.fact(id, core.ItemCardFact{Item: h}).(core.ItemCard)
}

// purse is the party's coins; loot the fight's.
func (e *Engine) purse(loot bool) core.Purse {
	return e.fact(core.NoChar, core.PurseFact{Loot: loot}).(core.Purse)
}

// resolve is a's outcome outside any situation, for an action the rules cannot refuse.
func (e *Engine) resolve(a core.Action) core.Outcome {
	o, err := e.ruleset().Resolve(nil, a, nil)
	if err != nil {
		panic(fmt.Sprintf("rules refused %T: %v", a, err))
	}
	return o
}

// must is resolve for an action with no outcome to read.
func (e *Engine) must(a core.Action) { e.resolve(a) }
