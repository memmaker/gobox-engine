package engine

import (
	"gobox/rules"
	"gobox/rules1e"
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
	e.ruleset().BeginFight(rules.Fight{Board: rules.Board{Straight: true}, Terms: rules.FightTerms{PartyAt: image.Pt(8, 8)}}, nil)
	if _, err := e.ruleset().Resolve(rules.Fight{}, rules.BuildGround{Walls: e.gmap(), Dungeon: true}, nil); err != nil {
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

// TestDeskewFoesNotThroughWall: with Deskew the foes take the cell ahead only
// where that side is open; a wall or a door ahead (FreeSteps' distance 0)
// leaves them on the party's floor, as the original's placement does. They
// stood behind the Hobgoblin Barracks' secret door, out of reach.
func TestDeskewFoesNotThroughWall(t *testing.T) {
	for _, c := range []struct {
		name  string
		plane func(*Engine)
		wantX [2]int // the foes' columns
	}{
		{"wall", func(*Engine) {}, [2]int{0x15, 0x1A}},
		{"door", func(e *Engine) { e.gs.geo.Plane[0x300+8*16+8] = 1 << 2 }, [2]int{0x15, 0x1A}},
		{"open", func(e *Engine) { e.gs.geo.Plane[8*16+8] &= 0xF0; e.gs.geo.Plane[0x100+8*16+9] &= 0xF0 }, [2]int{0x1B, 0x20}},
	} {
		e := &Engine{}
		e.initVM()
		for i := range 0x200 { // every side of every cell a wall
			e.gs.geo.Plane[i] = 0x11
		}
		c.plane(e) // cell (8,8): its east side
		e.posX, e.posY, e.facing = 8, 8, 2
		s := e.state()
		s.Party = nil
		for i := range 12 {
			ch := &rules1e.Character{}
			ch.Okay, ch.Size = 1, 1
			if i >= 6 {
				ch.Side = rules1e.SideMonsters
			}
			s.Party = append(s.Party, ch)
		}
		s.Game.PartyCount, s.Game.MonsterDist = 6, 0
		e.ruleset().BeginFight(rules.Fight{Board: rules.Board{Straight: true}, Terms: rules.FightTerms{PartyAt: image.Pt(8, 8), PartyFacing: 2}}, nil)
		for _, a := range []rules.Action{rules.BuildGround{Walls: e.gmap(), Dungeon: true}, rules.PlaceFight{Walls: e.gmap()}} {
			if _, err := e.ruleset().Resolve(rules.Fight{}, a, nil); err != nil {
				t.Fatal(err)
			}
		}
		for i := 7; i <= 12; i++ {
			if p := s.Cpos[i]; p.Size == 0 || int(p.X) < c.wantX[0] || int(p.X) > c.wantX[1] {
				t.Errorf("%s ahead: foe %d at (%d,%d) size %d, want x %#x..%#x", c.name, i, p.X, p.Y, p.Size, c.wantX[0], c.wantX[1])
			}
		}
	}
}
