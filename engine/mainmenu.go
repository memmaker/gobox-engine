package engine

// The main menu (play.EndGame) drops the game in
// play and goes back to the party menu, the game's main menu. Ignored where a
// slot load could not unwind either (no game yet, a modal screen).

// mainMenu is play.EndGame; advance puts the party menu up.
func (e *Engine) mainMenu() {
	if e.ext.intro {
		return
	}
	e.stopVoice()
	if e.beginFlag() == 5 {
		e.combatViewLeave()
	}
	e.endGame()
}
