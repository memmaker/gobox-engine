package rulesxxvc

import (
	"testing"

	"gobox/rules"
)

// PlaceAllCreatures by hand from combatst.pas (open country, party facing north, the monsters one block off):
// the party's first member: ring 0 quad 0, FindPlacement's leg start (ringStartColumns/Rows[0][0] = 5, 3), cell
// (5,3) -> x 5+22, y 3+10. The second finds (5,3) taken, legShift 1 steps the leg by quarterDirections[2]=3
// (1,1): (6,4). The monster: team 1 faces south (quad 2): start (5,2) at block (0,-1): x 5-5+22, y 2-5+10.
func TestPlaceAllCreatures(t *testing.T) {
	r := New()
	a, b := r.Add(member("a", 10, 0, 0, 0)), r.Add(member("b", 10, 0, 0, 0))
	m := member("m", 10, 0, 0, 0)
	m.Side = rules.SideMonsters
	mid := r.Add(m)
	r.memberCount = 2
	f := rules.Fight{Party: rules.Party{a, b, mid}, Terms: rules.FightTerms{MonsterDist: 1, Area: rules.AreaWilderness}}
	if err := r.buildGround(rules.Fight{}, rules.BuildGround{Terrain: 1}, &rules.RNG{State: 1}); err != nil {
		t.Fatal(err)
	}
	r.BeginFight(f, nil)
	o := r.placeAllCreatures(f, nil)
	want := map[rules.CharID]rules.BoardPos{a: {X: 27, Y: 13}, b: {X: 28, Y: 14}, mid: {X: 22, Y: 7}}
	for _, u := range r.ground.Units {
		if u.At != want[u.ID] {
			t.Errorf("unit %d at %v, want %v", u.ID, u.At, want[u.ID])
		}
		delete(want, u.ID)
	}
	if len(want) != 0 || len(o.Effects) != 0 {
		t.Errorf("not placed %v, effects %v", want, o.Effects)
	}
	if u := r.ground.Units[0]; u.Facing != 7 || r.ground.Units[2].Facing != 3 {
		t.Errorf("facing %v %v, want 7 and 3 (quarterDirections)", u.Facing, r.ground.Units[2].Facing)
	}
}

// The wilderness arena is 50x25, every unset cell given ground, and the same seed lays the same ground. Terrain 0
// is "Cant identify wild terrain".
func TestWildernessArena(t *testing.T) {
	a, err := buildWildernessArena(1, &rules.RNG{State: 7})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := buildWildernessArena(1, &rules.RNG{State: 7})
	if *a != *b {
		t.Error("same seed, different ground")
	}
	for y := range a {
		for x := range a[y] {
			if a[y][x] == unsetTerrain || a[y][x] == 0 {
				t.Fatalf("cell %d,%d left %d", x, y, a[y][x])
			}
		}
	}
	if _, err := buildWildernessArena(0, &rules.RNG{}); err == nil {
		t.Error("terrain 0 must not lay")
	}
}

// oneWall is a 16x16 map with a wall on the north side of (8,8) and the south side of (8,7).
type oneWall struct{}

func (oneWall) WallAt(x, y, dir int) byte {
	if x == 8 && (y == 8 && dir == dirNorth || y == 7 && dir == dirNorth+4) {
		return 1
	}
	return 0
}

// DoorAt is the barrier plane: 0 a solid wall, 1 where there is no wall (View3D.GetBarrierType).
func (o oneWall) DoorAt(x, y, dir int) byte {
	if o.WallAt(x, y, dir) != 0 {
		return 0
	}
	return 1
}
func (oneWall) MapCell(x, y int) byte { return 0 }

// BuildDungeonArena by hand: floor 45 (kind $80), the party's square (8,8) at block origin col 21, row 10 with a
// north wall: PlaceNorthWall plots 24 at (3,0),(4,0) and 25 at (3,1),(4,1) (cells hold terrain + 1), and the corner
// states 1 (left) and 4 (right) lay leftWallTiles[1] = (24,24)/(25,25) at columns 1-2 and rightWallTiles[4] the same
// at columns 5-6. An open map lays nothing but the floor (7 for kind $88).
func TestBuildDungeonArena(t *testing.T) {
	c := buildDungeonArena(rules.TraitHasKind, oneWall{}, rules.BoardPos{X: 8, Y: 8}, 5, false)
	const (
		top, bottom = 25, 26 // 24 + 1, 25 + 1
		col0, row0  = blockX0, blockY0
	)
	for _, cell := range []struct {
		x, y int
		want rules.TerrainID
	}{
		{col0 + 1, row0, top}, {col0 + 2, row0, top}, {col0 + 3, row0, top}, {col0 + 4, row0, top},
		{col0 + 5, row0, top}, {col0 + 6, row0, top},
		{col0 + 1, row0 + 1, bottom}, {col0 + 2, row0 + 1, bottom}, {col0 + 3, row0 + 1, bottom},
		{col0 + 4, row0 + 1, bottom}, {col0 + 5, row0 + 1, bottom}, {col0 + 6, row0 + 1, bottom},
	} {
		if got := c[cell.y][cell.x]; got != cell.want {
			t.Errorf("cell %d,%d = %d, want %d", cell.x, cell.y, got, cell.want)
		}
	}
	open := buildDungeonArena(rules.TraitHasKind|rules.TraitLowTech|rules.TraitSecondTileSet, openMap{}, rules.BoardPos{X: 8, Y: 8}, 5, false)
	for y := range open {
		for x := range open[y] {
			if open[y][x] != floorLotek {
				t.Fatalf("open map cell %d,%d = %d, want the floor", x, y, open[y][x])
			}
		}
	}
}

type openMap struct{}

func (openMap) WallAt(x, y, dir int) byte { return 0 }
func (openMap) DoorAt(x, y, dir int) byte { return 1 }
func (openMap) MapCell(x, y int) byte     { return 0 }

// TestBuildDungeonArenaSeed: the arena is drawn from the script block's own stream: the same block lays the same
// arena, another block another (scatter on: 65 draws of Random(2) plus the walls' clutter).
func TestBuildDungeonArenaSeed(t *testing.T) {
	place := rules.TraitHasKind | rules.TraitScatterTiles
	at := rules.BoardPos{X: 8, Y: 8}
	a, b, c := buildDungeonArena(place, oneWall{}, at, 16, false), buildDungeonArena(place, oneWall{}, at, 16, false), buildDungeonArena(place, oneWall{}, at, 17, false)
	if *a != *b {
		t.Error("the same block laid two arenas")
	}
	if *a == *c {
		t.Error("two blocks laid the same arena")
	}
}

// TestWallBetweenBarrierPlane: GetWallKind: barrier 0 is 1, else no wall 0, else 3 (combatst.pas:185).
func TestWallBetweenBarrierPlane(t *testing.T) {
	if got := wallBetween(oneWall{}, 8, 8, 8, dirNorth); got != wallPlain {
		t.Errorf("solid wall: %d", got)
	}
	if got := wallBetween(openMap{}, 8, 8, 8, dirNorth); got != wallNone {
		t.Errorf("no wall: %d", got)
	}
	if got := wallBetween(doorMap{}, 8, 8, 8, dirNorth); got != wallDoor {
		t.Errorf("a wall with barrier 2: %d", got)
	}
}

// doorMap has a passable wall (wall type 1, barrier 2) everywhere.
type doorMap struct{}

func (doorMap) WallAt(x, y, dir int) byte { return 1 }
func (doorMap) DoorAt(x, y, dir int) byte { return 2 }
func (doorMap) MapCell(x, y int) byte     { return 0 }
