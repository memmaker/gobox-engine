package pool

import (
	"errors"
	"image"
	"slices"

	"gobox/core"
	"gobox/core/grid"
	"gobox/game"
	"gobox/game/play"
)

// A question a play action asks before it is taken (a fight's Flee? or
// Attack Ally?, the spell or member a cast or an item use goes on, the
// treasure's "sure?") is data: the replies and, per reply, the call it
// makes. The next action, a Reply, makes it.

// callKind is what a pendingCall makes.
type callKind int

const (
	callNone       callKind = iota // nothing (a No)
	callTurn                       // the fight's turn actions, the first sure
	callAimClose                   // a fight's No: an aim it asked over gives up (ovr14:3212)
	callItem                       // UseItem
	callPower                      // UsePower
	callLoot                       // LootLeave
	callDrop                       // dropping a member, target the questions still to ask
	callClass                      // NewClass, target the class
	callCastChoice                 // the reply to a spell's own question (SpellAim.Choose): the waiting cast goes off
)

// askOf is what a pending question picks: a menu entry, a spell, a member.
type askOf int

const (
	askMenu askOf = iota
	askSpell
	askMember
)

// pendingCall is one rule call with its arguments.
type pendingCall struct {
	kind                  callKind
	acts                  []TurnAct // callTurn
	member, spell, target int
	item                  core.ItemHandle
	class                 core.Class // callClass
	power                 game.PowerKind
	sure                  bool
	choice                int32 // callCastChoice: the reply's index
}

// playQuestion is the question up and the call each reply makes.
type playQuestion struct {
	q     play.Question
	of    askOf
	calls []pendingCall
}

// errAsked: the call asked its question, which is now the pending one
// (e.ext.pq); its reply makes the call again. runCall takes it as done.
var errAsked = errors.New("asked")

// runCall makes p; a question it asks becomes the pending one.
func (e *Engine) runCall(p pendingCall) error {
	if err := e.makeCall(p); err != errAsked {
		return err
	}
	return nil
}

// makeCall makes p; errAsked: it asked.
func (e *Engine) makeCall(p pendingCall) error {
	switch p.kind {
	case callTurn:
		return e.turnActs(p.acts, p.sure)
	case callAimClose:
		if e.cs.aim != nil {
			return e.turnActs([]TurnAct{{Verb: VerbClose}}, false)
		}
	case callItem:
		return e.withItems(p.member, func() error { return e.UseItem(p.item, p.spell, p.target, true) })
	case callPower:
		return e.UsePower(p.member, p.power, p.spell, p.target, p.sure)
	case callLoot:
		return e.LootLeave(p.sure)
	case callClass:
		return e.newClass(p.member, p.class)
	case callCastChoice:
		e.castChosen(p.choice)
	case callDrop:
		if p.target > 0 { // ovr17:1F93: asked again
			e.ext.pq = &playQuestion{q: play.Question{Text: "Are you sure? ", YesNo: true, Replies: []string{"No", "Yes"}, Default: 1},
				calls: []pendingCall{{}, {kind: callDrop, member: p.member}}}
			return nil
		}
		e.selectMember(p.member)
		return e.DropMember(p.member)
	}
	return nil
}

// askSure asks text, yes/no, before call p: Yes makes it again sure
// (the treasure's question the other way round: Yes goes back to it).
func (e *Engine) askSure(text string, p pendingCall) error {
	yes, no := p, pendingCall{}
	yes.sure = true
	switch p.kind {
	case callLoot: // ovr06:10B8: Yes goes back to the treasure, No leaves it
		yes, no = pendingCall{}, yes
	case callTurn:
		no = pendingCall{kind: callAimClose}
	}
	q := play.Question{Text: text, YesNo: true, Replies: []string{"No", "Yes"}, Default: 1}
	if p.kind == callTurn {
		q.NoEsc, q.NoPick = true, true // Flee:, Attack Ally?: ovr26:10D1, Y/N only
	}
	e.ext.pq = &playQuestion{q: q, calls: []pendingCall{no, yes}}
	return errAsked
}

// askSpell asks which of spells call p goes on with.
func (e *Engine) askSpell(spells []Spell, p pendingCall) error {
	pq := &playQuestion{q: play.Question{Text: "Which spell?"}, of: askSpell}
	for _, sp := range spells {
		c := p
		c.spell = sp.ID
		pq.q.Replies, pq.calls = append(pq.q.Replies, sp.Name), append(pq.calls, c)
	}
	e.ext.pq = pq
	return errAsked
}

// askTarget asks text: the member call p goes on.
func (e *Engine) askTarget(text string, p pendingCall) error {
	pq := &playQuestion{q: play.Question{Text: text, List: true}, of: askMember}
	for i, id := range e.partyIDs() {
		c := p
		c.target = i
		pq.q.Replies, pq.calls = append(pq.q.Replies, e.name(id)), append(pq.calls, c)
	}
	e.ext.pq = pq
	return errAsked
}

// askClass asks member n's new class among the rules' offers (NewClass, "Pick New Class:").
func (e *Engine) askClass(offers []core.Offer, n int) error {
	pq := &playQuestion{q: play.Question{Text: "Pick New Class:"}}
	for _, o := range offers {
		pq.q.Replies, pq.calls = append(pq.q.Replies, o.Name), append(pq.calls, pendingCall{kind: callClass, member: n, class: o.Action.(core.ChangeClass).To})
	}
	e.ext.pq = pq
	return errAsked
}

// turnActs gives as to the acting member's turn in order, the first sure;
// a question the turn asks becomes the pending one, its Yes the rest sure.
func (e *Engine) turnActs(as []TurnAct, sure bool) error {
	if !e.battleUp() {
		return core.ErrNotNow
	}
	for i, x := range as {
		if err := e.TurnChoice(x, sure && i == 0); err != nil {
			return err
		}
		if q := e.cs.ask; q != "" { // Flee:, Attack Ally?: Yes makes the rest again, sure
			e.cs.ask = ""
			e.askSure(q, pendingCall{kind: callTurn, acts: slices.Clone(as[i:])})
			return nil
		}
	}
	return nil
}

// replyPending answers the pending question with reply n.
func (e *Engine) replyPending(n int) error {
	pq := e.ext.pq
	if n < 0 || n >= len(pq.calls) {
		return &play.Refusal{Text: "no such option"}
	}
	e.ext.pq = nil
	return e.runCall(pq.calls[n])
}

// leavePending backs out of the pending question: a yes/no's No, a pick
// given up.
func (e *Engine) leavePending() error {
	if e.ext.pq.q.YesNo {
		return e.replyPending(0)
	}
	e.ext.pq = nil
	return nil
}

// withItems runs f with member n's items up, as the engine's item calls
// take them (the items screen is gobox's): opened for f and closed after
// when none was up.
func (e *Engine) withItems(n int, f func() error) error {
	opened := false
	if !e.view.on && !e.snap.invOn {
		if err := e.OpenItems(n, true); err != nil {
			return err
		}
		opened = true
	}
	if err := e.ShowMember(n); err != nil {
		if opened {
			e.CloseItems(true)
		}
		return err
	}
	err := f()
	if opened && e.snap.invOn {
		e.CloseItems(true)
	}
	return err
}

// moveActs are a fight's step to to as turn actions: a step off the map
// flees, into a foe attacks it, into an ally aims at it (which asks
// Attack Ally?, BUG-CLASSES 41), as moveAttack's (ovr09:09DF).
func (e *Engine) moveActs(to image.Point, reach []game.Reach) ([]TurnAct, error) {
	if c := e.cs.actor; c != core.NoChar && e.battleUp() {
		at := e.fact(c, core.AtFact{}).(core.BoardPos)
		for dir := 0; dir < 8; dir++ {
			if to != image.Pt(at.X+grid.StepX[dir], at.Y+grid.StepY[dir]) {
				continue
			}
			st := e.fact(c, core.StepFact{Dir: dir}).(core.StepInfo)
			switch {
			case st.Unit == 0 && st.Ground == 0:
				return []TurnAct{{Verb: VerbFlee}}, nil
			case st.Foe:
				return []TurnAct{{Verb: VerbAttack, N: int(st.Unit)}}, nil
			case st.Unit > 0:
				as := []TurnAct{{Verb: VerbAim, X: to.X, Y: to.Y}}
				if e.cs.aim == nil { // a No left the aim up: it asks again
					as = append([]TurnAct{{Verb: VerbAttack}}, as...)
				}
				return as, nil
			case !slices.ContainsFunc(reach, func(q game.Reach) bool { return q.At == to }):
				return nil, &play.Refusal{Text: "Blocked"} // ovr09:0C6B, the string at 09D7
			}
		}
	}
	return []TurnAct{{Verb: VerbWalk, X: to.X, Y: to.Y}}, nil
}
