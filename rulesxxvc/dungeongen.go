package rulesxxvc

import "gobox/rules"

// wallBetween is combatst.pas WallBetween's kind from the map's sides: each side's GetWallKind or-ed. Off the 16x16
// map a side counts as 1, except east and west along the party's row (partyRow).
// GetWallKind reads the barrier plane first: barrier type 0 (DoorAt 0, a solid wall) is kind 1, else no wall (WallAt 0)
// is 0, else 3 (a wall that can be passed: door, secret). DoorAt answers 1 on a side with no wall, as GetBarrierType.
// The old engine's side3D (cmap.go) reads the same: passable 0 -> 1, wallAt 0 -> 0, else 3.
func wallBetween(walls rules.MapSides, partyRow, x, y, dir int) int {
	side := func(cx, cy, facing int) int {
		if cx >= 0 && cx < arenaMapSide && cy >= 0 && cy < arenaMapSide {
			switch {
			case walls.DoorAt(cx, cy, facing) == 0:
				return wallPlain
			case walls.WallAt(cx, cy, facing) == 0:
				return wallNone
			}
			return wallDoor
		}
		if cy == partyRow && (facing == dirEast || facing == dirWest) {
			return wallNone
		}
		return wallPlain
	}
	return side(x, y, dir) | side(x+dirStep[dir].X, y+dirStep[dir].Y, (dir+4)%8)
}

// Dungeon arena floors: lotek's (rules.TraitLowTech) is terrain 7, the others' 45 (BuildDungeonArena).
const (
	floorLotek      = 7
	floorHitek      = 45
	squareVariantEv = 0x40 // GetMapEvent and $40
)

// The wall kinds a side can have.
const (
	wallNone  = 0
	wallPlain = 1
	wallDoor  = 3
)

// dungeonArena is the arena while BuildDungeonArena lays it: GDVars.combatArena^.cells, arenaOX/OY and the wall kinds
// of the map square being laid.
type dungeonArena struct {
	cells                         [arenaRows][arenaCols]rules.TerrainID
	rng                           *rules.RNG
	ox, oy                        int
	wallWest, wallNorth, wallEast int
	squareVariant                 int
	straight                      bool // Board.Straight (extra, Deskew): the skew is taken out, see rules.StraightX
}

// cell is the arena cell of the block coordinates (px, py) of the current map square: PlotCell's column and row,
// the column deskewed on the straight board.
func (d *dungeonArena) cell(px, py int) (col, row int) {
	col, row = d.ox*blockCols+d.oy*blockRows+blockX0+px, d.oy*blockRows+blockY0+py
	if d.straight {
		col = rules.StraightX(col, row)
	}
	return col, row
}

// plot is combatst.pas PlotCell: terrain pv + 1 at (px, py) of the current map square's block, inside the arena.
func (d *dungeonArena) plot(px, py, pv int) {
	col, row := d.cell(px, py)
	if col >= 0 && col < arenaCols && row >= 0 && row < arenaRows {
		d.cells[row][col] = rules.TerrainID(byte(pv + 1))
	}
}

// plotOnFloor is ScatterWallTiles' PlaceTileOn: PlotCell, onto a floor cell (terrain 7 or 45) only. The original's
// first upper-bound test reads row where col is meant; a column past the arena then reads another row's cell and
// PlotCell's own test plots nothing, so it is skipped here.
func (d *dungeonArena) plotOnFloor(px, py, pv int) {
	col, row := d.cell(px, py)
	if col >= 0 && col < arenaCols && row >= 0 && row < arenaRows {
		if t := d.cells[row][col]; t == floorLotek || t == floorHitek {
			d.plot(px, py, pv)
		}
	}
}

// roll is Random(n) from the one generator, rules.RNG (high word mod n); the original's runtime reduces by multiply.
func (d *dungeonArena) roll(n int) int { return d.rng.N(n) }

// scatterWallTiles is combatst.pas ScatterWallTiles(mode): tiles at random along the north and east walls of the
// square being laid, Random's draws in the original's order.
func (d *dungeonArena) scatterWallTiles(place rules.PlaceTraits) {
	// eachNorth: for i := 4 to 6, a Random(10) < chance, then Random(len(tiles)) picks the tile.
	eachNorth := func(chance int, tiles ...int) {
		for i := 4; i <= 6; i++ {
			if d.roll(10) < chance {
				d.plotOnFloor(i, 2, tiles[d.roll(len(tiles))])
			}
		}
	}
	// eachEast: for j := 2 to 4, a Random(10) < 8, then the tile's roll, on the east wall's row.
	eachEast := func(tiles ...int) {
		for j := 2; j <= 4; j++ {
			if d.roll(10) < 8 {
				d.plotOnFloor(j+4, j, tiles[d.roll(len(tiles))])
			}
		}
	}
	if !place.Has(rules.TraitSecondTileSet) {
		if d.wallNorth == wallPlain {
			for i := 4; i <= 6; i++ {
				if d.roll(10) < 3 {
					d.plotOnFloor(i, 2, d.roll(2)+51)
				}
			}
		}
		if d.wallEast != wallPlain {
			return
		}
		if d.squareVariant == 0 {
			eachEast(45, 54)
			return
		}
		for j := 2; j <= 4; j++ {
			if d.roll(10) < 5 {
				t := [3]int{46, 49, 50}[d.roll(3)]
				if d.roll(10) < 8 {
					d.plotOnFloor(j+4, j, t)
					if d.roll(10) < 5 {
						d.plotOnFloor(j+3, j, 47)
					}
				}
			} else if t := [6]int{45, 46, 48, 49, 53, 54}[d.roll(6)]; d.roll(10) < 8 {
				d.plotOnFloor(j+4, j, t)
			}
		}
		return
	}
	if d.wallNorth == wallPlain {
		if d.squareVariant != 0 {
			eachNorth(8, 59, 60, 62, 63, 66, 68)
		} else {
			eachNorth(8, 62, 66, 67)
		}
	}
	if d.wallEast != wallPlain {
		return
	}
	if d.squareVariant == 0 {
		eachNorth(8, 62, 66, 67) // the north row (i, 2), as the original does; the other east-wall branches use (j + 4, j)
		return
	}
	for j := 2; j <= 4; j++ {
		if d.roll(10) < 8 {
			switch t := [6]int{61, 62, 65, 66, 67, 63}[d.roll(6)]; t {
			case 63:
				d.plotOnFloor(j+4, j, 63)
				if d.roll(10) < 5 {
					d.plotOnFloor(j+3, j, 64)
				}
			default:
				d.plotOnFloor(j+4, j, t)
			}
		}
	}
}

// wallTile is a corner table cell: a tile, or noTile (255) to leave the cell.
const noTile = 255

// leftWallTiles is gdconsts.pas leftWallTiles[state 1..15][row 0..1][column 1..2] (index state - 1, column - 1).
var leftWallTiles = [15][2][2]int{
	{{24, 24}, {25, 25}}, {{36, noTile}, {28, 26}}, {{7, 24}, {9, 10}}, {{31, noTile}, {32, noTile}},
	{{24, 24}, {25, 25}}, {{14, noTile}, {13, 26}}, {{11, 24}, {13, 10}}, {{34, noTile}, {38, noTile}},
	{{17, 24}, {18, 25}}, {{26, noTile}, {28, 26}}, {{15, 24}, {28, 10}}, {{22, noTile}, {23, noTile}},
	{{21, 24}, {25, 25}}, {{16, noTile}, {13, 26}}, {{40, 24}, {13, 10}},
}

// rightWallTiles is gdconsts.pas rightWallTiles[state 1..15][row 0..1][column 5..6] (index state - 1, column - 5).
var rightWallTiles = [15][2][2]int{
	{{noTile, 31}, {noTile, 33}}, {{noTile, 35}, {noTile, 39}}, {{noTile, noTile}, {noTile, 8}}, {{24, 24}, {25, 25}},
	{{24, 24}, {25, 25}}, {{24, 24}, {25, 12}}, {{24, 24}, {25, 12}}, {{27, 29}, {noTile, 37}},
	{{27, 28}, {noTile, 27}}, {{27, 28}, {noTile, 27}}, {{27, 28}, {noTile, 27}}, {{19, 20}, {25, 25}},
	{{19, 20}, {25, 25}}, {{19, 20}, {25, 12}}, {{19, 20}, {25, 12}},
}

// layMapSquare is BuildDungeonArena's body for one map square: its walls' tiles, the corners, then the scatter.
func (d *dungeonArena) layMapSquare(walls rules.MapSides, partyAt rules.BoardPos, place rules.PlaceTraits) {
	d.layStructure(walls, partyAt)
	d.scatterSquare(walls, partyAt, place)
}

// scatterSquare is the map square's scatter: its wall kinds and variant, then ScatterWallTiles.
func (d *dungeonArena) scatterSquare(walls rules.MapSides, partyAt rules.BoardPos, place rules.PlaceTraits) {
	mapX, mapY := partyAt.X+d.ox, partyAt.Y+d.oy
	wall := func(x, y, dir int) int { return wallBetween(walls, partyAt.Y, x, y, dir) }
	d.wallWest, d.wallNorth, d.wallEast = wall(mapX, mapY, dirWest), wall(mapX, mapY, dirNorth), wall(mapX, mapY, dirEast)
	// GetMapEvent and $40; off the 16x16 map the original wraps (GetWallType).
	d.squareVariant = int(walls.MapCell((mapX+arenaMapSide)%arenaMapSide, (mapY+arenaMapSide)%arenaMapSide)) & squareVariantEv
	d.scatterWallTiles(place)
}

// layStructure lays the map square's walls' tiles and the corners (no random draw).
func (d *dungeonArena) layStructure(walls rules.MapSides, partyAt rules.BoardPos) {
	mapX, mapY := partyAt.X+d.ox, partyAt.Y+d.oy
	wall := func(x, y, dir int) int { return wallBetween(walls, partyAt.Y, x, y, dir) }
	d.wallWest, d.wallNorth, d.wallEast = wall(mapX, mapY, dirWest), wall(mapX, mapY, dirNorth), wall(mapX, mapY, dirEast)
	// PlaceWestWall
	switch d.wallWest {
	case wallPlain:
		for i := 2; i <= 4; i++ {
			d.plot(i-1, i, 27)
			d.plot(i, i, 28)
			d.plot(i+1, i, 26)
		}
	case wallDoor:
		for _, t := range [][3]int{{1, 2, 27}, {2, 2, 29}, {3, 2, 34}, {2, 3, 37}, {3, 3, 38}, {4, 4, 35}, {5, 4, 36}} {
			d.plot(t[0], t[1], t[2])
		}
	}
	// PlaceNorthWall
	if d.wallNorth == wallPlain {
		d.plot(3, 0, 24)
		d.plot(4, 0, 24)
		d.plot(3, 1, 25)
		d.plot(4, 1, 25)
	}
	// PlaceCornerTiles
	state := 0
	if d.wallNorth != wallNone {
		state++
	}
	if d.wallWest != wallNone {
		state += 2
	}
	if wall(mapX-1, mapY, dirNorth) != wallNone {
		state += 4
	}
	if wall(mapX, mapY-1, dirWest) != wallNone {
		state += 8
	}
	if state > 0 {
		for i := range 2 {
			for j := 1; j <= 2; j++ {
				if t := leftWallTiles[state-1][i][j-1]; t < noTile {
					d.plot(j, i, t)
				}
			}
		}
	}
	state = 0
	if wall(mapX+1, mapY, dirNorth) != wallNone {
		state++
	}
	if d.wallEast != wallNone {
		state += 2
	}
	if d.wallNorth != wallNone {
		state += 4
	}
	if wall(mapX, mapY-1, dirEast) != wallNone {
		state += 8
	}
	if state > 0 {
		for i := range 2 {
			for j := 5; j <= 6; j++ {
				if t := rightWallTiles[state-1][i][j-5]; t < noTile {
					d.plot(j, i, t)
				}
			}
		}
	}
	if d.wallNorth == wallDoor {
		d.plot(2, 0, 30)
		if d.wallWest == wallNone {
			d.plot(2, 1, 32)
		} else {
			d.plot(2, 1, 42)
		}
		if wall(mapX, mapY-1, dirEast) == wallNone {
			d.plot(5, 0, 31)
		} else {
			d.plot(5, 0, 43)
		}
		d.plot(5, 1, 33)
	}
	if wall(mapX, mapY-1, dirEast) == wallDoor {
		if d.wallNorth == wallNone {
			d.plot(5, 0, 39)
		} else {
			d.plot(5, 0, 41)
		}
	}
}

// The dungeon block's reach from the party: 6 columns and 2 rows of map squares each way.
const (
	dungeonReachX = 6
	dungeonReachY = 2
)

// buildDungeonArena is combatst.pas BuildDungeonArena(mode): the arena filled with floor (7 for the lotek kinds,
// else 45), tile groups scattered with the scatter trait, and each map square within 6 columns and 2 rows of the party
// laid from its walls. The generator is seeded with the ECL script's id (ReseedRandom(eclId)), so one script draws
// the same sequence: seed is that stream's start; the call's own stream is not touched (RestoreRandom).
// straight (Board.Straight) is the deskewed arena: the same squares with the skew taken out (rules.StraightX) and their
// walls one square thick, laid on the bare floor before the scatter (rules1e's squareWalls); the original's skewed pieces
// are not laid.
func buildDungeonArena(place rules.PlaceTraits, walls rules.MapSides, partyAt rules.BoardPos, seed uint32, straight bool) *[arenaRows][arenaCols]rules.TerrainID {
	d := &dungeonArena{rng: &rules.RNG{State: seed}, straight: straight}
	floor := rules.TerrainID(floorHitek)
	if place.Has(rules.TraitLowTech) {
		floor = floorLotek
	}
	for y := range d.cells {
		for x := range d.cells[y] {
			d.cells[y][x] = floor
		}
	}
	if place.Has(rules.TraitScatterTiles) { // ScatterRandomTiles
		for d.oy = -dungeonReachY; d.oy <= dungeonReachY; d.oy++ {
			for d.ox = -dungeonReachX; d.ox <= dungeonReachX; d.ox++ {
				if d.roll(2) == 0 {
					d.plot(4, 2, 55)
					d.plot(5, 2, 56)
					d.plot(5, 3, 57)
					d.plot(6, 3, 58)
				}
			}
		}
	}
	if straight { // the straight walls, then the scatter on the floor left (the draws keep the original's order); the original's
		// skewed wall pieces (layStructure, no draws) are not laid: off the straight walls' lines they stood as leftovers
		d.squareWalls(walls, partyAt, floor)
		d.eachSquare(func() { d.scatterSquare(walls, partyAt, place) })
		return &d.cells
	}
	d.eachSquare(func() { d.layMapSquare(walls, partyAt, place) })
	return &d.cells
}

// eachSquare runs lay for every map square within the reach, row by row, as the original's loops.
func (d *dungeonArena) eachSquare(lay func()) {
	for d.oy = -dungeonReachY; d.oy <= dungeonReachY; d.oy++ {
		for d.ox = -dungeonReachX; d.ox <= dungeonReachX; d.ox++ {
			lay()
		}
	}
}

// The straight board's wall tiles (terrain pv, as plot takes them): the skewed west wall's body, the north wall's top,
// and the post where walls end (leftWallTiles' state 15 corner), after the old engine's flatWest, flatNorth, flatPost.
const (
	straightWest  = 28
	straightNorth = 24
	straightPost  = 40
)

// squareWalls (extra, Deskew) lays the straight board's walls on the bare floor, before the scatter: one
// square thick with square corners. A square's west wall is its block's first column, its north wall the first row, a
// door the two middle squares between posts (rules1e's squareWalls, the same cells with this game's tiles).
func (d *dungeonArena) squareWalls(walls rules.MapSides, partyAt rules.BoardPos, floor rules.TerrainID) {
	set := func(x, y int, wall, doorway bool, pv int) {
		if x < 0 || x >= arenaCols || y < 0 || y >= arenaRows {
			return
		}
		switch {
		case doorway: // the open middle of a door: no piece of the original stands in it
			d.cells[y][x] = floor
		case wall: // the column is one tile (the old test's)
			d.cells[y][x] = rules.TerrainID(pv + 1)
		}
	}
	edge := func(x, y, dir int) int { return wallBetween(walls, partyAt.Y, x, y, dir) }
	gap := func(i int) bool { return i == 2 || i == 3 } // a door's squares, of the floor's 1..4
	for dr := -dungeonReachY; dr <= dungeonReachY; dr++ {
		for c := -4; c <= 5; c++ { // the board holds 10 squares
			x, y := partyAt.X+c, partyAt.Y+dr
			w, n := edge(x, y, dirWest), edge(x, y, dirNorth)
			x0, y0 := rules.StraightParty+c*rules.StraightCell, dr*blockRows+blockY0
			set(x0, y0, w != 0 || n != 0 || edge(x, y-1, dirWest) != 0 || edge(x-1, y, dirNorth) != 0, false, straightPost) // wherever a wall or door ends
			for i := 1; i < rules.StraightCell; i++ {
				set(x0+i, y0, n == wallPlain || n == wallDoor && !gap(i), n == wallDoor && gap(i), straightNorth)
				set(x0, y0+i, w == wallPlain || w == wallDoor && !gap(i), w == wallDoor && gap(i), straightWest)
			}
		}
	}
}
