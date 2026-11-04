package parity

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

// TestPacingParity: State.Pacing (the pace, the rest and the count of states)
// over the dev scenes, Enter on.
func TestPacingParity(t *testing.T) {
	seen := map[game.Pace]int{}
	for _, scene := range []string{"", "autocombat", "shop"} {
		e := engine(t, true, scene)
		var last uint64
		for i := 0; i < 6; i++ {
			at(t, scene, e)
			p := e.Rules().State().Pacing
			if p == nil || p.Seq < last {
				t.Fatalf("%s step %d: pacing %+v after seq %d", scene, i, p, last)
			}
			last = p.Seq
			seen[p.Pace]++
			gameOf(e).Answer(front.Action{Kind: front.ActOK})
			e.Step()
		}
	}
	t.Logf("paces seen: %v", seen)
	if len(seen) < 2 {
		t.Errorf("one pace only: %v", seen)
	}
}
