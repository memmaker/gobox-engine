// creation.go is the presenter's creation (and Modify) sheet: opened from
// the party menu (Create, Modify) over CreateCharacter and ModifyCharacter,
// its changes calls (Roll, Rename, SetStat, SetField), left by
// EndCreation; the game keeps the character made (CreationView).

package front

// CreateAct takes a decision a on the creation sheet (or one opening it),
// creating whether it is up: false leaves it to the front.
func CreateAct(h Host, creating *bool, a Action) bool {
	if !*creating {
		switch a.Kind {
		case ActCreate:
			h.Edit(func(e Game) { refused(e, e.CreateCharacter()); *creating = e.CreationView() != nil })
		case ActModify:
			h.Edit(func(e Game) { refused(e, e.ModifyCharacter(a.N)); *creating = e.CreationView() != nil })
		default:
			return false
		}
		return true
	}
	switch a.Kind {
	case ActRoll:
		h.Edit(func(e Game) { refused(e, e.Roll(a.N == 1, a.Text)) })
	case ActName:
		h.Edit(func(e Game) { refused(e, e.Rename(a.Text)) })
	case ActStat:
		h.Edit(func(e Game) { refused(e, e.SetStat(a.N, a.Text)) })
	case ActSetField:
		h.Edit(func(e Game) { refused(e, e.SetField(a.N, a.X)) })
	case ActKeep, ActClose: // a refused Keep leaves the sheet up, so Esc still discards
		h.Edit(func(e Game) { refused(e, e.EndCreation(a.Kind == ActKeep)); *creating = e.CreationView() != nil })
	default:
		return false
	}
	return true
}

// refused shows the refusal err (nil: none) as the game's Refuse.
func refused(e Game, err error) {
	if err != nil {
		e.Refuse(err.Error())
	}
}
