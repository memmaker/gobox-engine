package pool

import (
	"gobox/rules1e"
	"testing"
)

// TestCombatGroundBlocks: every ground publishes the original's data as it is:
// Wall is what stops a move (groundTable[g][0] == 0xFF, combatrules.go), so
// gobox's render mode (decal, mesh, extruded sprite) follows the blocking flag.
func TestCombatGroundBlocks(t *testing.T) {
	e := &Engine{}
	e.initVM()
	for g, r := range rules1e.GroundTable {
		q := e.state().GroundSquare(byte(g))
		if q.Wall != (r[0] == 0xFF) || q.Tile != int(r[3]) || q.Height != int(r[2]) {
			t.Errorf("ground %#x: published %+v, data %v", g, q, r)
		}
	}
}

// TestGroundTilesExist (BUG-CLASSES 16): every ground's tile goes through the
// frame table (groundTable[g][3]) to a frame SQRPACI has: tiles under 0x80
// in set 1, the rest in set 2 (drawTile, TilePix's t*0x80+f).
func TestGroundTilesExist(t *testing.T) {
	e := testEngine(t)
	e.Init()
	for g := range rules1e.GroundTable {
		tl := e.state().GroundSquare(byte(g)).Tile
		im, f := e.sqrpaci[tl>>7], tl&0x7F
		if im == nil || f >= im.Count || im.Frame(f) == nil {
			t.Errorf("ground %#x: tile %#x has no SQRPACI frame", g, tl)
		}
	}
}
