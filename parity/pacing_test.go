package parity

import (
	"testing"

	"gobox/front"
	"gobox/front/classic"
)

// TestPacingParity: the pace, the rest and the count of views classic takes
// from State equal por's, over the dev scenes, Enter on.
func TestPacingParity(t *testing.T) {
	seen := map[front.Pace]int{}
	for _, scene := range []string{"", "autocombat", "shop"} {
		e := engine(t, true, scene)
		for i := 0; i < 6; i++ {
			v, s := e.View(), e.Rules().State()
			b := v
			b.Pace, b.Resting, b.Seq = front.PaceNone, false, 0
			c := classic.Pacing(b, s)
			if c.Pace != v.Pace || c.Resting != v.Resting || c.Seq != v.Seq {
				t.Errorf("%s: pace %v rest %v seq %v, want %v %v %v", scene, c.Pace, c.Resting, c.Seq, v.Pace, v.Resting, v.Seq)
			}
			seen[v.Pace]++
			e.Apply(front.Action{Kind: front.ActOK})
			e.Step()
		}
	}
	t.Logf("paces seen: %v", seen)
	if len(seen) < 2 {
		t.Errorf("one pace only: %v", seen)
	}
}
