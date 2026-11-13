package engine

import "gobox/game"

// The screens State reports, read off the engine's own state: the ones
// publish puts in a View, without making one.

// campUp: camp is the screen (nothing opened over it).
func (e *Engine) campUp() bool { return e.onTop() == onCamp && !e.view.on && !e.snap.invOn }

// sheetUp: the character sheet's menu waits.
func (e *Engine) sheetUp() bool { return e.view.on && !e.snap.invOn }

// menuUp: the party menu is the screen (a pick owed stands in for it).
func (e *Engine) menuUp() bool {
	return !e.snap.invOn && !e.view.on && e.onTop() == onMenu && e.powersView() == nil
}

// battleUp: the combat map is the screen (not under a modal).
func (e *Engine) battleUp() bool { return e.battleShown() }

// stateMode says what the game waits for.
func (e *Engine) stateMode() game.Mode {
	switch {
	case e.ext.intro:
		return game.ModeTitle
	case e.question() != nil && e.inFight() && e.beginFlag() == 5: // a fight's own question (Continue Battle:): still the fight, with its Ask (BUG-CLASSES 46)
		return game.ModeCombat
	case e.question() != nil:
		return game.ModeAsk
	case e.trainingOpen(): // the hall's box over its party menu (OpenTraining)
		return game.ModeTraining
	case e.menuUp():
		return game.ModeParty
	case e.battleUp():
		return game.ModeCombat
	case e.waiting(PendShop):
		return game.ModeShop
	case e.waiting(PendTemple):
		return game.ModeTemple
	case e.waiting(PendDoor): // the Locked. menu over the 3D view (ovr15:0D6D): Exploring's Door
		return game.ModeExplore
	case e.pending == PendLoot:
		return game.ModeLoot
	case e.campUp():
		return game.ModeCamp
	case e.exploring() && e.powersView() == nil:
		return game.ModeExplore
	}
	return game.ModeNone
}

// statePage is the text page as a game.Page: its lines and colours, its pictures.
func (e *Engine) statePage() *game.Page {
	pics := e.pagePictures()
	p := &game.Page{Prompt: gameRuns(e.runs(e.pg.tx.prompt)), Layers: make([]game.Layer, 0, len(pics))}
	if pane := e.pagePane(); pane != nil {
		p.Pane = gameLines(pane.Lines)
		if more := e.pg.tx.more; len(more) > 0 {
			p.Pane, p.PaneRows = append(gameLines(more), p.Pane...), TextPaneRows
		}
	}
	p.View = gameLines(e.pageText())
	for _, l := range pics {
		g := game.Layer{Picture: gamePicture(l), At: l.At}
		g.File, g.Block, g.Frame = l.File, l.Block, l.Frame
		p.Layers = append(p.Layers, g)
	}
	return p
}
