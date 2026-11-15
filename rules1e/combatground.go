package rules1e

import "gobox/rules"

// The straight board (extra, Deskew): a cell is a square block of
// straightCell squares, its west wall the first column, its north wall the
// first row, its floor 4x4 as the original's; the party's cell starts at
// column straightParty, so the board holds 10 cells by 5.
const (
	straightCell  = 5
	straightParty = 20
)

// StraightX is the column on the straight board of square (x, y) of the
// original's skewed dungeon map: the skew taken out (a square west for each
// row down), then the original's 6-wide block made 5: its two wall columns
// are the one. Floor squares keep their own square; two squares of a wall
// line open to the west share one.
func StraightX(x, y int) int {
	sx := x - (y - 0xA) - 0x15 + 60 // unskewed, from the block 10 cells west of the party's
	return straightParty + (sx/6-10)*straightCell + max(sx%6-1, 0)
}

// squareWalls (extra, Deskew) lays the straight board's walls over the
// original's pieces, before anyone is placed: one square thick with square
// corners (the original's west wall is two wide, its door posts and wall
// ends a square apart along the slant). A cell's west wall is its block's
// first column, its north wall the first row, a door the two middle squares
// between posts.
func (r *Rules) squareWalls(m rules.MapSides) {
	const west, north, corner = 3, 5, 2 // the pieces' tiles (ovr11:02BA, :0377, :040F)
	floor := r.S.grounds().Floor
	set := func(x, y int, wall bool, tile byte) {
		if x < 0 || x > 0x31 || y < 0 || y > 0x18 {
			return
		}
		switch is := r.S.GroundInfo(r.S.Cmap.Ground[y][x])[0] == 0xFF; {
		case wall && !is:
			r.S.Cmap.Ground[y][x] = Ground(tile + 1)
		case !wall && is:
			r.S.Cmap.Ground[y][x] = floor
		}
	}
	gap := func(i int) bool { return i == 2 || i == 3 } // a door's squares, of the floor's 1..4
	for dr := -2; dr <= 2; dr++ {
		for c := -4; c <= 5; c++ {
			x, y := r.S.PosX+c, r.S.PosY+dr
			w, n := CellEdge(m, r.S.PosY, x, y, 6), CellEdge(m, r.S.PosY, x, y, 0)
			x0, y0 := straightParty+c*straightCell, dr*5+0xA
			set(x0, y0, w != 0 || n != 0 || CellEdge(m, r.S.PosY, x, y-1, 6) != 0 || CellEdge(m, r.S.PosY, x-1, y, 0) != 0, corner) // wherever a wall or door ends
			for i := 1; i < straightCell; i++ {
				set(x0+i, y0, n == 1 || n == 3 && !gap(i), north)
				set(x0, y0+i, w == 1 || w == 3 && !gap(i), west)
			}
		}
	}
}

// boardX is the board's column of square (x, y) of the original's map.
func (r *Rules) boardX(x, y int) int {
	if r.straight {
		return StraightX(x, y)
	}
	return x
}

// dungeonMap = ovr11:0810: every cell within 6 east-west and 2 north-south
// of the party becomes a 6x5 block of tiles, rows skewed 5 squares (extra,
// Deskew: the same squares unskewed, boardX).
func (r *Rules) dungeonMap(g rules.BuildGround) {
	gen := &r.S.T().Gen
	if gen.Fill != 0 {
		for y := range r.S.Cmap.Ground {
			for x := range r.S.Cmap.Ground[y] {
				r.S.Cmap.Ground[y][x] = gen.Fill
			}
		}
	}
	for dr := -2; dr <= 2; dr++ {
		for c := -6; c <= 6; c++ {
			x, y := r.S.PosX+c, r.S.PosY+dr
			w, n, ea := CellEdge(g.Walls, r.S.PosY, x, y, 6), CellEdge(g.Walls, r.S.PosY, x, y, 0), CellEdge(g.Walls, r.S.PosY, x, y, 2)
			put := func(col, row int, t byte) { // ovr11:0221
				py := row + dr*5 + 0xA
				px := r.boardX(col+c*6+dr*5+0x15, py)
				if px >= 0 && px <= 0x31 && py >= 0 && py <= 0x18 {
					r.S.Cmap.Ground[py][px] = Ground(t + 1)
				}
			}
			// nested ovr11:02BA: floor, west wall or door
			for row := 2; row <= 4; row++ {
				for col := 0; col <= 5; col++ {
					put(col, row, 0x16)
				}
			}
			if w == 1 {
				for i := 2; i <= 4; i++ {
					put(i-1, i, 4)
					put(i, i, 3)
					put(i+1, i, 0xD)
				}
			} else if w == 3 {
				put(1, 2, 8)
				put(5, 4, 0)
			}
			// nested ovr11:0377: north wall
			if n == 1 {
				put(3, 0, 5)
				put(4, 0, 5)
				put(3, 1, 0xA)
				put(4, 1, 0xA)
			} else {
				put(3, 0, 0x16)
				put(4, 0, 0x16)
				put(3, 1, 0x16)
				put(4, 1, 0x16)
			}
			// nested ovr11:040F: the north-west corner
			nw := CellEdge(g.Walls, r.S.PosY, x, y-1, 6) == 0 && CellEdge(g.Walls, r.S.PosY, x-1, y, 0) == 0
			pick := func(ok bool, a, b byte) byte {
				if ok {
					return a
				}
				return b
			}
			var a, b, cc, d byte
			if n == 0 {
				switch w {
				case 0:
					a = 0x16
				case 3:
					a = pick(nw, gen.Corners.NWDoorOpen, 0xD)
				case 1:
					a = pick(nw, 0, 0xD)
				}
			} else if w == 0 {
				a = pick(nw, 0xF, 5)
			} else {
				a = pick(nw, 0x12, 2)
			}
			b = map[byte]byte{0: 0x16, 3: 0x11, 1: 5}[n]
			switch w {
			case 0:
				cc = pick(n == 0, 0x16, pick(nw, 0x10, 0xA))
			case 3:
				cc = pick(nw, 0x14, 7)
			case 1:
				cc = pick(nw, 1, 3)
			}
			if w == 0 || w == 3 {
				d = map[byte]byte{0: 0x16, 3: 0x17, 1: 0xA}[n]
			} else {
				d = map[byte]byte{0: 0xD, 3: 0x15, 1: 6}[n]
			}
			put(1, 0, a)
			put(2, 0, b)
			put(1, 1, cc)
			put(2, 1, d)
			// nested ovr11:0615: the north-east corner
			ne7, ne8 := CellEdge(g.Walls, r.S.PosY, x, y-1, 2), CellEdge(g.Walls, r.S.PosY, x+1, y, 0)
			ne := ne7 == 0 && ne8 == 0
			a = map[byte]byte{0: pick(ne7 == 1, 4, 0x16), 3: 0xF, 1: 5}[n]
			if n == 0 {
				switch ne7 {
				case 0:
					b = 0x16
				case 3:
					b = pick(ea == 0 && ne8 != 0, 0x18, pick(ea == 0, gen.Corners.NEDoorOpen, 1))
				case 1:
					b = pick(ea == 0 && ne8 != 0, 0xB, pick(ea == 0, gen.Corners.NEWallOpen, 3))
				}
			} else {
				switch {
				case ea != 0:
					b = 9
				case ne8 != 0:
					b = 5
				default:
					b = pick(ne, 0x11, 0x13)
				}
			}
			cc = map[byte]byte{0: 0x16, 3: 0x10, 1: 0xA}[n]
			if n == 0 {
				switch {
				case ne7 == 0:
					d = 0x16
				case ea != 0:
					d = 4
				default:
					d = pick(ne8 == 0, 8, 0xC)
				}
			} else {
				switch {
				case ea != 0:
					d = 0xE
				default:
					d = pick(ne8 == 0, 0x17, 0xA)
				}
			}
			put(5, 0, a)
			put(6, 0, b)
			put(5, 1, cc)
			put(6, 1, d)
			if gen.Rubble != nil {
				r.dungeonRubble(gen.Rubble, g.Walls.MapCell(x, y)&gen.Rubble.CellFlag != 0, c, dr, n, ea, CellEdge(g.Walls, r.S.PosY, x, y, 4), w)
			}
		}
	}
	if r.straight {
		r.squareWalls(g.Walls)
	}
	r.floorClutter(g)
}

// floorClutter = ovr11:0000: now and then something lies on open ground.
func (r *Rules) floorClutter(g rules.BuildGround) {
	cl := r.S.T().Gen.Clutter
	if cl == nil {
		return
	}
	for x := 0; x <= 0x31; x++ {
		for y := 0; y <= 0x18; y++ {
			cell := &r.S.Cmap.Ground[y][x]
			if r.S.GroundInfo(*cell)[3] != GroundProp(cl.Tile) {
				continue
			}
			switch r.S.Dice(1, 100) {
			case 0x62:
				*cell = cl.Rare[0]
			case 0x63:
				*cell = cl.Rare[1]
			case 0x64:
				if !g.NoRemains && r.S.Dice(1, 100) == 1 {
					if roll := r.S.Dice(1, 10); roll <= 6 {
						*cell = cl.Remains[0]
					} else {
						*cell = cl.Remains[1]
					}
				}
			}
		}
	}
}

// flagValue is the value of the last entry of list whose flag the terrain
// has, else def.
func flagValue(f TerrainFlag, def int, list []FlagValue) int {
	for _, fv := range list {
		if f&fv.Flag != 0 {
			def = fv.Value
		}
	}
	return def
}

// TerrainFlags = ovr11:08A4: what the wilderness square holds (the game's
// Tables.Terrain by the fight's terrain; past its end the original reads
// what follows it, here 0).
func (e *RuleState) TerrainFlags(boss uint16) TerrainFlag {
	var k TerrainFlag
	if t := e.T().Terrain; int(e.Terrain) < len(t) {
		k = t[e.Terrain]
	}
	if w := e.T().Gen.Wild; w != nil && w.BossHides && k&TerrainBare != 0 && boss == 0xFF {
		k = k - TerrainBare | 1
	}
	return k
}

// openAt: the square's tile is the open one, where things may be put.
func (r *Rules) openAt(w *WildGen, x, y int) bool {
	return r.S.GroundInfo(r.S.Cmap.Ground[y][x])[3] == GroundProp(w.Open)
}

// wildMap = ovr11:11B4: the grass, then the game's passes in order.
// wildMap reads BuildGround's Walls, Dungeon, MapShift, NoRemains, Terrain, River and MapX/MapY/Boss; its Block and Place
// (Buck Rogers' arena seed and area kind) are XXVc's and ignored here.
func (r *Rules) wildMap(g rules.BuildGround) {
	w := r.S.T().Gen.Wild
	if w == nil {
		return
	}
	for y := range r.S.Cmap.Ground {
		for x := range r.S.Cmap.Ground[y] {
			r.S.Cmap.Ground[y][x] = w.Grass
		}
	}
	switch w.Terrain {
	case TerrainFromMap:
		a, b := int(byte(g.MapX)), int(byte(g.MapY))
		a += g.MapShift
		r.S.Terrain = Terrain(byte(TerrainMap[b][a])) // DS:465A
	case TerrainFromGiven:
		r.S.Terrain = Terrain(g.Terrain)
	default:
		r.S.Terrain = 0
	}
	r.strip(w.Strip, g)
	r.groves(w)
	r.trees(w)
	r.plants(w)
	r.scatter(w)
	r.floorClutter(g)
}

// WildRoad = ovr11:0BDE: POR's strip with no engine terms.
func (r *Rules) WildRoad() { r.strip(r.S.T().Gen.Wild.Strip, rules.BuildGround{}) }

// strip lays the road, cliff or river down the map: it starts at a column
// 34 - 5d4, moved west to one of the table's, and leans one column east a row.
func (r *Rules) strip(sg *StripGen, g rules.BuildGround) {
	if sg == nil {
		return
	}
	switch sg.Style {
	case StripRoad:
		r.road(sg)
	case StripCliff:
		r.cliff(sg)
	case StripRiver:
		r.river(sg, g.River)
	}
}

// lean is the next row's column: east, unless the board is deskewed.
func (r *Rules) lean(col int) int {
	if r.straight {
		return col
	}
	return col + 1
}

// road: a diagonal road (0x32/0x33), maybe with stones.
func (r *Rules) road(sg *StripGen) {
	var stones int
	if r.S.Dice(1, 100) > flagValue(r.S.TerrainFlags(r.boss), 0, sg.Chance) {
		return
	}
	col := int(int8(byte(0x22 - r.S.Dice(5, 4))))
	for int(int8(byte(col+2)))%7 > 0 {
		col--
	}
	start := col
	// pair sets the square and its right neighbour. fixed, as the Pascal reference: at column 0x31 the
	// original wrote the neighbour into the next row's column 0 (a stray bank
	// tile in the bottom-left corner); off the map it is dropped
	pair := func(row, col int, left, right Ground) {
		r.S.Cmap.Ground[row][col] = left
		if col+1 <= 0x31 {
			r.S.Cmap.Ground[row][col+1] = right
		}
	}
	stone := func(row, col int) { // ovr11:0B97
		pair(row, col, sg.Tiles[2], sg.Tiles[3])
		stones++
	}
	for row := 0; row <= 0x18; row++ {
		if col > 0x31 {
			continue
		}
		pair(row, col, sg.Tiles[0], sg.Tiles[1])
		if r.S.Dice(1, 20) == 1 || stones&1 != 0 {
			stone(row, col)
		}
		col = r.lean(col) // the road leans with the perspective
	}
	if stones == 0 {
		row := 0xC - r.S.Dice(1, 9) + 4
		stone(row, row+start)
		stone(row+1, row+start+1)
	}
}

// cliff (CotAB O05D:0A00): a cliff face down the map with a fallen block at
// its foot here and there.
func (r *Rules) cliff(sg *StripGen) {
	if r.S.Dice(1, 100) > flagValue(r.S.TerrainFlags(r.boss), 0, sg.Chance) {
		return
	}
	g := &r.S.Cmap.Ground
	x := 34 - r.S.Dice(5, 4)
	for (x+2)%7 > 0 {
		x--
	}
	for y := range g {
		if x > len(g[y])-1 {
			continue
		}
		g[y][x] = sg.Tiles[0] + Ground(r.S.Dice(1, 2))
		if x < len(g[y])-1 {
			g[y][x+1] = sg.Tiles[1] + Ground(r.S.Dice(1, 2))
		}
		if r.S.Dice(1, 20) == 1 { // O05D:09AC
			if x < len(g[y])-1 {
				g[y][x+1] = sg.Tiles[2]
			}
			if y < len(g)-1 && x < len(g[y])-1 {
				g[y+1][x+1] = sg.Tiles[3]
			}
		}
		x = r.lean(x)
	}
}

// river (PoD O065:0803): when the engine says so a river runs down the map
// with fords getting likelier the longer the water runs.
func (r *Rules) river(sg *StripGen, wanted bool) {
	if !wanted {
		return
	}
	bank, water, ford, fordEnd := sg.Tiles[0], sg.Tiles[1], sg.Tiles[2], sg.Tiles[3]
	g := &r.S.Cmap.Ground
	put := func(y, x int, t Ground) {
		if x < len(g[y]) {
			g[y][x] = t
		}
	}
	x := 34 - r.S.Dice(5, 4)
	for (x+2)%7 > 0 {
		x--
	}
	chance := 1
	for y := 0; y < len(g) && x <= len(g[y])-1; y++ {
		put(y, x, bank)
		if x < len(g[y])-1 {
			if r.S.Dice(1, 20) <= chance {
				put(y, x+1, ford)
				put(y, x+2, ford+Ground(r.S.Dice(1, 2)))
				put(y, x+3, ford+Ground(r.S.Dice(1, 2)))
				put(y, x+4, fordEnd)
				chance = 1
			} else {
				chance++
				put(y, x+1, bank+Ground(r.S.Dice(1, 2)))
				put(y, x+2, water)
				put(y, x+3, water)
				put(y, x+4, water+Ground(r.S.Dice(1, 2)))
			}
		}
		put(y, x+5, bank)
		x = r.lean(x)
	}
}

// groves = ovr11:0D13 (POR): trees one or two squares high on open ground.
func (r *Rules) groves(w *WildGen) {
	gg := w.Groves
	if gg == nil {
		return
	}
	chance := flagValue(r.S.TerrainFlags(r.boss), gg.Default, gg.Chance)
	g := &r.S.Cmap.Ground
	for x := 0; x <= 0x31; x++ {
		for y := 1; y <= 0x18; y++ {
			if !r.openAt(w, x, y) || !r.openAt(w, x, y-1) || r.S.Dice(1, 100) > chance {
				continue
			}
			v := r.S.Dice(1, 10)
			if v > 8 {
				v = r.S.Dice(1, 2) + 4
			} else {
				v = (v + 1) >> 1
			}
			if r.S.TerrainFlags(r.boss)&gg.RoughFlag != 0 {
				v = r.S.Dice(1, 3) + 3
			}
			g[y][x] = gg.Base + Ground(v)
			if v <= groveTopSizes {
				g[y-1][x] = gg.Top + Ground(v)
			}
		}
	}
}

// trees (CotAB O05D:0B0B, PoD O065:09A1): on open ground under open ground,
// now and then a shrub or a tree (its top on the square above).
func (r *Rules) trees(w *WildGen) {
	tg := w.Trees
	if tg == nil {
		return
	}
	t := r.S.TerrainFlags(r.boss)
	if t&tg.Bare != 0 {
		return
	}
	chance := flagSum(t, tg.Base, tg.Adjust)
	if chance < 0 {
		chance = 1
	}
	g := &r.S.Cmap.Ground
	for x := range g[0] {
		for y := 1; y < len(g); y++ {
			if !r.openAt(w, x, y) || !r.openAt(w, x, y-1) || chance < r.S.Dice(1, 100) {
				continue
			}
			if chance >= r.S.Dice(1, 100) {
				g[y][x] = tg.Shrub + Ground(r.S.Dice(1, tg.ShrubN))
			} else {
				g[y-1][x] = tg.Top + Ground(r.S.Dice(1, tg.TreeN))
				g[y][x] = tg.Trunk + Ground(r.S.Dice(1, tg.TreeN))
			}
		}
	}
}

// plants = ovr11:0E9E (POR): trees, bushes and the like on open ground.
func (r *Rules) plants(w *WildGen) {
	pg := w.Plants
	if pg == nil {
		return
	}
	odds := pg.Odds
	f := r.S.TerrainFlags(r.boss)
	for _, pf := range pg.Flags {
		if f&pf.Flag != 0 {
			odds[pf.Index] = pf.Value
		}
	}
	p1, p2, p3, p4, p5 := odds[0], odds[1], odds[2], odds[3], odds[4]
	t := pg.Tiles
	g := &r.S.Cmap.Ground
	for x := 0; x <= 0x31; x++ {
		for y := 0; y <= 0x18; y++ {
			if !r.openAt(w, x, y) {
				continue
			}
			n := r.S.Dice(1, pg.Die)
			switch {
			case n <= p1:
				v := r.S.Dice(1, 4)
				if v == 4 && y > 0 && r.openAt(w, x, y-1) {
					g[y-1][x], g[y][x] = t.Top, t.Bottom
				}
				if v < 4 {
					g[y][x] = t.Tall + Ground(v)
				}
			case n <= p1+p2:
				g[y][x] = Ground(r.S.Dice(1, 3)) + t.Low
			case n <= p3+p1+p2:
				g[y][x] = Ground(r.S.Dice(1, 4)) + t.Mid
			case n <= p5+p3+p1+p2:
				g[y][x] = Ground((r.S.Dice(1, 10)-1)/3+1) + t.Bush
			case n <= p4+p5+p3+p1+p2:
				v := r.S.Dice(1, 4)
				if v == 4 && f&t.RoughFlag == 0 {
					v = r.S.Dice(1, 3)
				}
				g[y][x] = Ground(v) + t.Fruit
			}
		}
	}
}

// scatter (CotAB O05D:0E4A, PoD O065:0CB6): the small things on open
// ground, by how rough the terrain is.
func (r *Rules) scatter(w *WildGen) {
	sg := w.Scatter
	if sg == nil {
		return
	}
	rough := flagSum(r.S.TerrainFlags(r.boss), sg.Start, sg.Rough)
	var percent []int
	for _, o := range sg.Odds {
		if rough >= o.Lo && rough <= o.Hi {
			percent = o.Percent
			break
		}
	}
	if percent == nil {
		return
	}
	g := &r.S.Cmap.Ground
	for x := range g[0] {
		for y := range g {
			if !r.openAt(w, x, y) {
				continue
			}
			n, upTo := r.S.Dice(1, 100), 0
			for i, k := range sg.Kinds {
				upTo += percent[i]
				if n <= upTo {
					g[y][x] = k.Base
					if k.N > 0 {
						g[y][x] += Ground(r.S.Dice(1, k.N))
					}
					break
				}
			}
		}
	}
}

// flagSum is start plus the values of every flag the terrain has.
func flagSum(f TerrainFlag, start int, list []FlagValue) int {
	for _, fv := range list {
		if f&fv.Flag != 0 {
			start += fv.Value
		}
	}
	return start
}

// dungeonRubble = CotAB O05D:00D3 (for the cell at column c, row dr with the
// edges n, e, s, w): in a cell flagged for it that is neither a corridor
// nor has a door, heaps of rubble on the floor.
func (r *Rules) dungeonRubble(rb *RubbleGen, flagged bool, c, dr int, n, ea, s, w byte) {
	room := (n == 1 || ea == 1 || s == 1 || w == 1) &&
		(n != 1 || s != 1 || ea == 1 && w == 1) &&
		(ea != 1 || w != 1 || n == 1 && s == 1) &&
		n != 3 && ea != 3 && s != 3 && w != 3
	g := &r.S.Cmap.Ground
	inMap := func(x, y int) bool { return y >= 0 && y < len(g) && x >= 0 && x < len(g[y]) }
	open := func(x, y int) bool { return r.S.GroundInfo(g[y][x])[3] == GroundProp(rb.Tile) }
	for i := 2; i <= 3; i++ {
		for j := 2; j <= 4; j++ {
			y := dr*5 + 10 + j
			x := r.boardX(c*6+dr*5+21+i+j, y) // sic: the row is added to the column too
			if !inMap(x, y) || !open(x, y) || !flagged || !room || r.S.Dice(1, 10) > rb.HeapOdds {
				continue
			}
			g[y][x] = rb.Heap
			for _, d := range rubbleSpread {
				nx, ny := x+d[0], y+d[1]
				if inMap(nx, ny) && open(nx, ny) && r.S.Dice(1, 10) <= rb.EdgeOdds {
					g[ny][nx] = rb.Edge
				}
			}
		}
	}
}

// rubbleSpread = the four squares round a rubble heap.
var rubbleSpread = [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

// buildGround = combatMapSetup's map half (ovr11:1244): a fresh combat map,
// the dungeon's walls or the wilderness' terrain, then clutter.
func (r *Rules) buildGround(g rules.BuildGround) {
	r.S.Cmap = &CombatMap{}
	r.boss = g.Boss // TerrainFlags' Boss word, for this call
	defer func() { r.boss = 0 }()
	r.S.Dungeon = g.Dungeon
	if g.Dungeon {
		r.dungeonMap(g)
	} else {
		r.wildMap(g)
	}
}
