// statecalls.go: camp and the party menu are the presenter's: their
// decisions are rule calls (CampRest, PlanSpell, Scribe, DropMember,
// LeaveCamp; Begin, RemoveMember, QuickStart, QuitGame; Lead).

package ui

import (
	"gobox/game"
	"os"
)

// StateAct makes a decision in camp or on the party menu a rule call;
// false leaves it to the others.
func StateAct(h RuleHost, a Action) bool {
	v := h.Shown()
	if v.Ask || v.Powers || v.Sheet || v.Inventory {
		return false
	}
	n := a.N
	do := func(f func(e *classicGame) error) { h.Call(func(e *classicGame) { f(e) }) }
	switch {
	case v.Camp:
		switch a.Kind {
		case ActClose:
			do(func(e *classicGame) error { return e.do(game.BreakCamp{}) })
		case ActLead:
			do(func(e *classicGame) error { return e.do(game.SelectMember{Member: n}) })
		case ActPlan:
			do(func(e *classicGame) error { return e.do(game.Plan{Member: n, Spell: a.X, Count: a.Y}) })
		case ActScribe:
			do(func(e *classicGame) error { return e.do(game.Scribe{Member: n, Spell: a.X}) })
		case ActDropMember:
			do(func(e *classicGame) error { return e.do(game.DropMember{Member: n}) })
		case ActRest:
			do(func(e *classicGame) error {
				if a.X > 0 { // a picked duration, X minutes
					return e.RestFor(a.X)
				}
				return e.do(game.CampRest{Kind: n})
			})
		default:
			return false
		}
	case v.MainMenu:
		switch a.Kind {
		case ActLead:
			do(func(e *classicGame) error { return e.do(game.SelectMember{Member: n}) })
		case ActBegin:
			do(func(e *classicGame) error { return e.do(game.Begin{}) })
		case ActRemove:
			do(func(e *classicGame) error { return e.do(game.RemoveCharacter{Roster: n}) })
		case ActDropMember:
			do(func(e *classicGame) error { return e.do(game.DropMember{Member: n}) })
		case ActQuit:
			h.Call(func(e *classicGame) {
				if e.do(game.Quit{}) == nil { // the player confirmed: every game's run ends here
					os.Exit(0)
				}
			})
		case ActQuickStart:
			h.Call(func(e *classicGame) { e.do(game.QuickStart{}) })
		default:
			return false
		}
	default:
		return false
	}
	return true
}
