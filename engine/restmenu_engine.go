package engine

import "gobox/rules"

// restMenuText is BR's rest menu (explore.pas RestMenu: 'Rest Save Change Exit'), the hotkey text MenuInput takes. Save and
// Change (SaveQuit, CampAlterMenu) are not ported: the menu offers only what runs.
const restMenuText = "~Rest ~Exit"

// The rest menu's places.
const (
	restMenuRest = iota
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
	if n != restMenuRest {
		return
	}
	e.say(restTeamText)
	e.do(rules.AdvanceClock{Count: 1, Unit: clockDay})
	const minutesPerDay = 24 * 60
	if _, err := e.ruleset().Resolve(rules.Camping{Party: e.partyIDs()}, rules.Rest{Minutes: minutesPerDay}, nil); err != nil {
		panic("rules: Rest: " + err.Error())
	}
	e.askRest()
}
