package rules1e

import (
	"testing"
	"time"
)

// TestPlaceOneGivesUp: with every slot of every team taken, placing one more
// fails instead of running its rows off the block for ever (a soft-lock of a
// big fight on a deskewed map), whichever way each side faces.
func TestPlaceOneGivesUp(t *testing.T) {
	for _, deskew := range []bool{true, false} {
		for d := range byte(4) {
			for s := 0; s <= 1; s++ {
				r := &Rules{S: &RuleState{}, straight: deskew}
				r.S.Place.Side, r.S.Place.Dir[s], r.S.Place.Rows[s] = byte(s), Dir(d), 3
				done := make(chan bool)
				go func() { done <- r.placeOne(nil, 1) }() // nil: no walls between the team blocks
				select {
				case ok := <-done:
					if ok {
						t.Fatalf("deskew %v dir %d side %d: placed with no free slot", deskew, d, s)
					}
				case <-time.After(time.Second):
					t.Fatalf("deskew %v dir %d side %d: placeOne hangs", deskew, d, s)
				}
			}
		}
	}
}

// TestFacingDir: deskewed, north is north; else the original's table.
func TestFacingDir(t *testing.T) {
	r := &Rules{straight: true}
	if r.facingDir(0) != 0 || r.facingDir(2) != 4 {
		t.Errorf("deskewed facing: north %d, south %d", r.facingDir(0), r.facingDir(2))
	}
	r.straight = false
	if r.facingDir(0) != byte(FacingDir[0]) {
		t.Errorf("facing %d", r.facingDir(0))
	}
}
