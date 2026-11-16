package engine

import (
	"testing"

	"gobox/game"
	"gobox/rules"
)

// br1FarFacing: the party faces east, where the start square has free squares; SetupMonster's distance is the free
// squares ahead (at most the one asked), so a party facing a wall starts the fight with the monsters adjacent.
const br1FarFacing = 2

// br1FarFight is block 201's body with one Hexadillo (monster 5: it walks up to a target, buckrogers-go engaged_test)
// set away from the party (BR1 block 16: SetupMonster(32, distance=2, closeup=32, 0)).
const br1FarFight = "    ClearMonsters()\n    SetupMonster(5, distance=2, closeup=32, 0)\n    LoadMonster(5, count=1, icon=32)\n    Combat()"

// br1Step is one choice the played fight asked for: the unit's turn as the state showed it, and what was chosen.
type br1Step struct {
	Round, Unit, Moves int
	At                 rules.BoardPos
	Reach              int // the squares Allowed.Moves offered
	Foe                int // squares to the nearest living foe (Chebyshev)
	Did                game.Action
}

// br1Near is the nearest living foe of the turn's unit in squares (Chebyshev), and where the unit stands.
func br1Near(f game.Fighting) (me rules.BoardPos, near int) {
	near = 1 << 30
	for _, u := range f.Field.Units {
		if u.ID == f.Field.Turn {
			me = u.At
		}
	}
	for _, u := range f.Field.Units {
		if u.Foe && !u.Dead && !u.Gone {
			near = min(near, max(rules.Abs(me.X-u.At.X), rules.Abs(me.Y-u.At.Y)))
		}
	}
	return me, near
}

// br1Walker is the walk-up: one step at a time toward the nearest foe (a square one away from the unit that is
// nearer), the Attack the state offers once beside it, EndTurn when no step is left.
func br1Walker(f game.Fighting) game.Action {
	me, near := br1Near(f)
	if near <= 1 {
		for _, a := range f.Allowed.Plain {
			if _, ok := a.(game.Attack); ok {
				return a
			}
		}
		return game.EndTurn{}
	}
	var foes []rules.BoardPos
	for _, u := range f.Field.Units {
		if u.Foe && !u.Dead && !u.Gone {
			foes = append(foes, u.At)
		}
	}
	dist := func(p rules.BoardPos) int {
		d := 1 << 30
		for _, q := range foes {
			d = min(d, max(rules.Abs(p.X-q.X), rules.Abs(p.Y-q.Y)))
		}
		return d
	}
	best, bestD := rules.BoardPos{}, near
	for _, m := range f.Allowed.Moves {
		if max(rules.Abs(m.At.X-me.X), rules.Abs(m.At.Y-me.Y)) == 1 && dist(m.At) < bestD {
			best, bestD = m.At, dist(m.At)
		}
	}
	if bestD < near {
		return game.MoveTo{To: best}
	}
	return game.EndTurn{}
}

// br1Walked plays the far fight with the walker, recording each choice and the board at each of the party's turns.
func br1Walked(t *testing.T) (steps []br1Step, trace []string, e *Engine) {
	t.Helper()
	e, trace, stop := br1FoughtIn(t, br1FarFight, br1FarFacing, func(f game.Fighting) game.Action {
		a := br1Walker(f)
		me, near := br1Near(f)
		steps = append(steps, br1Step{Round: f.Field.Round, Unit: f.Field.Turn, Moves: f.Field.Moves, At: me, Reach: len(f.Allowed.Moves), Foe: near, Did: a})
		return a
	})
	if stop != "" {
		t.Fatalf("the fight did not play to its end: %s\ntrace: %q", stop, trace)
	}
	return steps, trace, e
}

// TestBR1WalkUp: the monster starts apart; a member's Move offers squares, each step spends movement and a turn
// takes several steps until it is used; the fight closes to melee and ends.
func TestBR1WalkUp(t *testing.T) {
	steps, trace, e := br1Walked(t)
	if len(steps) == 0 {
		t.Fatal("no turns")
	}
	if steps[0].Foe < 3 {
		t.Fatalf("the units start %d apart, want apart (>2)", steps[0].Foe)
	}
	if steps[0].Reach == 0 {
		t.Error("the first turn's Move offers no square")
	}
	inTurn := 0 // steps taken in one turn, the most
	run := 0
	for i, s := range steps {
		if _, ok := s.Did.(game.MoveTo); !ok {
			run = 0
			continue
		}
		run++
		inTurn = max(inTurn, run)
		if i+1 < len(steps) && steps[i+1].Unit == s.Unit && steps[i+1].Round == s.Round {
			if steps[i+1].Moves >= s.Moves {
				t.Errorf("a step cost nothing: %+v then %+v", s, steps[i+1])
			}
			if steps[i+1].Foe >= s.Foe {
				t.Errorf("a step did not close in: %+v then %+v", s, steps[i+1])
			}
		}
	}
	if inTurn < 2 {
		t.Errorf("at most %d steps in a turn, want several until the movement is used: %+v", inTurn, steps)
	}
	last := steps[len(steps)-1]
	if _, ok := last.Did.(game.Attack); !ok {
		t.Errorf("the last choice %T, want an attack beside the foe", last.Did)
	}
	if _, ok := e.build().(game.Fighting); ok { // won or lost (a loss leaves the monster in the party list)
		t.Error("still Fighting after the play")
	}
	if len(trace) < 5 {
		t.Errorf("the fight left %d turns: %q", len(trace), trace)
	}
}

// br1FoeDist is the squares between the nearest monster and the nearest active party member.
func br1FoeDist(f game.Fighting) int {
	d := 1 << 30
	for _, m := range f.Field.Units {
		if !m.Foe || m.Dead {
			continue
		}
		for _, p := range f.Field.Units {
			if !p.Foe && !p.Dead && p.Active {
				d = min(d, max(rules.Abs(m.At.X-p.At.X), rules.Abs(m.At.Y-p.At.Y)))
			}
		}
	}
	return d
}

// TestBR1Approach: the party only ends its turns; the monster closes in over its own turns, round after round,
// toward the party, until it is beside it.
func TestBR1Approach(t *testing.T) {
	var dists []int
	br1FoughtIn(t, br1FarFight, br1FarFacing, func(f game.Fighting) game.Action {
		if f.Field.Round > len(dists) { // the first turn of each round
			dists = append(dists, br1FoeDist(f))
		}
		if f.Field.Round > 8 { // the play only ends a fight it has won
			panic("approach stopped")
		}
		return game.EndTurn{}
	})
	if len(dists) < 3 || dists[0] < 3 {
		t.Fatalf("distances at each round's start: %v", dists)
	}
	for i := 1; i < len(dists); i++ {
		if dists[i] > dists[i-1] {
			t.Errorf("the monster moved away: %v", dists)
		}
	}
	if dists[len(dists)-1] >= dists[0] {
		t.Errorf("the monster did not close in: %v", dists)
	}
	t.Logf("distance at each round's start: %v", dists)
}
