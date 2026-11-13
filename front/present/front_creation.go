// creation.go is the presenter's creation (and Modify) sheet: opened from
// the party menu (Create, Modify) over CreateCharacter and ModifyCharacter,
// its changes calls (Roll, Rename, SetStat, SetField), left by
// EndCreation; the game keeps the character made (CreationView).

package present

import "gobox/front"

// CreateAct takes a decision a on the creation sheet (or one opening it),
// creating whether it is up: false leaves it to the front.
func CreateAct(h RuleHost, creating *bool, a front.Action) bool {
	if !*creating {
		switch a.Kind {
		case front.ActCreate:
			h.Edit(func(e *classicGame) { refused(e, e.CreateCharacter()); *creating = e.CreationView() != nil })
		case front.ActModify:
			h.Edit(func(e *classicGame) { refused(e, e.ModifyCharacter(a.N)); *creating = e.CreationView() != nil })
		default:
			return false
		}
		return true
	}
	switch a.Kind {
	case front.ActRoll:
		h.Edit(func(e *classicGame) { refused(e, e.Roll(a.N == 1, a.Text)) })
	case front.ActName:
		h.Edit(func(e *classicGame) { refused(e, e.Rename(a.Text)) })
	case front.ActStat:
		h.Edit(func(e *classicGame) { refused(e, e.SetStat(a.N, a.Text)) })
	case front.ActSetField:
		h.Edit(func(e *classicGame) { refused(e, e.SetField(a.N, a.X)) })
	case front.ActKeep, front.ActClose: // a refused Keep leaves the sheet up, so Esc still discards
		h.Edit(func(e *classicGame) {
			refused(e, e.EndCreation(a.Kind == front.ActKeep))
			*creating = e.CreationView() != nil
		})
	default:
		return false
	}
	return true
}

// refused shows the refusal err (nil: none) as the game's Refuse.
func refused(e *classicGame, err error) {
	if err != nil {
		e.Refuse(err.Error())
	}
}
