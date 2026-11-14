package engine

import (
	"gobox/game"
	"slices"
	"strings"
)

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

// textPage is the text the game wrote: the pane's lines (earlier lots
// first), the prompt row, the lines over the view window (Scene.Placed).
type textPage struct {
	Pane   []string
	Prompt string
	View   [][]game.TextRun
}

// statePage is the text page as the state shows it.
func (e *Engine) statePage() *textPage {
	p := &textPage{Prompt: runsText(e.runs(e.pg.tx.prompt))}
	if pane := e.pagePane(); pane != nil {
		for _, l := range append(slices.Clone(e.pg.tx.more), pane.Lines...) {
			p.Pane = append(p.Pane, runsText(l))
		}
	}
	for _, l := range e.pageText() {
		p.View = append(p.View, viewLine(l))
	}
	return p
}

// runsText is a row's runs as plain text.
func runsText(rs []PaneRun) string {
	var b strings.Builder
	for _, r := range rs {
		b.WriteString(r.Text)
	}
	return b.String()
}

// viewLine is a row of text over the view window as a line: its runs in
// order in their colours, the gaps between them as spaces; where the row
// starts is gobox's (Scene.Placed).
func viewLine(rs []PaneRun) []game.TextRun {
	var l []game.TextRun
	end := -1
	for _, r := range rs {
		t := r.Text
		if end >= 0 && r.X > end {
			t = strings.Repeat(" ", r.X-end) + t
		}
		end = r.X + len(r.Text)
		l = append(l, game.TextRun{Text: t, Fg: r.Fg, Bg: r.Bg})
	}
	return l
}
