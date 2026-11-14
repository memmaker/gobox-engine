package engine

import "testing"

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
