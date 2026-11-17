package rulesxxvc

import (
	"testing"

	"gobox/rules"
)

// TestDeskewWestColumn (old deskew_test.go TestDeskewWestColumn): on the deskewed arena every map square with a
// west wall has it as one straight column, rows 1..4 of its block at the block's first column, the post above it;
// a door leaves its two middle rows open; the skewed arena is the original's, unchanged by the option.
func TestDeskewWestColumn(t *testing.T) {
	var m rules.Map
	for i := range 0x200 {
		m.Plane[i] = 0x11 // every side of every square a wall (no door: DoorAt 0)
	}
	m.Plane[0x300+8*16+8] = 1 << 6 // (8,8): its west side a door
	party := rules.BoardPos{X: 8, Y: 8}
	straight := buildDungeonArena(0, &m, party, 1, true)
	skewed := buildDungeonArena(0, &m, party, 1, false)
	walls := 0
	for dr := -dungeonReachY; dr <= dungeonReachY; dr++ {
		for c := -4; c <= 5; c++ {
			x0, y0 := rules.StraightParty+c*rules.StraightCell, dr*blockRows+blockY0
			if got := straight[y0][x0]; got != rules.TerrainID(straightPost+1) {
				t.Errorf("square %d,%d: post is terrain %d, want %d", c, dr, got, straightPost+1)
			}
			for b := 1; b <= 4; b++ {
				want := rules.TerrainID(straightWest + 1)
				if c == 0 && dr == 0 && (b == 2 || b == 3) {
					want = floorHitek // the door's two middle rows
				}
				if got := straight[y0+b][x0]; got != want {
					t.Errorf("square %d,%d: west wall row %d is terrain %d, want %d", c, dr, b, got, want)
				}
				if want != floorHitek {
					walls++
				}
			}
		}
	}
	if walls == 0 {
		t.Fatal("no west wall in the window")
	}
	// the option changes the arena, the skewed one has no such column
	if skewed[blockY0+1][rules.StraightParty] == rules.TerrainID(straightWest+1) {
		t.Error("the skewed arena shows the straight column")
	}
}

// TestDeskewNoSkewedPieces: the straight arena holds none of the original's skewed wall pieces (layStructure's terrains
// that are no straight wall and no scatter): off the straight walls' lines they stood on the board as leftovers. Detects
// the class: anything on the straight board drawn from the skewed layout.
func TestDeskewNoSkewedPieces(t *testing.T) {
	var m rules.Map
	for i := range 0x200 {
		m.Plane[i] = 0x11
	}
	party := rules.BoardPos{X: 8, Y: 8}
	const place = rules.TraitScatterTiles
	pieces := func(lay func(d *dungeonArena)) map[rules.TerrainID]bool {
		d := &dungeonArena{rng: &rules.RNG{State: 1}}
		lay(d)
		got := map[rules.TerrainID]bool{}
		for y := range d.cells {
			for _, c := range d.cells[y] {
				got[c] = true
			}
		}
		return got
	}
	skewed := pieces(func(d *dungeonArena) { d.eachSquare(func() { d.layStructure(&m, party) }) })
	scatter := pieces(func(d *dungeonArena) { d.eachSquare(func() { d.scatterSquare(&m, party, place) }) })
	straight := buildDungeonArena(place, &m, party, 1, true)
	for y := range straight {
		for x, c := range straight[y] {
			wall := c == straightWest+1 || c == straightNorth+1 || c == straightPost+1
			if skewed[c] && !wall && !scatter[c] {
				t.Fatalf("%d,%d: terrain %d is a skewed wall piece", x, y, c)
			}
		}
	}
}
