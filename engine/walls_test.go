package engine

import (
	"testing"

	"gobox/asset"
	"gobox/game"
	"gobox/rules"
)

// TestAdvWallsCOTAB: cotab's imported wall table gives the 3D view its pool
// textures for the walls loaded: Tilverton's (set 2, WALLDEF 1..3), a
// two-chunk block (14: slot 2 its second chunk) and block 10 (recorded as
// 0). After Begin, the slots set by hand: the engine still opens every game on
// POR's set 3, area 0 (partyMenu), which loads no walls in cotab.
func TestAdvWallsCOTAB(t *testing.T) {
	const series = "pool"
	e := advStarted(t, advEnvCOTAB)
	for _, a := range []game.Action{game.QuickStart{}, game.Begin{}} {
		if _, err := e.doNow(a); err != nil {
			t.Fatalf("%T: %v", a, err)
		}
	}
	for n := 0; n < 200; n++ { // past the opening scene
		if _, ok := e.build().(game.Exploring); ok {
			break
		}
		press(e)
	}
	e.viewMode = 1
	for _, c := range []struct {
		set   rules.Set
		loads [rules.WallSlots]int // a slot's block; -1 none
		want  map[int]string
	}{
		{2, [3]int{1, 2, 3}, map[int]string{1: "wall_tilverton_v01", 6: "wall_podol_v06", 11: "wall_tilverton_v11"}},
		{5, [3]int{14, -1, 8}, map[int]string{}},
		{4, [3]int{9, 10, 11}, map[int]string{}},
	} {
		e.set = c.set
		for s, b := range c.loads {
			e.mapPairA[s] = -1
			if b >= 0 {
				e.LoadWalldef(b, s+1)
			}
		}
		w := e.build().(game.Exploring).Map.Walls
		if len(w.Side) != rules.AreaSize || len(w.Pool) == 0 {
			t.Fatalf("set %d %v: no walls", c.set, c.loads)
		}
		for s := range rules.WallSlots { // every slot has textures
			if w.Pool[s*rules.SetWalls+1] == "" {
				t.Errorf("set %d %v: slot %d untextured: %v", c.set, c.loads, s+1, w.Pool)
			}
		}
		for v, tex := range c.want {
			if w.Pool[v] != tex {
				t.Errorf("set %d %v: wall %d is %q, want %q", c.set, c.loads, v, w.Pool[v], tex)
			}
		}
		for v, tex := range w.Pool {
			if !asset.Has(series, tex) {
				t.Errorf("wall %d: %q not in the %s series", v, tex, series)
			}
		}
	}
	e.viewMode = 2 // off the 3D view: none
	if w := e.build().(game.Exploring).Map.Walls; w.Side != nil {
		t.Errorf("overland has walls")
	}
}
