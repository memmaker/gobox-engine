package pool

import (
	"gobox/core"
	"time"

	"gobox/core/goldbox"
	"gobox/game"
)

// partyFn is a function of the party menu (partyFunctions).
type partyFn int

const (
	partyNone partyFn = iota
	partyBegin
	partyAdd
	partyCreate
	partyDrop
	partyModify
	partyLoad
	partyQuickStart
	partyForget
	partyQuit
	partyTraining
	partyNewClass
	partyView
	partyRemove
)

// Start runs body (Run, or Play after Init), which opens the first slot
// (the party menu, exploring), and runs the engine on to its first wait.
func (e *Engine) Start(body func()) {
	body()
	e.advance()
}

// advance runs the engine on to its next wait. With nothing open, the
// party menu starts a game.
func (e *Engine) advance() {
	e.offerPicks()                      // extra: a power a member owes (powers.go)
	if e.over && e.top() == onExplore { // a script's quit, from a continuation
		e.endExplore()
	}
	if e.top() == onNone && !e.ext.intro {
		e.PartyMenu() // ovr17:0161, then ovr04:32F8
	}
}

// Pace is the loop that goes on with time (game.Pace).
type Pace = game.Pace

const (
	PaceNone  = game.PaceNone
	PaceFight = game.PaceFight
	PaceRest  = game.PaceRest
	PaceStory = game.PaceStory
	PaceAnim  = game.PaceAnim
)

// pace is the loop that goes on now.
func (e *Engine) pace() Pace {
	switch {
	case e.waiting(PendFight) && e.cs.actor == core.NoChar:
		return PaceFight
	case e.resting != nil:
		return PaceRest
	case e.waiting(PendStory):
		return PaceStory
	case e.animating():
		return PaceAnim
	}
	return PaceNone
}

// Animate shows the picture's next frame when it is time.
func (e *Engine) Animate() { e.animateWhileWaiting() }

// Step is the call View.Pace names (a headless run's pacing: tests).
func (e *Engine) Step() {
	switch e.pace() {
	case PaceFight:
		e.CombatStep()
	case PaceRest:
		e.RestStep()
	case PaceStory:
		e.StoryStep()
	case PaceAnim:
		e.Animate()
	}
}

// say shows s as the window's note: what a direct action reports instead
// of the original's message line. It is cut to the bottom line (noteCols).
func (e *Engine) say(s string) { e.ext.note, e.ext.noteAt = goldbox.Cut(s, noteCols), time.Now() }
