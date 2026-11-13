// statecalls.go: camp and the party menu are the presenter's: their
// decisions are rule calls (CampRest, PlanSpell, Scribe, DropMember,
// LeaveCamp; Begin, RemoveMember, NewClass, QuickStart, QuitGame; Lead).

package present

import (
	"gobox/front"
	"os"
)

// StateAct makes a decision in camp or on the party menu a rule call;
// false leaves it to the others.
func StateAct(h RuleHost, a front.Action) bool {
	v := h.Shown()
	if v.Ask || v.Powers || v.Sheet || v.Inventory {
		return false
	}
	n := a.N
	do := func(f func(e *classicGame) error) { h.Call(func(e *classicGame) { f(e) }) }
	switch {
	case v.Camp:
		switch a.Kind {
		case front.ActClose:
			do(func(e *classicGame) error { return e.LeaveCamp() })
		case front.ActLead:
			do(func(e *classicGame) error { return e.Lead(n) })
		case front.ActPlan:
			do(func(e *classicGame) error { return e.PlanSpell(n, a.X, a.Y) })
		case front.ActScribe:
			do(func(e *classicGame) error { return e.Scribe(n, a.X) })
		case front.ActDropMember:
			do(func(e *classicGame) error { return e.DropMember(n) })
		case front.ActRest:
			do(func(e *classicGame) error {
				if a.X > 0 { // a picked duration, X minutes
					return e.RestFor(a.X)
				}
				return e.CampRest(n)
			})
		default:
			return false
		}
	case v.MainMenu:
		switch a.Kind {
		case front.ActLead:
			do(func(e *classicGame) error { return e.Lead(n) })
		case front.ActBegin:
			do(func(e *classicGame) error { return e.Begin() })
		case front.ActRemove:
			do(func(e *classicGame) error { return e.RemoveMember(n) })
		case front.ActDropMember:
			do(func(e *classicGame) error { return e.DropMember(n) })
		case front.ActQuit:
			h.Call(func(e *classicGame) {
				if e.QuitGame() == nil { // the player confirmed: every game's run ends here
					os.Exit(0)
				}
			})
		case front.ActQuickStart:
			h.Call(func(e *classicGame) { e.QuickStart() })
		case front.ActNewClass:
			NewClass(h, n, -1)
		default:
			return false
		}
	default:
		return false
	}
	return true
}

// NewClass calls NewClass (the game asks for the class itself).
func NewClass(h RuleHost, n, class int) {
	h.Call(func(e *classicGame) { e.NewClass(n, class) })
}
