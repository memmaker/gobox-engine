package ui

import (
	"testing"

	"gobox/game"
	"gobox/rules"
)

// TestWalkPlaysStepByStep: a walk the game reports in one outcome reaches
// the front as one held moment a square, each on the board with the unit
// on that square, so the figure walks it rather than jumping to its end.
func TestWalkPlaysStepByStep(t *testing.T) {
	fight := game.Fighting{Common: game.Common{Allowed: game.Allowed{Plain: []game.Action{game.Attack{Target: 2}}}},
		Field: game.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, Member: 0}, {ID: 2, Member: -1, Foe: true, At: rules.BoardPos{X: 1, Y: 1}}}}}
	path := []game.Step{{At: rules.BoardPos{X: 2, Y: 1}}, {At: rules.BoardPos{X: 3, Y: 2}}, {At: rules.BoardPos{X: 4, Y: 2}}}
	f := &fake{st: fight, out: game.Outcome{Events: []game.Event{game.Walked{Unit: 2, Path: path}}}}
	g := NewGame(f, game.Data{})
	g.Settle() // the fight's first board
	if err := g.Answer(Action{Kind: ActAttack, N: 2}); err != nil {
		t.Fatal(err)
	}
	evs := settled(g)
	var at []rules.BoardPos
	for _, e := range evs {
		if e.Kind != game.EvMove {
			continue
		}
		if e.Ms <= 0 || boardOfFrame(e) == nil {
			t.Fatalf("step not held: %+v", e)
		}
		for _, u := range boardOfFrame(e).Units {
			if u.ID == 2 {
				at = append(at, rules.BoardPos{X: u.X, Y: u.Y})
			}
		}
	}
	if len(at) != len(path) || at[0] != path[0].At || at[2] != path[2].At {
		t.Fatalf("steps shown at %v, walked %v", at, path)
	}
}
