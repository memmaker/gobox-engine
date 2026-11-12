package pool

import (
	"rules1e"
	"testing"
)

// TestTileOfNoWrap: squares just off the left and right edges are bare
// floor, not the far edge's wall pieces (the map never wraps).
func TestTileOfNoWrap(t *testing.T) {
	e := &Engine{}
	e.state().Cmap = &rules1e.CombatMap{}
	e.initVM()
	bare := e.state().GroundInfo(rules1e.PorGroundIDs.Floor)[3]
	var wall byte
	for g := range rules1e.GroundTable {
		if rules1e.GroundTable[g][3] != bare {
			wall = byte(g)
			break
		}
	}
	e.state().Cmap.Ground[3][0], e.state().Cmap.Ground[3][49] = rules1e.Ground(wall), rules1e.Ground(wall)
	if got := e.state().TileOf(50, 2); got != bare {
		t.Errorf("right of row 2: tile %d, the next row's first square leaks in", got)
	}
	if got := e.state().TileOf(-1, 4); got != bare {
		t.Errorf("left of row 4: tile %d, the previous row's last square leaks in", got)
	}
	if got := e.state().TileOf(0, 3); got == bare {
		t.Errorf("on the map: bare ground, want the wall")
	}
}
