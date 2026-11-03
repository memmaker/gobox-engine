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
			h.Edit(func(e Game) { e.CreateCharacter(); *creating = e.CreationView() != nil })
		case ActModify:
			h.Edit(func(e Game) { e.ModifyCharacter(a.N); *creating = e.CreationView() != nil })
		default:
			return false
		}
		return true
	}
	switch a.Kind {
	case ActRoll:
		h.Edit(func(e Game) { e.Roll(a.N == 1, a.Text) })
	case ActName:
		h.Edit(func(e Game) { e.Rename(a.Text) })
	case ActStat:
		h.Edit(func(e Game) { e.SetStat(a.N, a.Text) })
	case ActSetField:
		h.Edit(func(e Game) { e.SetField(a.N, a.X) })
	case ActKeep, ActClose:
		*creating = false
		h.Edit(func(e Game) { e.EndCreation(a.Kind == ActKeep) })
	default:
		return false
	}
	return true
}
