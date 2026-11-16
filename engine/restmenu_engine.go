package engine

import (
	"gobox/game"
	"gobox/rules"
)

// restMenuText is BR's rest menu (explore.pas RestMenu: 'Rest Save Change Exit'), the hotkey text MenuInput takes. Change
// is the Change menu (change_engine.go).
const restMenuText = "~Rest ~Save ~Change ~Exit"

// The rest menu's places.
const (
	restMenuRest = iota
	restMenuSave
	restMenuChange
	restMenuExit
)

// restTeamText is what the rest menu says as the team rests (explore.pas RestMenu).
const restTeamText = "The team rests."

// askRest asks the rest menu: the script that opened it goes on once Exit answers (opProgram's holdRest).
func (e *Engine) askRest() {
	e.askChoice("", restMenuText, false, false, askDone{kind: dRestMenu})
}

// restMenuAnswered takes the rest menu's answer: Rest is a day (gametime.pas RestAndHeal(1): the clock a day on, the
// ruleset's rest of that day), then the menu again; Exit leaves.
func (e *Engine) restMenuAnswered(n int) {
	switch n {
	case restMenuRest:
		e.say(restTeamText)
		e.restDays(1)
		e.askRest()
	case restMenuSave:
		e.saveQuit(dRestSaveQuit)
	case restMenuChange:
		e.openChange(false)
	}
}

// saveQuitText is SaveQuit's question (explore.pas SaveQuit: 'Quit TO DOS ', default No).
const saveQuitText = "Quit TO DOS "

// saveQuit is explore.pas SaveQuit: the save screen (the player picks the slot), then the question whether to quit; done is its kind.
func (e *Engine) saveQuit(done doneKind) {
	e.event(game.SaveOffered{})
	q := yesNo(saveQuitText, false)
	q.Default = 0
	e.askYesNo(q, false, askDone{kind: done})
}

// saveQuitAnswered takes SaveQuit's answer: yes ends the run (the front ends it at game.Ended); false: the menu goes on.
func (e *Engine) saveQuitAnswered(n int) bool {
	if n != 1 {
		return false
	}
	e.event(game.Ended{})
	return true
}

// restDays is gametime.pas RestAndHeal(days): the clock days on, the ruleset's rest of those days.
func (e *Engine) restDays(days int) {
	e.do(rules.AdvanceClock{Count: int32(days), Unit: clockDay})
	const minutesPerDay = 24 * 60
	if _, err := e.ruleset().Resolve(rules.Camping{Party: e.partyIDs()}, rules.Rest{Minutes: int32(days * minutesPerDay)}, nil); err != nil {
		panic("rules: Rest: " + err.Error())
	}
}
