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

// TestStraightIsOriginalUnskewed: the Deskew board is the original's, square
// for square, with the skew taken out (rules1e.StraightX): the same ground
// but for the door and corner posts squared up, and every combatant on the
// same square, whatever stands ahead. The foes
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
				sx := rules1e.StraightX(x, y)
				if sx < 0 || sx >= 50 || skew.Cmap.Ground[y][x] == 0 {
					continue
				}
				got := straight.Cmap.Ground[y][sx]
				if row := (y - 0xA) % 5; (got == 1 || got == 3) && (row == 0 || row == 1 || row == 4) {
					continue // a door's or a corner's post, squared up
				}
				if got != skew.Cmap.Ground[y][x] {
					t.Errorf("%s: ground (%d,%d) is %d, straight (%d,%d) %d", c.name, x, y, skew.Cmap.Ground[y][x], sx, y, straight.Cmap.Ground[y][sx])
				}
			}
		}
		for i := 1; i <= len(skew.Party); i++ {
			a, b := skew.Cpos[i], straight.Cpos[i]
			if a.Size == 0 {
				t.Fatalf("%s: combatant %d has no place", c.name, i)
			}
			sx := rules1e.StraightX(int(a.X), int(a.Y))
			if b.Size == a.Size && straight.GroundInfo(straight.Cmap.Ground[a.Y][sx])[0] == 0xFF {
				continue // its square is a squared-up post: the next free one
			}
			if b.Size != a.Size || int(b.X) != sx || b.Y != a.Y {
				t.Errorf("%s: combatant %d at (%d,%d), straight at (%d,%d)", c.name, i, a.X, a.Y, b.X, b.Y)
			}
		}
	}
}

// TestStraightWalls: on the Deskew board a walled cell is a rectangle: the
// west wall the block's columns 0 and 1, the north wall its row 0; a west
// door is rows 2 and 3 of the wall, between posts the wall's width.
func TestStraightWalls(t *testing.T) {
	s := straightFight(t, true, 1, 0)
	for y := 0xA + 5; y < 0xA+10; y++ { // cell (8,9): walled all round
		for x := 0x15; x < 0x15+6; x++ {
			wall := s.GroundInfo(s.Cmap.Ground[y][x])[0] == 0xFF
			if want := y == 0xA+5 || x <= 0x16; wall != want {
				t.Errorf("(%d,%d): wall %v, want %v", x, y, wall, want)
			}
		}
	}
	for y := 0xA; y < 0xA+5; y++ { // cell (8,8): its west side a door
		for x := 0x15; x <= 0x16; x++ {
			wall := s.GroundInfo(s.Cmap.Ground[y][x])[0] == 0xFF
			if want := y != 0xA+2 && y != 0xA+3; wall != want {
				t.Errorf("door (%d,%d): wall %v, want %v", x, y, wall, want)
			}
		}
	}
}
