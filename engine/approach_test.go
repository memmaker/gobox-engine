package engine

import (
	"testing"

	"gobox/game"
)

// TestApproachOut: on a party turn, a target the melee Attack walks up to
// carries that walk (Unit.Approach) out to the front, so hovering it shows
// the path (lost in 71e314fe, when the front built Paths from Reach alone).
func TestApproachOut(t *testing.T) {
	e := partyTurnScene(t)
	walks := 0
	f := fighting(e)
	for _, u := range f.Field.Units {
		if len(u.Squares) == 0 || !u.Foe {
			continue
		}
		want := e.snap.paths[u.Squares[0]]
		if len(want) != len(u.Approach) {
			t.Errorf("unit %d: Approach %v, engine walk %v", u.ID, u.Approach, want)
		}
		if len(u.Approach) > 0 {
			walks++
		}
	}
	t.Logf("%d foes with an approach walk", walks)
}

// TestWalkUpOrder: an Attack that walks up first hands out the walk before
// anything else (no message held ahead of it), then the blow.
func TestWalkUpOrder(t *testing.T) {
	e := partyTurnScene(t)
	for _, u := range fighting(e).Field.Units {
		if len(u.Approach) == 0 {
			continue
		}
		e.events = nil
		out, err := e.doNow(game.Attack{Target: u.ID})
		if err != nil {
			t.Fatal(err)
		}
		walked, attacked := false, false
		for _, ev := range out.Events {
			switch ev.(type) {
			case game.Walked:
				if attacked {
					t.Error("walked after the blow")
				}
				walked = true
			case game.Attacked:
				attacked = true
			case game.Narrated:
				if !walked {
					t.Errorf("message %v before the walk", ev)
				}
			}
		}
		if !walked || !attacked {
			t.Errorf("walked %v attacked %v", walked, attacked)
		}
		return
	}
	t.Skip("no foe to walk up to")
}
