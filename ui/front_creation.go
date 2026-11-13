// creation.go is the presenter's creation (and Modify) sheet: opened from
// the party menu (Create, Modify) over CreateCharacter and ModifyCharacter,
// its changes calls (Roll, Rename, SetStat, SetField), left by
// EndCreation; the game keeps the character made (CreationView).

package ui

import (
	"gobox/game/play"
)

// CreateAct takes a decision a on the creation sheet (or one opening it),
// creating whether it is up: false leaves it to the front.
func CreateAct(h RuleHost, creating *bool, a Action) bool {
	if !*creating {
		switch a.Kind {
		case ActCreate:
			h.Edit(func(e *classicGame) { refused(e, e.do(play.NewCharacter{})); *creating = e.CreationView() != nil })
		case ActModify:
			h.Edit(func(e *classicGame) { refused(e, e.do(play.Modify{Member: a.N})); *creating = e.CreationView() != nil })
		default:
			return false
		}
		return true
	}
	switch a.Kind {
	case ActRoll:
		h.Edit(func(e *classicGame) { refused(e, e.do(play.Reroll{Minimal: a.N == 1})) })
	case ActName:
		h.Edit(func(e *classicGame) { refused(e, e.do(play.Rename{Name: a.Text})) })
	case ActStat:
		h.Edit(func(e *classicGame) { refused(e, e.do(play.SetStat{Stat: a.N, Text: a.Text})) })
	case ActSetField:
		h.Edit(func(e *classicGame) { refused(e, e.SetField(a.N, a.X)) })
	case ActKeep, ActClose: // a refused Keep leaves the sheet up, so Esc still discards
		h.Edit(func(e *classicGame) {
			refused(e, e.EndCreation(a.Kind == ActKeep))
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
