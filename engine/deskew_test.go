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
	s.Game.PartyCount, s.Game.MonsterDist = 6, dist
	e.ruleset().BeginFight(rules.Fight{Board: rules.Board{Straight: straight}, Terms: rules.FightTerms{PartyAt: image.Pt(8, 8), PartyFacing: 2}}, nil)
	for _, a := range []rules.Action{rules.BuildGround{Walls: e.gmap(), Dungeon: true}, rules.PlaceFight{Walls: e.gmap()}} {
		if _, err := e.ruleset().Resolve(rules.Fight{}, a, nil); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

// TestStraightIsOriginalUnskewed: the Deskew board is the original's with
// the skew taken out (rules1e.StraightX): every combatant on the same
// square, whatever stands ahead, and the same floor; only the walls are
// squared up after (TestStraightWalls). The foes were once put a cell ahead
// through walls and doors, on a map of its own.
func TestStraightIsOriginalUnskewed(t *testing.T) {
	for _, c := range []struct {
		name string
		east byte
		dist uint16
	}{{"wall ahead", 1, 0}, {"open ahead", 0, 1}, {"open ahead, close", 0, 0}} {
		skew, straight := straightFight(t, false, c.east, c.dist), straightFight(t, true, c.east, c.dist)
		for y := range 25 {
			for x := range 50 {
				sx := rules1e.StraightX(x, y)
				if sx < 0 || sx >= 50 || skew.Cmap.Ground[y][x] == 0 || (y-0xA)%5 == 0 || (sx-0x15+60)%6 < 2 {
					continue // off the board, or a cell's wall row or columns
				}
				if got := straight.Cmap.Ground[y][sx]; got != skew.Cmap.Ground[y][x] {
					t.Errorf("%s: ground (%d,%d) is %d, straight (%d,%d) %d", c.name, x, y, skew.Cmap.Ground[y][x], sx, y, got)
				}
			}
		}
		for i := 1; i <= len(skew.Party); i++ {
			a, b := skew.Cpos[i], straight.Cpos[i]
			if a.Size == 0 {
				t.Fatalf("%s: combatant %d has no place", c.name, i)
			}
			if b.Size != a.Size || int(b.X) != rules1e.StraightX(int(a.X), int(a.Y)) || b.Y != a.Y {
				t.Errorf("%s: combatant %d at (%d,%d), straight at (%d,%d)", c.name, i, a.X, a.Y, b.X, b.Y)
			}
			if straight.GroundInfo(straight.Cmap.Ground[b.Y][b.X])[0] == 0xFF {
				t.Errorf("%s: combatant %d stands in a wall at (%d,%d)", c.name, i, b.X, b.Y)
			}
		}
	}
}

// TestStraightWalls: on the Deskew board the walls are one square thick: a
// walled cell's west wall is its block's column 0, its north wall its row
// 0; a west door is rows 2 and 3 of the wall. No one is walled in or over.
func TestStraightWalls(t *testing.T) {
	s := straightFight(t, true, 1, 0)
	stands := func(x, y int) bool { who, _ := s.MapAt(x, y); return who != 0 }
	wall := func(x, y int) bool { return s.GroundInfo(s.Cmap.Ground[y][x])[0] == 0xFF }
	for y := 0xA + 5; y < 0xA+10; y++ { // cell (8,9): walled all round
		for x := 0x15; x < 0x15+6; x++ {
			if want := y == 0xA+5 || x == 0x15; wall(x, y) != want && !stands(x, y) {
				t.Errorf("(%d,%d): wall %v, want %v", x, y, wall(x, y), want)
			}
		}
	}
	for y := 0xA + 1; y < 0xA+5; y++ { // cell (8,8): its west side a door
		if want := y != 0xA+2 && y != 0xA+3; wall(0x15, y) != want && !stands(0x15, y) {
			t.Errorf("door (%d,%d): wall %v, want %v", 0x15, y, wall(0x15, y), want)
		}
	}
}
