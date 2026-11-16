package engine

import (
	"gobox/rules"
	"gobox/rules1e"
	"image"
	"testing"
)

// straightFight sets up a fight at cell (8,8) facing east on a map of walled
// cells: (8,8) has a door west, an open side north, and east the side given
// (0 open, 1 a wall); 6 of the party, 23 foes, no clutter (its dice fall by
// board square).
func straightFight(t *testing.T, straight bool, east byte, dist uint16) *rules1e.RuleState {
	e := &Engine{}
	e.initVM()
	for i := range 0x200 { // every side of every cell a wall
		e.gs.geo.Plane[i] = 0x11
	}
	e.gs.geo.Plane[0x300+8*16+8] = 1 << 6 // its west side a door
	e.gs.geo.Plane[8*16+8] = east         // north open
	e.gs.geo.Plane[0x100+7*16+8] = 0x01   // (8,7): south open
	if east == 0 {
		e.gs.geo.Plane[0x100+8*16+9] = 0x10 // (9,8): west open
	}
	e.posX, e.posY, e.facing = 8, 8, 2
	s := e.state()
	tables := *s.T()
	tables.Gen.Clutter = nil
	s.Tables = &tables
	s.Party = nil
	for i := range 29 {
		ch := &rules1e.Character{}
		ch.Okay, ch.Size = 1, 1
		if i >= 6 {
			ch.Side = rules1e.SideMonsters
		}
		s.Party = append(s.Party, ch)
	}
	s.Game.PartyCount = 6
	e.ruleset().BeginFight(rules.Fight{Board: rules.Board{Straight: straight}, Terms: rules.FightTerms{MonsterDist: int32(dist), PartyAt: image.Pt(8, 8), PartyFacing: 2}}, nil)
	for _, a := range []rules.Action{rules.BuildGround{Walls: e.gmap(), Dungeon: true}, rules.PlaceFight{Walls: e.gmap()}} {
		if _, err := e.ruleset().Resolve(rules.Fight{}, a, nil); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

// straightCellOf is the cell (column, row from the party's) of board square (x, y).
func straightCellOf(x, y int) image.Point { return image.Pt((x-20+50)/5-10, (y-0xA+50)/5-10) }

// TestStraightIsOriginalUnskewed: the Deskew board is the original's with
// the skew taken out and its cells made square (rules.StraightX): the same
// floor, and everyone the original placed is placed, in its cell or (its
// square a post here) the next, on a square of their own and never in a
// wall, whatever stands ahead. The foes
// were once put a cell ahead through walls and doors, on a map of its own.
func TestStraightIsOriginalUnskewed(t *testing.T) {
	for _, c := range []struct {
		name string
		east byte
		dist uint16
	}{{"wall ahead", 1, 0}, {"open ahead", 0, 1}, {"open ahead, close", 0, 0}} {
		skew, straight := straightFight(t, false, c.east, c.dist), straightFight(t, true, c.east, c.dist)
		for y := range 25 {
			for x := range 50 {
				sx := rules.StraightX(x, y)
				if sx < 0 || sx >= 50 || skew.Cmap.Ground[y][x] == 0 || (y-0xA)%5 == 0 || (sx-20+50)%5 == 0 {
					continue // off the board, or a cell's wall row or column
				}
				if got := straight.Cmap.Ground[y][sx]; got != skew.Cmap.Ground[y][x] {
					t.Errorf("%s: ground (%d,%d) is %d, straight (%d,%d) %d", c.name, x, y, skew.Cmap.Ground[y][x], sx, y, got)
				}
			}
		}
		if len(straight.Party) != len(skew.Party) {
			t.Errorf("%s: %d placed, the original %d", c.name, len(straight.Party), len(skew.Party))
		}
		taken, same := map[image.Point]int{}, 0
		for i := 1; i <= len(straight.Party) && i <= len(skew.Party); i++ {
			a, b := skew.Cpos[i], straight.Cpos[i]
			at, was := image.Pt(int(b.X), int(b.Y)), image.Pt(rules.StraightX(int(a.X), int(a.Y)), int(a.Y))
			switch {
			case a.Size == 0 || b.Size == 0:
				t.Errorf("%s: combatant %d has no place", c.name, i)
			case straight.GroundInfo(straight.Cmap.Ground[b.Y][b.X])[0] == 0xFF:
				t.Errorf("%s: combatant %d stands in a wall at %v", c.name, i, at)
			case taken[at] != 0:
				t.Errorf("%s: combatants %d and %d share %v", c.name, taken[at], i, at)
			case !straightCellOf(at.X, at.Y).In(image.Rectangle{Min: straightCellOf(was.X, was.Y), Max: straightCellOf(was.X, was.Y)}.Inset(-1)):
				t.Errorf("%s: combatant %d in cell %v, far from the original's %v", c.name, i, straightCellOf(at.X, at.Y), straightCellOf(was.X, was.Y))
			}
			taken[at] = i
			if at == was {
				same++
			}
		}
		t.Logf("%s: %d of %d on the original's own square", c.name, same, len(straight.Party))
	}
}

// TestStraightWalls: on the Deskew board a cell is a 5x5 block, its walls
// one square thick: a walled cell's west wall is its block's column 0, its
// north wall its row 0, its floor 4x4; a west door is rows 2 and 3 of the
// wall, whoever is placed beside it.
func TestStraightWalls(t *testing.T) {
	s := straightFight(t, true, 1, 0)
	wall := func(x, y int) bool { return s.GroundInfo(s.Cmap.Ground[y][x])[0] == 0xFF }
	for y := 0xA + 5; y < 0xA+10; y++ { // cell (8,9): walled all round
		for x := 20; x < 25; x++ {
			if want := y == 0xA+5 || x == 20; wall(x, y) != want {
				t.Errorf("(%d,%d): wall %v, want %v", x, y, wall(x, y), want)
			}
		}
	}
	for y := 0xA + 1; y < 0xA+5; y++ { // cell (8,8): its west side a door
		if want := y != 0xA+2 && y != 0xA+3; wall(20, y) != want {
			t.Errorf("door (20,%d): wall %v, want %v", y, wall(20, y), want)
		}
	}
}
