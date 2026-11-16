package engine

import (
	"gobox/game"
	"gobox/rules"
)

// The gear menu's verbs beyond Ready and Drop (viewchar.pas UseItemMenu: Trade, Halve, Join) are the ruleset's: it offers
// them per item (rules.Handling), the engine puts each in Allowed under the game's action and runs the one asked.

// gearActs are the ruleset's gear verbs for member n's item h, as the game's actions.
func (e *Engine) gearActs(n int, h rules.ItemHandle) []game.Action {
	var out []game.Action
	for _, o := range e.ruleset().Allowed(rules.Handling{Who: e.partyIDs()[n], Item: h}).Offers {
		switch o.Action.(type) {
		case rules.GiveItem:
			out = append(out, game.Give{From: n, Item: h})
		case rules.Halve:
			out = append(out, game.Halve{Member: n, Item: h})
		case rules.JoinStack:
			out = append(out, game.Join{Member: n, Item: h})
		}
	}
	return out
}

// gear runs a gear verb of the selected member's item h on the open inventory; the ruleset's words for a refusal are
// the bottom row's.
func (e *Engine) gear(h rules.ItemHandle, act func(who rules.CharID) rules.Action) error {
	who, h := e.carriedBy(h)
	if h == rules.NoItem {
		return rules.ErrNotNow
	}
	o, err := e.ruleset().Resolve(nil, act(who), nil)
	for _, f := range o.Effects {
		if n, ok := f.(rules.Narrated); ok {
			e.refuse(n.Text)
		}
	}
	return err
}

// give gives the selected member's item h to member to (game.Give).
func (e *Engine) give(h rules.ItemHandle, to int) error {
	if to < 0 || to >= e.partySize() {
		return rules.ErrNotNow
	}
	return e.gear(h, func(who rules.CharID) rules.Action {
		return rules.GiveItem{From: who, To: e.partyIDs()[to], Item: h}
	})
}
