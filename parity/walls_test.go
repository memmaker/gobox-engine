package parity

import (
	"testing"

	"gobox/front"
	"gobox/front/classic"
)

// TestWallsParity: por builds no 3D view of its own (View.Walls nil), so
// classic, with no State.Walls, keeps it nil at each stop.
func TestWallsParity(t *testing.T) {
	for _, scene := range []string{"", "shop", "autocombat"} {
		e := engine(t, true, scene)
		settled(e, front.Action{})
		for i := 0; i < 6; i++ {
			v, s := e.View(), e.Rules().State()
			if c := classic.Walls(v, s, e.Data().Art); c.Walls != v.Walls || s.Walls != nil {
				t.Errorf("%s: walls %+v, state %+v", scene, c.Walls, s.Walls)
			}
			settled(e, front.Action{Kind: front.ActOK})
		}
	}
}
