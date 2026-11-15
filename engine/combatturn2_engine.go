package engine

import (
	"fmt"
	"gobox/game"
	"gobox/rules"
	"slices"
)

// aimConfirm = ovr14:2C3D (c, cand, mode): take cand as the target. With
// mode 1 (a missile attack) the attack is made at once, after "Attack
// Ally?" (mayAttack) for a friend; the results are found and the target.
func (e *Engine) aimConfirm(cID, candID rules.CharID, mode int, ranged bool) (found bool, t aimTarget) {
	t = aimTarget{C: candID, At: e.at(candID)}
	found = true
	if mode == 1 { // the blow is core.Rules Resolve(Attack); an ally asks first
		var a rules.Action = rules.Attack{Target: candID, Ranged: ranged}
		if e.cs.sure { // the call asks first (cs.ask: turnActs makes it the pending one)
			a = rules.Confirmed{Action: a}
		}
		e.as(e.sel)
		out, err := e.ruleset().Resolve(e.fightBy(cID), a, nil)
		if err != nil {
			panic(fmt.Sprintf("attack: %v", err))
		}
		if out.Confirm != "" {
			e.cs.ask = out.Confirm
			return false, aimTarget{}
		}
		found = out.TurnOver
	}
	if !found {
		t = aimTarget{}
	}
	return
}

// aimTarget is an aim's result: the square, and who stood there (NoChar: nobody).
type aimTarget struct {
	C  rules.CharID
	At rules.BoardPos
}

// playerAim = ovr14:3558 (c, rng, showRange, allowGround, var found, var t):
// the player aims for c. extra: through Apply: verbPoint moves the aim (the
// View shows why it cannot be taken, a spell's area), verbAim takes a
// square, verbClose gives up. Battle.Targets are the foes to cycle through
// (with a weapon: the ones it can hit, while there are any): the rules'
// AimFact. rng 0xFF takes the range of the readied weapon. The aim is a
// phase of the turn (cs.aim): then gets the results once it is gone (aimFinished).
func (e *Engine) playerAim(cID rules.CharID, rangeArg, showRange, allowGround int, then aimThen) {
	e.as(e.sel)
	o := e.fact(cID, rules.AimFact{Range: rangeArg, ShowRange: showRange, FoesOnly: e.ext.enemiesOnly, Ranged: then.ranged}).(rules.AimOffer)
	if len(o.Near) == 0 { // extra: nobody to aim at
		e.aimFinished(then, false, aimTarget{})
		return
	}
	h := &aimRun{e: e, c: cID, rng: o.Range, showRange: showRange, allowGround: allowGround, list: o.Near, then: then}
	h.turn, h.targets, h.hits = e.snap.turn, e.snap.targets, e.snap.hits
	e.snap.targets, e.snap.hits = nil, map[int]int{}
	for i, f := range o.Targets {
		n := e.unit(f)
		e.snap.targets = append(e.snap.targets, n)
		if o.Hits[i] >= 0 {
			e.snap.hits[n] = o.Hits[i]
		}
	}
	e.snap.turn = &Turn{Unit: e.unit(cID), Aim: true}
	if e.ext.aoeSpell != 0 { // aimSpell's: which spell is being aimed
		e.snap.turn.Casting = e.fact(cID, rules.SpellAimFact{Spell: e.spellID(e.ext.aoeSpell)}).(rules.SpellAim).Name
	}
	at := cID
	if len(o.Targets) > 0 {
		at = o.Targets[0]
	}
	p := e.at(at)
	h.point(p.X, p.Y)
	e.cs.aim = h
}

// aimRun is the player aiming (playerAim): the aim is a phase of the turn
// (cs.aim); then says what the results go on with.
type aimRun struct {
	e                      *Engine
	c                      rules.CharID
	rng                    int
	showRange, allowGround int
	list                   []rules.CharID // those in sight (AimOffer.Near)
	t                      aimTarget
	found, finished        bool  // finished: the target taken after the question over the aim
	marked                 bool  // the area preview's marks are up
	turn                   *Turn // the turn's menu, put back when the aim is gone
	targets                []int
	hits                   map[int]int
	then                   aimThen
}

// check is the rules' word on the aim at at: who would be the target and why not.
func (h *aimRun) check(at rules.BoardPos) rules.AimCheck {
	return h.e.fact(h.c, rules.AimCheckFact{Near: h.list, At: at, Range: h.rng, ShowRange: h.showRange, Ground: h.allowGround != 0, Ranged: h.then.ranged}).(rules.AimCheck)
}

// squares is the squares of b the aim takes (check): what gobox offers.
func (h *aimRun) squares(b *game.Combat) []rules.BoardPos {
	return slices.DeleteFunc(inRange(b, h.rng), func(at rules.BoardPos) bool { return h.check(at).Why != "" })
}

// point moves the aim to (x, y): who would be the target and why not.
func (h *aimRun) point(x, y int) rules.AimCheck {
	e := h.e
	if h.marked {
		e.aoeUnmark()
	}
	e.as(e.sel)
	a := h.check(rules.BoardPos{X: x, Y: y})
	why := a.Why
	e.snap.aimWhy = why
	e.aoeMarks(x, y)
	h.marked = e.ext.aoe != 0
	return a
}

// apply is aiming's decision: a square pointed at, taken, or Close; true
// when the aim is over.
func (h *aimRun) apply(a TurnAct) bool {
	e := h.e
	if h.finished || a.Verb == VerbClose { // finished: the target taken after the question over the aim
		return true
	}
	if a.Verb != VerbPoint && a.Verb != VerbAim || a.X < 0 || a.Y < 0 || a.X > 0x31 || a.Y > 0x18 {
		return false
	}
	ck := h.point(a.X, a.Y)
	if a.Verb == VerbPoint {
		return false
	}
	if ck.Why != "" {
		e.refuse(ck.Why)
		return false
	}
	h.t = aimTarget{C: ck.Target, At: rules.BoardPos{X: a.X, Y: a.Y}}
	if h.showRange == 1 || ck.Target != rules.NoChar && ck.Occupied {
		f, at := e.aimConfirm(h.c, ck.Target, h.showRange, h.then.ranged) // ovr14:2C3D: the target's own square; a weapon strikes
		h.found, h.finished = f, f
		if f {
			h.t = at
		} else {
			h.t = aimTarget{}
		}
		return h.finished
	}
	h.found = true
	return true
}

// leave puts the turn's menu back, and the results go on.
func (h *aimRun) leave() {
	e := h.e
	if h.marked {
		e.aoeUnmark()
	}
	e.snap.turn, e.snap.targets, e.snap.hits, e.snap.aimWhy = h.turn, h.targets, h.hits, ""
	e.aimFinished(h.then, h.found, h.t)
}
