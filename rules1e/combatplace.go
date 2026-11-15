package rules1e

import (
	"gobox/rules"
)

// combat placement directions, by party direction and side (DS:02D2).
var PlaceDirs = [4][4]Dir{
	{0x08, 0x04, 0x06, 0x02},
	{0x08, 0x06, 0x04, 0x00},
	{0x08, 0x00, 0x06, 0x02},
	{0x08, 0x02, 0x00, 0x04},
}

// combat placement team directions, by party direction and team (DS:02E2).
var PlaceTeamDirs = [4][4]Dir{
	{0x00, 0x00, 0x02, 0x06},
	{0x02, 0x02, 0x00, 0x04},
	{0x04, 0x04, 0x02, 0x06},
	{0x06, 0x06, 0x04, 0x00},
}

// the eight-way direction of each of the four facings (DS:02F2).
var FacingDir = [4]Dir{0x07, 0x02, 0x03, 0x06}

// combatBlocks = ovr11:12E1: a fresh combat block for every combatant;
// those past the party count are not party members.
func (r *Rules) combatBlocks() {
	s := r.S
	for i, c := range r.S.Party {
		s.ApplyItems(c) // 00BA:0BB8
		c.Combat = &CombatInfo{}
		if i+1 > int(r.S.Game.PartyCount) {
			c.Combat.Extra = 1
		}
		c.Combat.Facing = r.facingDir(s.Facing / 2)
		if c.Side == SideMonsters {
			c.Combat.Facing = (c.Combat.Facing + 4) % 8
		}
		m := c.Control & 0x7F
		if c.Side == SideParty && c.Combat.Delay == 1 && (m == 0 || m > 0x66) {
			c.Control = byte(r.S.Game.Morale + 0x80)
		}
	}
}

// cellSide = ovr11:012B (x, y, dir): 1 wall, 0 open, 3 door. Off the
// 16x16 map only the party's row is open east and west.
func CellSide(m rules.MapSides, posY, x, y, dir int) byte {
	if x >= 0 && x <= 0xF && y >= 0 && y <= 0xF {
		if m.DoorAt(x, y, dir) == 0 {
			return 1
		}
		if m.WallAt(x, y, dir) == 0 {
			return 0
		}
		return 3
	}
	if y == posY && (dir == 2 || dir == 6) {
		return 0
	}
	return 1
}

// cellEdge = ovr11:01AD (x, y, dir): the side seen from both cells.
func CellEdge(m rules.MapSides, posY, x, y, dir int) byte {
	dx, dy := DirDX[dir], DirDY[dir]
	return CellSide(m, posY, x, y, dir) | CellSide(m, posY, x+dx, y+dy, (dir+4)%8)
}

// facingDir is the eight-way direction of facing f (0..3): the original's
// table leans north-south moves into its perspective; deskewed, north is north.
func (r *Rules) facingDir(f int) byte {
	if r.straight {
		return byte(f * 2)
	}
	return byte(FacingDir[f])
}

// placeCombatants = ovr11:19F5.
func (r *Rules) placeCombatants(m rules.MapSides) []rules.Effect {
	var left []rules.Effect
	r.S.CountSides()
	r.S.Cpos = [MaxCombatants + 1]CombatPos{}
	r.S.CombatantSlots = make([]*Character, MaxCombatants)
	r.S.Ncomb = 0
	r.S.CombatRedrawMap()
	p := &r.S.Place
	p.Ox[0], p.Oy[0] = 0, 0
	p.Dir[0] = Dir(byte(r.S.Facing / 2))
	dist := int(r.S.Game.MonsterDist)
	// extra: deskewed, a side's formation is its own cell's floor, so the foes take the cell ahead
	// (they shared the party's 5x4 floor, among it); never through a wall or a door, as the
	// original's distance (FreeSteps) never is: there they share the floor still
	if r.straight && dist == 0 && (m == nil || CellEdge(m, r.S.PosY, r.S.PosX, r.S.PosY, r.S.Facing) == 0) {
		dist = 1
	}
	p.Ox[1] = int8(dist*DirDX[r.S.Facing] + int(p.Ox[0]))
	p.Oy[1] = int8(dist*DirDY[r.S.Facing] + int(p.Oy[0]))
	p.Dir[1] = Dir(byte((r.S.Facing + 4) % 8 / 2))
	for s := 0; s <= 1; s++ {
		p.Rows[s] = Count(byte((int(r.S.Rv.Side[s]) + 1) >> 1))
	}
	for s := 0; s <= 1; s++ {
		for team := 0; team <= 3; team++ {
			d := int(p.Dir[s])
			if team == 1 {
				d = 4
			}
			for y := 0; y <= 5; y++ {
				lr := PlaceRows[d][y]
				if r.straight { // the cell's floor, whatever the formation
					lr = [2]int8{1, 0}
					if y >= 1 {
						lr = [2]int8{0, 4}
					}
				}
				for x := 0; x <= 0xA; x++ {
					v := byte(0)
					if int(lr[0]) <= x && int(lr[1]) >= x {
						v = 1
					}
					p.Slots[s][team][y][x] = Slot(v)
				}
			}
		}
	}
	idx := 1
	r.S.Ncomb = 1
	r.S.Cpos[0].Size = 1 // original quirk: the absolute DS:5F2A is record 0's size, counted up per combatant
	for i := 0; i < len(r.S.Party); i++ {
		c := r.S.Party[i]
		r.S.CombatantSlots[idx-1] = c
		p.Side = c.Side
		r.S.Cpos[idx].Id = byte(idx)
		r.S.Cpos[idx].Size = c.Size & 7
		if r.placeOne(m, idx) {
			if c.Okay == 0 {
				r.S.Cpos[idx].Size = 0
				if !r.duel && c.Combat.Extra == 0 {
					x, y := r.S.Cpos[idx].X, r.S.Cpos[idx].Y
					r.S.Nbodies++
					r.S.Bodies[r.S.Nbodies] = Body{C: c, X: x, Y: y, Ground: byte(r.S.Cmap.Ground[y][x])}
					r.S.Cmap.Ground[y][x] = r.S.grounds().Body
				}
			}
			r.S.CombatRedrawMap()
			idx++
			r.S.Cpos[0].Size++
			r.S.Ncomb++
			continue
		}
		r.S.Cpos[idx].Size = 0
		if c.Combat.Extra == 1 { // no room: the extra leaves
			r.S.CombatantSlots[idx-1] = nil
			left = append(left, rules.Left{Unit: r.ID(c)})
			r.Dismiss(r.ID(c)) // por RemoveSelected(keepCount): the party count stays
			i--
			continue
		}
		idx++
		r.S.Cpos[0].Size++
		r.S.Ncomb++
	}
	return left
}

// placeAt = ovr11:1434 (idx, x, y, ox, oy, team): put combatant idx on
// formation slot (x,y) of the team block at (ox,oy) if the ground lets it.
func (r *Rules) placeAt(idx, x, y, ox, oy, team int) bool {
	p := &r.S.Place
	if x < 0 || y < 0 || x > 0xA || y > 5 || p.Slots[p.Side][team][y][x] == 0 {
		return false
	}
	r.S.Cpos[idx].X = byte(x + ox*6 + oy*5 + 0x16)
	if r.straight {
		r.S.Cpos[idx].X -= byte(oy * 5)
	}
	r.S.Cpos[idx].Y = byte(oy*5 + y + 0xA)
	who, g := r.S.GroundAt(r.S.CombatantSlots[idx-1], 8)
	if who == 0 && g > 0 && r.S.GroundInfo(Ground(g))[0] < 0xFF {
		p.Slots[p.Side][team][y][x] = 0
		return true
	}
	return false
}

// placeOne = ovr11:1577 (idx): walk the formation of the combatant's side,
// row by row outwards from the middle, falling back on the other teams
// (blocks beside the first) until a free slot takes it.
func (r *Rules) placeOne(m rules.MapSides, idx int) bool {
	p := &r.S.Place
	s := p.Side
	first, fail, state := true, false, 1
	var row, team, k, n int
	var bx, by, x, y int
	var dx, dy int
	ox, oy := int(p.Ox[s]), int(p.Oy[s])
	for {
		dir := int(PlaceTeamDirs[p.Dir[s]][team] >> 1)
		turn := func(q int) {
			d := int(r.facingDir((dir + q) % 4))
			dx, dy = DirDX[d], DirDY[d]
		}
		switch state {
		case 1:
			turn(2)
			t := 0
			if team > 0 {
				t = 1
			}
			bx = int(PlaceStartX[t][dir]) + dx*row
			by = int(PlaceStartY[t][dir]) + dy*row
			if r.straight {
				bx = int(PlaceStartXFlat[t][dir]) + dx*row
				by = int(PlaceStartYFlat[t][dir]) + dy*row
			}
			x, y = bx, by
			k, state, n = 1, 2, 1
		case 2:
			turn(1)
			x, y = bx+dx*k, by+dy*k
			state = 3
			n++
		case 3:
			turn(3)
			x, y = bx+dx*k, by+dy*k
			state = 2
			k++
			n++
		}
		out := x < 0 || y < 0 || x > 0xA || y > 5
		// A row that starts off the block ends the team too (extra): a row whose first
		// square is off is skipped, and the rows only move further off along the same
		// line, so no slot is left. The original only stopped at a corner, off both
		// ways, which its slanted rows always reach; Deskew's rows run straight along
		// one axis and never do: with every slot taken they ran off for ever (a hang).
		edge := out && (rules.FormationEdge(x, y, formationW, formationH) || x == bx && y == by)
		if state > 1 {
			full := out && !edge
			if !full {
				if first {
					full = n >= int(p.Rows[s])
				} else {
					full = n > 0xB
				}
			}
			if full {
				row++
				if s == 0 && p.Dir[0]&1 != 0 && team == 0 && row == 1 {
					cx, cy := r.S.PosX+int(p.Ox[0]), r.S.PosY+int(p.Oy[0])
					open := false
					for a := 1; a <= 3; a++ {
						d := int(PlaceDirs[p.Dir[s]][a])
						if m == nil || CellEdge(m, r.S.PosY, cx, cy, d) != 1 {
							open = true
						}
					}
					if open {
						row++
					}
				}
				state, first = 1, false
			}
		}
		ok := false
		if edge {
			state = 0
			for team < 3 && state != 1 {
				team++
				cx, cy := r.S.PosX+int(p.Ox[s]), r.S.PosY+int(p.Oy[s])
				d := int(PlaceDirs[p.Dir[s]][team])
				if m != nil && CellEdge(m, r.S.PosY, cx, cy, d) == 1 {
					continue
				}
				ox, oy = DirDX[d]+int(p.Ox[s]), DirDY[d]+int(p.Oy[s])
				row, state = 0, 1
			}
			if state != 1 {
				fail = true
			}
		}
		if !out {
			ok = r.placeAt(idx, x, y, ox, oy, team)
		}
		if ok || fail {
			return !fail
		}
	}
}

// The formation is formationW x formationH squares.
const formationW, formationH = 0xB, 6
