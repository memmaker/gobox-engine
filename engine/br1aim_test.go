package engine

import (
	"slices"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// br1AimTurn is a party turn of the far fight (one Hexadillo out in the open), with the aim cursor up as the bar's
// Aim opens it (attackUnit with no target: ovr09 combatMenu's Aim).
func br1AimTurn(t *testing.T) (e *Engine, f game.Fighting, me, foe game.Unit) {
	t.Helper()
	e, f = br1AtTurn(t, br1FarFight, br1FarFacing)
	for _, u := range f.Field.Units {
		switch {
		case u.ID == f.Field.Turn:
			me = u
		case u.Foe:
			foe = u
		}
	}
	if err := e.TurnChoice(TurnAct{Verb: VerbAttack}, false); err != nil || e.cs.aim == nil {
		t.Fatalf("Aim opened no cursor: %v", err)
	}
	return e, f, me, foe
}

// br1Rounds is a unit's ammunition: the readied weapon's Count.
func br1Ammo(e *Engine, f game.Fighting) int32 {
	cID := e.unitID(f.Field.Turn)
	h := e.fact(cID, rules.WeaponFact{}).(rules.Handles)
	if len(h) == 0 {
		return -1
	}
	return e.invItem(cID, h[0]).Count
}

func br1HP(e *Engine, id int) int {
	for _, u := range e.build().(game.Fighting).Field.Units {
		if u.ID == id {
			return u.HP
		}
	}
	return -1
}

// TestBR1Aim: a member with a gun aims through the engine's cursor (the bar's Aim: AimFact, AimCheckFact): the foe
// nine squares off is offered with its hit chance and accepted; a square past the gun's range (16) and a wall are
// refused with ManualAimCursor's word ("Exit"), the aim staying up; taking the foe fires: the shot is resolved (a hit
// takes the damage the strike shows), the gun's shots go down (MakeAttack spends a shot per attack) and the
// turn is used.
func TestBR1Aim(t *testing.T) {
	e, f, me, foe := br1AimTurn(t)
	if got := e.cs.aim.rng; got != 16 {
		t.Errorf("aim range %d, want the Gun Bolt's 16", got)
	}
	if len(e.snap.targets) != 1 || e.snap.targets[0] != foe.ID {
		t.Fatalf("aim targets %v, want the foe %d", e.snap.targets, foe.ID)
	}
	if h, ok := e.snap.hits[foe.ID]; !ok || h < 5 || h > 95 {
		t.Errorf("hit chance %d (%v), want one in 5..95", h, ok)
	}
	far := rules.BoardPos{X: me.At.X + 22, Y: me.At.Y}
	var wall rules.BoardPos
	for _, g := range f.Field.Squares {
		if g.Block && max(rules.Abs(g.At.X-me.At.X), rules.Abs(g.At.Y-me.At.Y)) < 10 {
			wall = g.At
			break
		}
	}
	for _, c := range []struct {
		what string
		at   rules.BoardPos
		why  string
	}{{"foe in range", foe.At, ""}, {"past the range", far, "Exit"}, {"a wall", wall, "Exit"}} {
		if c.what == "a wall" && wall == (rules.BoardPos{}) {
			t.Fatal("no wall square near the unit")
		}
		e.cs.aim.apply(TurnAct{Verb: VerbPoint, X: c.at.X, Y: c.at.Y})
		if got := e.fightState().Why; got != c.why {
			t.Errorf("%s at %v: why %q, want %q", c.what, c.at, got, c.why)
		}
	}
	before, ammo := br1HP(e, foe.ID), br1Ammo(e, f)
	seq := e.fightState().Strike.Seq
	e.cs.aim.apply(TurnAct{Verb: VerbAim, X: far.X, Y: far.Y}) // refused: the aim stays
	if e.cs.aim == nil || e.fightState().Strike.Seq != seq {
		t.Fatal("a shot past the range was taken")
	}
	if err := e.TurnChoice(TurnAct{Verb: VerbAim, X: foe.At.X, Y: foe.At.Y}, false); err != nil {
		t.Fatal(err)
	}
	br1AssertShot(t, e, f, me, foe)
	st := e.fightState().Strike
	// one Aim is the round's attacks at that target (combat.pas CreatureAction: the swings start anew each turn),
	// so more than one strike may follow; the last one is shown
	if st.Seq <= seq || st.From != me.ID || st.To != foe.ID {
		t.Fatalf("strike %+v after the shot, want %d to %d, seq above %d", st, me.ID, foe.ID, seq)
	}
	if got := br1HP(e, foe.ID); got > before-st.Damage || st.Hit != (st.Damage > 0) && st.Hit {
		t.Errorf("hit points %d -> %d for strike %+v", before, got, st)
	}
	if got := br1Ammo(e, f); got >= ammo || got < 0 {
		t.Errorf("gun shots %d -> %d, want some spent (one per attack of the round)", ammo, got)
	}
	if e.cs.aim != nil || e.cs.turn != nil && e.cs.actor == e.unitID(me.ID) {
		t.Error("the shot did not use the turn")
	}
}

// br1AssertShot: the shot published its flight (game.Shot) once, from the member's square to the foe's, in the
// picture of AnimateMissile's art for the weapon's kind (the Gun Bolt: FlightGunBolt, COMSPR's icons 22..24 in some
// direction and pose), at the original's pace.
func br1AssertShot(t *testing.T, e *Engine, f game.Fighting, me, foe game.Unit) {
	t.Helper()
	h := e.fact(e.unitID(me.ID), rules.WeaponFact{}).(rules.Handles)
	kind := e.fact(rules.NoChar, rules.FlightFact{Item: e.invItem(e.unitID(me.ID), h[0]).Kind}).(rules.Flight).Kind
	if kind != rules.FlightGunBolt {
		t.Errorf("the Gun Bolt flies as kind %d, want FlightGunBolt", kind)
	}
	art, _ := e.missile(false, kind)
	var shots []game.Shot
	for _, ev := range e.events {
		if s, ok := ev.(game.Shot); ok {
			shots = append(shots, s)
		}
	}
	if len(shots) == 0 {
		t.Fatal("the shot published no game.Shot")
	}
	s := shots[0]
	if s.To != foe.At || s.StepMs != art.Pace || art.Pace != brMissilePaceWant {
		t.Errorf("shot to %v at %d ms, want %v at %d", s.To, s.StepMs, foe.At, brMissilePaceWant)
	}
	var want [][]byte
	for k := range 3 { // the three icons the directions use: Animate(22)'s 22, 23, 24
		for pose := range 2 {
			for _, flip := range []bool{false, true} {
				if im := e.icons[art.Slot+k][pose]; im != nil {
					want = append(want, e.picPix(im, 0, flip, nil).Pix)
				}
			}
		}
	}
	if s.Art.W != 24 || s.Art.H != 24 || !slices.ContainsFunc(want, func(p []byte) bool { return slices.Equal(p, s.Art.Pixels) }) {
		t.Errorf("shot art %dx%d is none of COMSPR's icons %d..%d", s.Art.W, s.Art.H, art.Slot, art.Slot+2)
	}
}

// brMissilePaceWant is AnimateMissile's pace (cbtacts.pas local.pace := 10).
const brMissilePaceWant = 10
