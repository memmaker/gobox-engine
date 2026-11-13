//go:build devviews

package engine

import (
	"gobox/game"
	"image"
	"testing"
)

// TestSidesApart (BUG-CLASSES 29): deskewed, at every monster distance, no
// foe starts inside the party's formation box. (The original's skewed
// layout does mix the front rows at distance 0; that is its own and is
// only logged.)
func TestSidesApart(t *testing.T) {
	for _, deskew := range []bool{false, true} {
		for dist := uint16(0); dist <= 3; dist++ {
			e := testEngine(t)
			e.Explore = true
			e.Deskew.Store(deskew)
			seed := uint32(1)
			e.FixedSeed = &seed
			scene := e.DevScene("autocombat")
			e.DevAction = func() {
				scene()
				e.state().Game.MonsterDist = dist
			}
			e.Init()
			e.Start(e.Play)
			f, ok := e.build().(game.Fighting)
			if !ok {
				t.Fatalf("deskew %v dist %d: no fight", deskew, dist)
			}
			in, box := foesInParty(f.Field.Units)
			switch {
			case in > 0 && deskew:
				t.Errorf("deskew dist %d: %d foes inside the party's box %v", dist, in, box)
			case in > 0:
				t.Logf("skewed dist %d: %d foes inside the party's box (the original's)", dist, in)
			}
		}
	}
}

// foesInParty counts the foes standing inside the party's bounding box.
func foesInParty(us []game.Unit) (in int, box image.Rectangle) {
	for _, u := range us {
		if !u.Foe {
			box = box.Union(image.Rectangle{u.At, u.At.Add(image.Pt(1, 1))})
		}
	}
	for _, u := range us {
		if u.Foe && u.At.In(box) {
			in++
		}
	}
	return in, box
}
