package pool

import (
	"gobox/core"
	"image"
	"testing"
)

// TestDeskewGrid: with Deskew every walled cell is a rectangle: the walls
// lie on the block grid's rows and columns, a door is a gap in the middle.
func TestDeskewGrid(t *testing.T) {
	e := &Engine{}
	e.initVM()
	e.Deskew.Store(true)
	for i := range 0x200 { // every side of every cell a wall
		e.gs.geo.Plane[i] = 0x11
	}
	e.gs.geo.Plane[0x300+8*16+8] = 1 << 6 // cell (8,8): its west side a door
	e.posX, e.posY = 8, 8
	e.ruleset().BeginFight(core.Fight{Board: core.Board{Straight: true}, Terms: core.FightTerms{PartyAt: image.Pt(8, 8)}}, nil)
	if _, err := e.ruleset().Resolve(core.Fight{}, core.BuildGround{Walls: e.gmap(), Dungeon: true}, nil); err != nil {
		t.Fatal(err)
	}
	for y := range 25 {
		for x := range 50 {
			wall := e.state().GroundInfo(e.state().Cmap.Ground[y][x])[0] == 0xFF
			onGrid := (y-0xA)%5 == 0 || (x-0x15)%6 == 0
			door := x == 0x15 && (y == 0xC || y == 0xD)
			if wall != (onGrid && !door) {
				t.Errorf("(%d,%d): wall %v, want %v", x, y, wall, !wall)
			}
		}
	}
}
