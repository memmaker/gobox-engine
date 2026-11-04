package parity

import (
	"testing"

	"gobox/front"
)

// TestWallsParity: por builds no 3D view of its own, so State.Walls is nil
// at each stop.
func TestWallsParity(t *testing.T) {
	for _, scene := range []string{"", "shop", "autocombat"} {
		e := engine(t, true, scene)
		settled(e, front.Action{})
		for i := 0; i < 6; i++ {
			if s := e.Rules().State(); s.Walls != nil {
				t.Errorf("%s: walls %+v", scene, s.Walls)
			}
			settled(e, front.Action{Kind: front.ActOK})
		}
	}
}
