package present

import (
	"gobox/front"
	"gobox/game"
	"gobox/game/play"
	"image"
	"slices"

	"gobox/combat"
	"gobox/gpu"
	"gobox/sprites"
)

// The adapter between the engine's combat snapshot and gobox/combat.

var combatKinds = map[front.ActKind]combat.Kind{
	front.ActWalk: combat.ActWalk, front.ActAttack: combat.ActAttack, front.ActAim: combat.ActAim, front.ActPoint: combat.ActPoint,
	front.ActClose: combat.ActClose, front.ActYes: combat.ActYes, front.ActNo: combat.ActNo, front.ActWait: combat.ActWait,
	front.ActEnd: combat.ActEnd, front.ActQuick: combat.ActQuick, front.ActGuard: combat.ActGuard, front.ActBandage: combat.ActBandage,
	front.ActFlee: combat.ActFlee, front.ActTurn: combat.ActTurn, front.ActCast: combat.ActCast, front.ActItems: combat.ActItems,
	front.ActRanged: combat.ActRanged, front.ActManeuver: combat.ActManeuver,
	front.ActStep: combat.ActStep, front.ActView: combat.ActView, front.ActOption: combat.ActOption,
}

var engineKinds = func() map[combat.Kind]front.ActKind {
	m := map[combat.Kind]front.ActKind{}
	for e, c := range combatKinds {
		m[c] = e
	}
	return m
}()

// syncBattle gives the tabletop its theme and Shift before it draws or takes a key.
func (p *Presenter) syncBattle() {
	p.battle.Theme, p.battle.Shift, p.battle.Alt = combat.DefaultTheme, gpu.IsKeyPressed(gpu.KeyShift), gpu.IsKeyPressed(gpu.KeyAlt)
	p.battle.Theme.Series = series
	p.battle.Meshes, p.battle.Props = sprites.Meshes(series), sprites.Combat(series)
}

// toAction is what the player chose as the engine's action.
func toAction(d combat.Decision) front.Action {
	return front.Action{Kind: engineKinds[d.Kind], N: d.N, X: d.X, Y: d.Y}
}

// keypadSteps are the keypad's directions (ActDir N) by square offset.
var keypadSteps = map[image.Point]int{{0, -1}: 8, {1, -1}: 9, {1, 0}: 6, {1, 1}: 3, {0, 1}: 2, {-1, 1}: 1, {-1, 0}: 4, {-1, -1}: 7}

// askedTurn: the question up is a turn the tabletop answers (AskSquare: an
// aim, whatever Turn the rules publish; AskMove: a step at a time, where they
// publish none).
func (p *Presenter) askedTurn() bool {
	q := p.askUp()
	return p.board != nil && q != nil && (q.Kind == front.AskSquare || q.Kind == front.AskMove && p.board.Turn == nil)
}

// askedBoard is b with the active unit's turn the question asks for
// (askedTurn): an aim, or a move whose reach is the open squares next to the
// unit and whose End ends it. c is the fight b is of.
func (p *Presenter) askedBoard(b *combat.Board, c *play.Combat) *combat.Board {
	if !p.askedTurn() {
		return b
	}
	o := *b
	aim := p.askUp().Kind == front.AskSquare
	t := &combat.Turn{Aim: aim, Can: []combat.Kind{combat.ActAim}, CastDirect: true}
	if pt := c.Turn; pt != nil { // the game's turn, now aiming (a spell's square): its casting only
		t.Casting = pt.Spell
	}
	var at image.Point
	for _, u := range o.Units {
		if u.Active {
			t.Unit, at = u.ID, image.Pt(u.X, u.Y)
		}
	}
	if !aim {
		t.Can, o.Reach, o.Paths = []combat.Kind{combat.ActEnd}, nil, nil
		for d := range keypadSteps {
			if q := at.Add(d); q.In(image.Rect(0, 0, len(o.Ground[0]), len(o.Ground))) && !o.Ground[q.Y][q.X].Wall {
				o.Reach = append(o.Reach, q)
			}
		}
	}
	o.Turn = t
	return &o
}

// askAction is the tabletop's decision d as the answer to the question up:
// on a move (askedTurn) a square walked to is that step (ActDir), End is
// done (ActOK); otherwise toAction's.
func (p *Presenter) askAction(d combat.Decision) front.Action {
	if !p.askedTurn() || p.askUp().Kind != front.AskMove {
		return toAction(d)
	}
	switch d.Kind {
	case combat.ActEnd:
		return front.Action{Kind: front.ActOK}
	case combat.ActStep:
		return front.Action{Kind: front.ActDir, N: d.N}
	case combat.ActWalk:
		for _, u := range p.board.Units {
			if n, ok := keypadSteps[image.Pt(d.X-u.At.X, d.Y-u.At.Y)]; u.Active && ok {
				return front.Action{Kind: front.ActDir, N: n}
			}
		}
		return front.Action{}
	}
	return toAction(d)
}

// keepFallen is c with the units of the fight drawn before the moments now
// playing that c has dropped, each kept until its own fall (EvDown) plays:
// the blows land on figures, not on empty squares. Outside the replay c is
// the fight kept for the next one.
func (p *Presenter) keepFallen(c *play.Combat) *play.Combat {
	if c == nil || !p.Playing() {
		p.lastBoard = c
		return c
	}
	last := p.lastBoard
	if last == nil {
		return c
	}
	o := *c
	o.Units = slices.Clone(c.Units)
	for _, u := range last.Units {
		if u.Gone || !slices.ContainsFunc(p.replay.q[1:], func(e frame) bool { return e.Kind == game.EvDown && e.Unit == u.ID }) { // the fall showing now (q[0]) is played: the unit is down
			continue
		}
		if i := slices.IndexFunc(o.Units, func(x game.Unit) bool { return x.ID == u.ID }); i < 0 {
			o.Units = append(o.Units, u)
		} else if o.Units[i].Gone {
			o.Units[i] = u
		}
	}
	return &o
}
