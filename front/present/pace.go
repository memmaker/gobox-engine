package present

import (
	"gobox/front"
	"slices"
)

// The presenter paces what goes on with time (View.Pace: the fight's
// turns, camp's rest, the finale, the auto-walk; a picture's frames
// animate in the draw, from the View's FrameMs, no call) and asks the
// yes/no questions the rules need answered (Flee?, Attack Ally?, Stop
// resting?) itself.

// Pace queues the call that goes on with time now; nothing while one of
// its questions waits.
func (p *Presenter) Pace(v front.View) {
	if p.q != nil || v.Battle != nil && v.Battle.Turn != nil { // a member's turn waits on TurnChoice: nothing goes on with time (a rest the fight broke into never answers it)
		return
	}
	switch v.Pace {
	case front.PaceFight:
		p.call(func(e front.Game) { e.CombatStep() })
	case front.PaceRest:
		p.call(func(e front.Game) { e.RestStep() })
	case front.PaceStory:
		p.call(func(e front.Game) { e.StoryStep() })
	case front.PaceWalk:
		p.call(func(e front.Game) { e.WalkStep() })
	}
}

// turnActs are the decisions of a member's combat turn (TurnChoice).
var turnActs = []front.ActKind{front.ActWalk, front.ActAttack, front.ActItems, front.ActMember, front.ActRanged,
	front.ActManeuver, front.ActCast, front.ActTurn, front.ActQuick, front.ActGuard, front.ActBandage,
	front.ActWait, front.ActEnd, front.ActFlee, front.ActPoint, front.ActAim, front.ActClose, front.ActOption}

// turnAct makes a decision of the turn up a rule call; false: no turn waits.
func (p *Presenter) turnAct(a front.Action) bool {
	v := p.shownView
	if v.Battle == nil || v.Battle.Turn == nil || v.Inventory != nil || v.Sheet != nil || !slices.Contains(turnActs, a.Kind) {
		return false
	}
	p.turnCall(a, false)
	return true
}

// turnCall calls TurnChoice; a question it is refused with is asked, Yes
// calls it again sure.
func (p *Presenter) turnCall(a front.Action, sure bool) {
	p.call(func(e front.Game) {
		front.AskSure(p, e.TurnChoice(a, sure), func() { p.turnCall(a, true) })
	})
}
