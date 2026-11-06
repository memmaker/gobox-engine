// statecalls.go: camp and the party menu are the presenter's: their
// decisions are rule calls (CampRest, PlanSpell, Scribe, DropMember,
// LeaveCamp; Begin, RemoveMember, NewClass, QuickStart, QuitGame; Lead).

package front

import (
	"errors"
	"os"
)

// StateAct makes a decision in camp or on the party menu a rule call;
// false leaves it to the others.
func StateAct(h Host, a Action) bool {
	v := h.Shown()
	if v.Ask || v.Powers || v.Sheet || v.Inventory {
		return false
	}
	n := a.N
	do := func(f func(e Game) error) { h.Call(func(e Game) { f(e) }) }
	switch {
	case v.Camp:
		switch a.Kind {
		case ActClose:
			do(func(e Game) error { return e.LeaveCamp() })
		case ActLead:
			do(func(e Game) error { return e.Lead(n) })
		case ActPlan:
			do(func(e Game) error { return e.PlanSpell(n, a.X, a.Y) })
		case ActScribe:
			do(func(e Game) error { return e.Scribe(n, a.X) })
		case ActDropMember:
			do(func(e Game) error { return e.DropMember(n) })
		case ActRest:
			do(func(e Game) error {
				if r, ok := e.(interface{ RestFor(int) error }); ok && a.X > 0 { // a picked duration, X minutes
					return r.RestFor(a.X)
				}
				return e.CampRest(n)
			})
		default:
			return false
		}
	case v.MainMenu:
		switch a.Kind {
		case ActLead:
			do(func(e Game) error { return e.Lead(n) })
		case ActBegin:
			do(func(e Game) error { return e.Begin() })
		case ActRemove:
			do(func(e Game) error { return e.RemoveMember(n) })
		case ActDropMember:
			do(func(e Game) error { return e.DropMember(n) })
		case ActQuit:
			h.Call(func(e Game) {
				if e.QuitGame() == nil { // the player confirmed: every game's run ends here
					os.Exit(0)
				}
			})
		case ActQuickStart:
			h.Call(func(e Game) { e.QuickStart() })
		case ActNewClass:
			NewClass(h, n, -1)
		default:
			return false
		}
	default:
		return false
	}
	return true
}

// NewClass calls NewClass; refused with PickClass, it asks for the class.
func NewClass(h Host, n, class int) {
	h.Call(func(e Game) {
		var pick PickClass
		if err := e.NewClass(n, class); errors.As(err, &pick) {
			AskChoice(h, pick.Error(), pick, func(id int) { NewClass(h, n, id) })
		}
	})
}

// AskChoice asks for one of items (AskMenu, with Exit); then gets its ID.
func AskChoice(h Host, prompt string, items []Spell, then func(id int)) {
	m := Menu{Prompt: prompt}
	for _, it := range items {
		m.Commands = append(m.Commands, Command{Text: it.Name, Key: it.Name[0]})
	}
	m.Commands = append(m.Commands, Command{Text: "Exit", Key: 'E'})
	h.Ask(&Question{Ask: Ask{Kind: AskMenu, Menu: m}, Answer: func(a Action) bool {
		switch a.Kind {
		case ActChoose:
			if a.N >= 0 && a.N < len(items) {
				then(items[a.N].ID)
			}
		case ActClose, ActNo:
		default:
			return false
		}
		return true
	}})
}
