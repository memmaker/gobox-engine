package ui

import (
	"image"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestWalkPlaysStepByStep: a walk the game reports in one outcome reaches
// the front as one held moment a square, each on the board with the unit
// on that square, so the figure walks it rather than jumping to its end.
func TestWalkPlaysStepByStep(t *testing.T) {
	fight := play.Fighting{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.Attack{Target: 2}}}},
		Field: play.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, Member: 0}, {ID: 2, Member: -1, Foe: true, At: image.Pt(1, 1)}}}}
	path := []game.Step{{At: image.Pt(2, 1)}, {At: image.Pt(3, 2)}, {At: image.Pt(4, 2)}}
	f := &fake{st: fight, out: play.Outcome{Events: []play.Event{play.Walked{Unit: 2, Path: path}}}}
	g := NewGame(f, game.Data{})
	g.Settle() // the fight's first board
	if err := g.Answer(FrontAction{Kind: ActAttack, N: 2}); err != nil {
		t.Fatal(err)
	}
	evs := settled(g)
	var at []image.Point
	for _, e := range evs {
		if e.Kind != game.EvMove {
			continue
		}
		if e.Ms <= 0 || boardOfFrame(e) == nil {
			t.Fatalf("step not held: %+v", e)
		}
		for _, u := range boardOfFrame(e).Units {
			if u.ID == 2 {
				at = append(at, image.Pt(u.X, u.Y))
			}
		}
	}
	if len(at) != len(path) || at[0] != path[0].At || at[2] != path[2].At {
		t.Fatalf("steps shown at %v, walked %v", at, path)
	}
}
