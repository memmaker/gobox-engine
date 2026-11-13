package rules1e

import (
	"gobox/core"
	"gobox/core/goldbox"
)

// The combat map (ovr31, ovr32, ovr33). The original drew 7x7 squares of
// the 50x25 map on its drawing page, 3x3 cells a square; that page is
// gone (the battle view draws the map from the snapshot, battle.go), but
// its scrolling window stays as what the view looks at (Battle.View).

// CombatMap is the 0x4E9-byte record at [DS:6716].
type CombatMap struct {
	// +2/+3 the view's top left square: the Engine's (presentation); combatOnScreen takes it as input
	// +4/+5 the cursor (icon, footprint): the Engine's (presentation)
	seeAll byte           // +6 line of sight ignores walls
	Ground [25][50]Ground // +7 ground type per square
}

// CombatPos is the 4-byte DS:5F27 record of combatant i.
type CombatPos struct{ X, Y, Id, Size byte }

// Body is a 7-byte DS:66D6 record: who fell where, and the ground it hid.
type Body struct {
	C      *Character
	X, Y   byte
	Ground byte
}

// Footprint = ovr32:0000 (size, part, var dx, var dy): offset of one square
// of a combatant that covers several (DS:2844).
func (e *RuleState) Footprint(size, part int) (int, int, bool) {
	if size == 0 {
		return 0, 0, false
	}
	b := footprintTable[size][part]
	return int(int8(b[0])), int(int8(b[1])), int8(b[0]) >= 0
}

// GroundInfo = DS:2744: 4 bytes per ground type: move cost (0xFF wall),
// view height, blocking height, tile.
//
// Past the table the original reads the data after it; CotAB's 64 records
// are read round (its ports' groundInfo[g%64]: the cliff foot 64 and 65 are
// records 0 and 1), as every game's table is.
func (e *RuleState) GroundInfo(g Ground) []GroundProp {
	t := e.T().Ground
	return t[int(g)%len(t)]
}

// TileOf is the tile of map square (x, y). The original indexes the 50-wide
// grid flat, so a square just off one edge showed the far edge's pieces; here
// off the map is bare floor (ground 0 would be a wall piece).
func (e *RuleState) TileOf(x, y int) GroundProp {
	if x < 0 || y < 0 || x >= 50 || y >= 25 { // off the map (the wide page shows past the edges): bare floor, never wrapped
		return e.GroundInfo(e.grounds().Floor)[3]
	}
	return e.GroundInfo(e.Cmap.Ground[y][x])[3]
}

// CombatRedrawMap = ovr32:03A0: rebuild the occupancy grid (DS:60DB).
func (e *RuleState) CombatRedrawMap() {
	e.Occ = [25][50]Slot{}
	for i := 1; i <= e.Ncomb; i++ {
		p := e.Cpos[i]
		if p.Size == 0 {
			continue
		}
		for part := 0; part <= 3; part++ {
			if dx, dy, ok := e.Footprint(int(p.Size), part); ok {
				x, y := int(p.X)+dx, int(p.Y)+dy
				if x >= 0 && x < 50 && y >= 0 && y < 25 {
					e.Occ[y][x] = Slot(i)
				}
			}
		}
	}
}

// MapAt = ovr32:04BE (x, y, var who, var ground).
func (e *RuleState) MapAt(x, y int) (Slot, Ground) {
	if x < 0 || x > 0x31 || y < 0 || y >= len(e.Occ) {
		return 0, 0
	}
	return e.Occ[y][x], e.Cmap.Ground[y][x]
}

// TurnTo is c facing dir, standing: the rule's part of combatTurn (pose 0,
// shown); present draws it (Turned).
func (e *RuleState) TurnTo(c *Character, dir byte) {
	c.Combat.Facing = dir
	e.Emit(core.Turned{Unit: unit(e.CombatIndex(c)), Facing: core.Facing(dir)})
}

func (e *RuleState) CombatX(c *Character) int    { return int(e.Cpos[e.CombatIndex(c)].X) }    // ovr32:0BE7
func (e *RuleState) CombatY(c *Character) int    { return int(e.Cpos[e.CombatIndex(c)].Y) }    // ovr32:0C0F
func (e *RuleState) CombatSize(c *Character) int { return int(e.Cpos[e.CombatIndex(c)].Size) } // ovr32:0C37

// CombatIndex = ovr32:0C5F (char): its index in DS:65B9, 0 if none.
func (e *RuleState) CombatIndex(c *Character) int {
	for i := 1; i <= e.Ncomb; i++ {
		if e.Combatant(byte(i)) == c {
			return i
		}
	}
	return 0
}

// GroundAt = ovr32:0CB9 (char, dir, var who, var ground): who stands on
// and the dearest ground of the squares the combatant would cover one step
// in direction dir (8: where it stands). 0 = off the map; a stinking cloud wins.
func (e *RuleState) GroundAt(c *Character, dir int) (byte, Ground) {
	var who byte
	ids := e.grounds()
	gr := ids.Floor
	best := byte(1)
	idx := e.CombatIndex(c)
	for part := 0; part <= 3; part++ {
		dx, dy, ok := e.Footprint(int(e.Cpos[idx].Size), part)
		if !ok {
			continue
		}
		o, g := e.MapAt(int(e.Cpos[idx].X)+dx+DirDX[dir], int(e.Cpos[idx].Y)+dy+DirDY[dir])
		if int(o) == idx {
			o = 0
		}
		if o > 0 {
			who = byte(o)
		}
		switch {
		case g == 0 || gr == 0:
			gr = 0
		case g == ids.StinkingCloud || gr == ids.StinkingCloud:
			gr = ids.StinkingCloud
		case byte(e.GroundInfo(g)[0]) >= best:
			best = byte(e.GroundInfo(g)[0])
			gr = g
		}
	}
	return who, gr
}

// combatRemoveFromMap = ovr32:0E11 (char): a fallen combatant leaves the
// map; unless summoned, a body (GroundIDs.Body) stays where it fell.
func (e *RuleState) combatRemoveFromMap(c *Character) {
	e.combatRemoveUnder(c, c)
}

// combatRemoveUnder: as combatRemoveFromMap, the body lying on at's square
// (a swallower's, sotsb O0A4:1C26).
func (e *RuleState) combatRemoveUnder(c, at *Character) {
	if e.BeginFlag != 5 {
		e.Sound(e.fallEvent(c))
		return
	}
	if int(e.bodyIndex) < len(e.Bodies) && e.Bodies[e.bodyIndex].C == c {
		return // the original's search (DS:70C7, not its loop index)
	}
	idx := e.CombatIndex(c)
	x, y := e.CombatX(at), e.CombatY(at)
	e.factFell(c)
	if c.Combat.Extra == 0 && e.Nbodies+1 < len(e.Bodies) {
		e.Nbodies++
		b := &e.Bodies[e.Nbodies]
		b.Ground = byte(e.Cmap.Ground[y][x])
		if e.Cmap.Ground[y][x] != e.grounds().StinkingCloud {
			e.Cmap.Ground[y][x] = e.grounds().Body
		}
		b.C, b.X, b.Y = c, byte(x), byte(y)
	}
	e.Sound(e.fallEvent(c))
	e.Cpos[idx].Size = 0
	e.CombatRedrawMap()
	ci := c.Combat
	ci.Delay, ci.Moves, ci.Spell, ci.Guarding = 0, 0, 0, 0
}

// combatSquareFree = ovr32:10A7 (char, x, y, uncover): put the combatant on
// (x,y) if every square it needs is free ground; with uncover, its body is
// taken off the map.
func (e *RuleState) combatSquareFree(c *Character, x, y int, uncover bool) bool {
	if e.BeginFlag != 5 {
		return true
	}
	idx := e.CombatIndex(c)
	p := &e.Cpos[idx]
	p.Size = c.Size & 0x7F
	p.X, p.Y = byte(x), byte(y)
	who, g := e.GroundAt(c, 8)
	if who != 0 || g == 0 || e.GroundInfo(Ground(g))[0] == 0xFF {
		p.Size = 0
		return false
	}
	if uncover && c.Combat.Extra == 0 {
		for k := 1; k <= e.Nbodies; k++ {
			if e.Bodies[k].C == c {
				if Ground(e.Bodies[k].Ground) != e.grounds().Body {
					g = Ground(e.Bodies[k].Ground)
				}
				e.Bodies[k] = Body{}
			}
		}
		covered := false
		for k := 1; k <= e.Nbodies; k++ {
			if e.Bodies[k].C != nil && int(e.Bodies[k].X) == x && int(e.Bodies[k].Y) == y {
				covered = true
			}
		}
		if !covered {
			e.Cmap.Ground[y][x] = Ground(g)
		}
	}
	e.CombatRedrawMap()
	return true
}

// nearSort = ovr31:002E: order the near list by distance.
func (e *RuleState) nearSort() {
	goldbox.NearSort(e.Near, func(n NearEntry) (byte, byte) { return n.dist, n.dir })
}

// LineOfSight = ovr31:0422 (map, var range, var tx, var ty, x, y): walk from
// (x,y) towards (tx,ty) until the target, a wall higher than the line or
// the range (in half squares) stops it.
func (e *RuleState) LineOfSight(rng *int, tx, ty *int, x, y int) bool {
	s := goldbox.Sight{SeeAll: e.Cmap.seeAll != 0, Height: func(x, y int) (int, int) {
		g := e.GroundInfo(e.Cmap.Ground[y][x])
		return int(g[1]), int(g[2])
	}}
	return goldbox.LineOfSight(s, rng, tx, ty, x, y)
}

// InArc = ovr31:057B (x1, y1, x2, y2, dir): (x2,y2) lies in the quarter
// facing dir from (x1,y1) (dir 8: anywhere).
func (e *RuleState) InArc(x1, y1, x2, y2, dir int) bool {
	return goldbox.InArc(len(e.Occ[0]), len(e.Occ), x1, y1, x2, y2, dir)
}

// BuildNear = ovr31:0909 (x, y, range, dir, size, map): every combatant on
// the map seen from the squares of a combatant of that size at (x,y),
// within range and the quarter dir (0xFF: all round), nearest first.
func (e *RuleState) BuildNear(x, y, rng, dir, size int) {
	var sx, sy [4]int
	var sok [4]bool
	for k := 0; k <= 3; k++ {
		dx, dy, ok := e.Footprint(size, k)
		sx[k], sy[k], sok[k] = x+dx, y+dy, ok
	}
	e.Near = e.Near[:0]
	for i := 1; i <= e.Ncomb; i++ {
		p := e.Cpos[i]
		if p.Size == 0 {
			continue
		}
		var tx, ty [4]int
		var tok [4]bool
		for m := 0; m <= 3; m++ {
			dx, dy, ok := e.Footprint(int(p.Size), m)
			tx[m], ty[m], tok[m] = int(p.X)+dx, int(p.Y)+dy, ok
		}
		found, best, bk, bm := false, 0xFF, 0, 0
		for m := 0; m <= 3; m++ {
			if !tok[m] {
				continue
			}
			for k := 0; k <= 3; k++ {
				if !sok[k] || !e.InArc(sx[k], sy[k], tx[m], ty[m], dir) {
					continue
				}
				ax, ay, r := tx[m], ty[m], rng
				if e.LineOfSight(&r, &ax, &ay, sx[k], sy[k]) {
					found = true
					if r < best {
						best, bk, bm = r, k, m
					}
				}
			}
		}
		if !found {
			continue
		}
		d := 0
		if dir < 8 {
			d = dir
		} else {
			for !e.InArc(sx[bk], sy[bk], tx[bm], ty[bm], d) {
				d++
			}
		}
		e.Near = append(e.Near, NearEntry{byte(i), byte(best), byte(d)})
	}
	e.nearSort()
}

// nearCombatants = ovr32:0C37 + ovr31:0909: everyone the combatant at (x,y)
// sees within range r, all round.
func (e *RuleState) nearCombatants(c *Character, x, y, r int) {
	e.BuildNear(x, y, r, 0xFF, e.CombatSize(c))
}
