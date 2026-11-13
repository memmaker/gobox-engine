package ui

import (
	"gobox/game"
	"slices"
)

// The presenter paces what goes on with time (Pacing: the fight's
// turns, camp's rest, the finale, the auto-walk; a picture's frames
// animate in the draw, from the View's FrameMs, no call) and asks the
// yes/no questions the rules need answered (Flee?, Attack Ally?, Stop
// resting?) itself.

// Pace queues the call that goes on with time now; nothing while one of
// its questions waits.
func (p *Presenter) Pace() {
	if p.q != nil || p.board != nil && p.board.Turn != nil { // a member's turn waits on TurnChoice: nothing goes on with time (a rest the fight broke into never answers it)
		return
	}
	switch p.pacing().Pace {
	case game.PaceFight:
		p.call(func(e *classicGame) { e.CombatStep() })
	case game.PaceRest:
		p.call(func(e *classicGame) { e.RestStep() })
	case game.PaceStory:
		p.call(func(e *classicGame) { e.StoryStep() })
	case game.PaceWalk:
		p.call(func(e *classicGame) { e.WalkStep() })
	}
}

// turnActs are the decisions of a member's combat turn (TurnChoice).
var turnActs = []ActKind{ActWalk, ActAttack, ActItems, ActMember, ActRanged,
	ActManeuver, ActCast, ActTurn, ActQuick, ActGuard, ActBandage,
	ActWait, ActEnd, ActFlee, ActPoint, ActAim, ActClose, ActOption}

// turnAct makes a decision of the turn up a rule call; false: no turn waits.
func (p *Presenter) turnAct(a Action) bool {
	if p.board == nil || p.board.Turn == nil || p.inventoryUp() || p.sheetShown() || !slices.Contains(turnActs, a.Kind) {
		return false
	}
	p.turnCall(a, false)
	return true
}

// turnCall calls TurnChoice; a question it is refused with is asked, Yes
// calls it again sure.
func (p *Presenter) turnCall(a Action, sure bool) {
	p.call(func(e *classicGame) {
		AskSure(p, e.TurnChoice(a, sure), func() { p.turnCall(a, true) })
	})
}
