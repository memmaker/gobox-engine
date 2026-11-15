package rulesxxvc

import "gobox/rules"

// wallBetween is combatst.pas WallBetween's kind from the map's sides: each side's GetWallKind or-ed. Off the 16x16
// map a side counts as 1, except east and west along the party's row (partyRow).
// ponytail: the MapSides give a wall (WallAt non-zero) and a door (DoorAt non-zero); GetWallKind's barrier plane
// (View3D.GetBarrierType 0 -> 1, else the wall type 0 -> 0, else 3) is taken as: no wall 0, a door 3, else 1.
func wallBetween(walls rules.MapSides, partyRow, x, y, dir int) int {
	side := func(cx, cy, facing int) int {
		if cx >= 0 && cx < arenaMapSide && cy >= 0 && cy < arenaMapSide {
			switch {
			case walls.WallAt(cx, cy, facing) == 0:
				return 0
			case walls.DoorAt(cx, cy, facing) != 0:
				return 3
			}
			return 1
		}
		if cy == partyRow && (facing == dirEast || facing == dirWest) {
			return 0
		}
		return 1
	}
	return side(x, y, dir) | side(x+dirStep[dir].X, y+dirStep[dir].Y, (dir+4)%8)
}

// Dungeon arena kinds (zone0 areaKind): lotek's floor is terrain 7, the others' 45; bit $40 scatters tile groups,
// bit $08 picks the second tile set (BuildDungeonArena, ScatterWallTiles).
const (
	AreaKindLotekA  = 0x88
	AreaKindLotekB  = 0x82
	floorLotek      = 7
	floorHitek      = 45
	kindScatterBit  = 0x40
	kindSecondSet   = 0x08
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
}

// plot is combatst.pas PlotCell: terrain pv + 1 at (px, py) of the current map square's block, inside the arena.
func (d *dungeonArena) plot(px, py, pv int) {
	col, row := d.ox*blockCols+d.oy*blockRows+blockX0+px, d.oy*blockRows+blockY0+py
	if col >= 0 && col < arenaCols && row >= 0 && row < arenaRows {
		d.cells[row][col] = rules.TerrainID(byte(pv + 1))
	}
}

// plotOnFloor is ScatterWallTiles' PlaceTileOn: PlotCell, onto a floor cell (terrain 7 or 45) only. The original's
// first upper-bound test reads row where col is meant; a column past the arena then reads another row's cell and
// PlotCell's own test plots nothing, so it is skipped here.
func (d *dungeonArena) plotOnFloor(px, py, pv int) {
	col, row := d.ox*blockCols+d.oy*blockRows+blockX0+px, d.oy*blockRows+blockY0+py
	if col >= 0 && col < arenaCols && row >= 0 && row < arenaRows {
		if t := d.cells[row][col]; t == floorLotek || t == floorHitek {
			d.plot(px, py, pv)
		}
	}
}

func (d *dungeonArena) roll(n int) int { return d.rng.N(n) }

// scatterWallTiles is combatst.pas ScatterWallTiles(mode): tiles at random along the north and east walls of the
// square being laid, Random's draws in the original's order.
func (d *dungeonArena) scatterWallTiles(mode int) {
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
	if mode&kindSecondSet == 0 {
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
func (d *dungeonArena) layMapSquare(walls rules.MapSides, partyAt rules.BoardPos, mode int) {
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
	// GetMapEvent and $40; off the 16x16 map the original wraps (GetWallType).
	d.squareVariant = int(walls.MapCell((mapX+arenaMapSide)%arenaMapSide, (mapY+arenaMapSide)%arenaMapSide)) & squareVariantEv
	d.scatterWallTiles(mode)
}

// The dungeon block's reach from the party: 6 columns and 2 rows of map squares each way.
const (
	dungeonReachX = 6
	dungeonReachY = 2
)

// buildDungeonArena is combatst.pas BuildDungeonArena(mode): the arena filled with floor (7 for kinds $88 and $82,
// else 45), tile groups scattered when mode has bit $40, and each map square within 6 columns and 2 rows of the party
// laid from its walls. The generator is seeded with the ECL script's id (ReseedRandom(eclId)), so one script draws
// the same sequence: seed is that stream's start; the call's own stream is not touched (RestoreRandom).
func buildDungeonArena(mode int, walls rules.MapSides, partyAt rules.BoardPos, seed uint32) *[arenaRows][arenaCols]rules.TerrainID {
	d := &dungeonArena{rng: &rules.RNG{State: seed}}
	floor := rules.TerrainID(floorHitek)
	if mode == AreaKindLotekA || mode == AreaKindLotekB {
		floor = floorLotek
	}
	for y := range d.cells {
		for x := range d.cells[y] {
			d.cells[y][x] = floor
		}
	}
	if mode&kindScatterBit != 0 { // ScatterRandomTiles
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
	for d.oy = -dungeonReachY; d.oy <= dungeonReachY; d.oy++ {
		for d.ox = -dungeonReachX; d.ox <= dungeonReachX; d.ox++ {
			d.layMapSquare(walls, partyAt, mode)
		}
	}
	return &d.cells
}
