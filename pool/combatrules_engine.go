package pool

import (
	"errors"
	"fmt"
	"gobox/core"
	"slices"
)

// HandIn makes s the rules' settings from the next decision point on (Do,
// CombatStep). Safe from any goroutine.
func (e *Engine) HandIn(s core.Settings) { e.handedIn.Store(&s) }

// takeSettings makes the settings handed in the rules' own.
func (e *Engine) takeSettings() {
	if s := e.handedIn.Swap(nil); s != nil {
		e.resolve(core.Configure{With: *s})
	}
}

// maneuver takes verbManeuver id for c: an AimWeapon is armed and aims as
// Attack, an AimFoe or AimAlly aims within its range, an AimSelf is used.
func (e *Engine) maneuver(cID core.CharID, id int, walks map[int]core.BoardPos) (done bool) {
	i := -1
	ms := e.maneuvers(cID)
	for j, m := range ms {
		if int(m.ID) == id {
			i = j
		}
	}
	if i < 0 {
		return false
	}
	switch m := ms[i]; m.Aim {
	case core.AimWeapon:
		if m.Move == 0 {
			e.maneuverArm(cID, id)
			return e.attackUnit(cID, 0, walks)
		}
		from := e.at(cID)
		if len(e.squares(cID, m.Move)) == 0 { // nowhere to move: straight to the attack's aim
			e.maneuverAttack(cID, id, from, from)
			return false
		}
		e.maneuverAimMove(cID, id, from)
	case core.AimSelf:
		done, _ = e.maneuverDo(cID, core.Maneuver{ID: core.ManeuverID(id)})
		return done
	case core.AimArea:
		e.ext.aoe = m.Area + 1 // the area preview, as a spell's
		e.playerAim(cID, m.Range, 0, 1, aimThen{kind: atArea, c: cID, id: id})
	default:
		e.ext.enemiesOnly = m.Aim == core.AimFoe
		e.playerAim(cID, m.Range, 1, 0, aimThen{kind: atTarget, c: cID, id: id})
	}
	return false
}

// maneuverDo resolves maneuver a of c's turn through core.Rules; ok false:
// not a target (ErrNotNow).
func (e *Engine) maneuverDo(cID core.CharID, a core.Maneuver) (done, ok bool) {
	e.as(e.sel)
	out, err := e.ruleset().Resolve(core.Fight{Actor: cID}, a, nil)
	if errors.Is(err, core.ErrNotNow) {
		return false, false
	}
	if err != nil {
		panic(fmt.Sprintf("maneuver: %v", err))
	}
	return out.TurnOver, true
}

// maneuverArm arms the weapon maneuver id for c (at: the square aimed to
// move to first, none when nil).
func (e *Engine) maneuverArm(cID core.CharID, id int, at ...core.BoardPos) {
	e.maneuverDo(cID, core.Maneuver{ID: core.ManeuverID(id), Ats: at})
}

// maneuverAttack arms the maneuver and aims its attack (from: where c began,
// to: the square moved to first).
func (e *Engine) maneuverAttack(cID core.CharID, id int, from, to core.BoardPos) {
	e.maneuverArm(cID, id, to)
	e.playerAim(cID, 0xFF, 1, 0, aimThen{kind: atManeuverAttack, c: cID, id: id, from: from}) // as the original: allies too
}

// maneuverAimMove aims the square c moves to first.
func (e *Engine) maneuverAimMove(cID core.CharID, id int, from core.BoardPos) {
	e.say("Aim where to move first (your own square: stay)")
	e.playerAim(cID, e.maneuver1(cID, id).Move, 0, 1, aimThen{kind: atManeuverMove, c: cID, id: id, from: from})
}

// maneuver1 is c's maneuver id.
func (e *Engine) maneuver1(cID core.CharID, id int) core.ManeuverCard {
	for _, m := range e.maneuvers(cID) {
		if int(m.ID) == id {
			return m
		}
	}
	return core.ManeuverCard{}
}

// maneuvers are id's maneuvers, the ruleset's.
func (e *Engine) maneuvers(id core.CharID) core.ManeuverCards {
	return e.fact(id, core.ManeuversFact{}).(core.ManeuverCards)
}

// maneuverAimed goes on with the results of a maneuver's aim.
func (e *Engine) maneuverAimed(th aimThen, found bool, t aimTarget) {
	cID, id := th.c, th.id
	m := e.maneuver1(cID, id)
	switch th.kind {
	case atManeuverMove: // the square first, then the attack's aim
		if !found {
			return
		}
		p := t.At
		if p != th.from && !slices.Contains(e.squares(cID, m.Move), p) {
			e.refuse("Cannot move there")
			e.maneuverAimMove(cID, id, th.from) // again
			return
		}
		e.maneuverAttack(cID, id, th.from, p)
	case atManeuverAttack:
		if !found { // the attack not taken: the move is undone, nothing armed
			e.maneuverDo(cID, core.Maneuver{ID: core.ManeuverID(id), Ats: []core.BoardPos{th.from}, Cancel: true})
		}
		e.later(resume{kind: rBool, ok: found})
	case atArea:
		e.ext.aoe = 0
		end := false
		if found {
			end, _ = e.maneuverDo(cID, core.Maneuver{ID: core.ManeuverID(id), Ats: []core.BoardPos{t.At}, Target: t.C})
		}
		e.later(resume{kind: rClear, ok: end, c: cID})
	case atTarget:
		e.ext.enemiesOnly = false
		end := false
		if found {
			var ok bool
			if end, ok = e.maneuverDo(cID, core.Maneuver{ID: core.ManeuverID(id), Target: t.C}); !ok {
				e.refuse("Not a target")
			}
		}
		e.later(resume{kind: rClear, ok: end, c: cID})
	}
}

// squares are the squares a walk of move squares takes id to.
func (e *Engine) squares(id core.CharID, move int) core.Squares {
	return e.fact(id, core.SquaresFact{Move: move}).(core.Squares)
}
