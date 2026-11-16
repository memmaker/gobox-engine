package rulesxxvc

import (
	"math/bits"
	"slices"

	"gobox/rules"
)

// (arenaCols x arenaRows: the arena, combatArena^.cells); a map square lays a 6x5 block of it
// (PlotCell: arenaOX*6 + arenaOY*5 + 21 + px, arenaOY*5 + 10 + py).
const (
	blockCols = 6
	blockRows = 5
	blockX0   = 21 // PlotCell's column origin
	blockY0   = 10 // PlotCell's row origin
	// The placement grid of a team's ring is 11 columns by 6 rows (quadFree[team][ring][6][11]), four rings.
	gridCols  = 11
	gridRows  = 6
	gridRings = 4
	// FindPlacement's cell of a block: x + baseQX*6 + baseQY*5 + 22, y + baseQY*5 + 10 (TryPlaceAt).
	placeX0 = blockX0 + 1
	// A later pass's longest leg (FindPlacement: tries > 11).
	laterPassTries = 11
	// ring 1 is laid with the column ranges of row set 4 (InitOccupancy).
	ringOneRangeSet = 4
	teamParty       = 0 // C_Side.TEAM
	teamHostile     = 1 // C_Side.HOSTILE
)

// ringBlockDirections is gdconsts.pas ringBlockDirections [facing quarter][ring]; ring 0 is 8, no direction.
var ringBlockDirections = [4][4]int{{8, 4, 6, 2}, {8, 6, 4, 0}, {8, 0, 6, 2}, {8, 2, 0, 4}}

// facingDirections is gdconsts.pas facingDirections [facing quarter][ring].
var facingDirections = [4][4]int{{0, 0, 2, 6}, {2, 2, 0, 4}, {4, 4, 2, 6}, {6, 6, 4, 0}}

// quarterDirections is gdconsts.pas quarterDirections: the direction of each quarter.
var quarterDirections = [4]int{7, 2, 3, 6}

// ringStartColumns and ringStartRows are gdconsts.pas ringStartColumns/Rows [ring 0 or later][quarter].
var ringStartColumns = [2][4]int{{5, 4, 5, 6}, {3, 8, 7, 2}}
var ringStartRows = [2][4]int{{3, 2, 2, 3}, {0, 2, 5, 3}}

// columnRange is Types.TColumnRange: the first and last open column of a grid row; lo > hi opens none.
type columnRange struct{ lo, hi int }

// columnRanges is gdconsts.pas columnRanges [facing quarter, or 4 for ring 1][row].
var columnRanges = [5][gridRows]columnRange{
	{{1, 0}, {1, 0}, {1, 0}, {2, 9}, {3, 10}, {4, 10}},
	{{0, 2}, {0, 3}, {1, 4}, {2, 5}, {3, 6}, {4, 7}},
	{{0, 6}, {0, 7}, {1, 8}, {1, 0}, {1, 0}, {1, 0}},
	{{3, 6}, {4, 7}, {5, 8}, {6, 9}, {7, 10}, {8, 10}},
	{{0, 6}, {0, 7}, {1, 8}, {2, 9}, {3, 10}, {4, 10}},
}

// placer is combatst.pas's placement state (GDVars quadFree, placeX/Y, teamDir, teamRows) for one
// PlaceAllCreatures, and what FindPlacement needs of the fight.
type placer struct {
	r         *Rules
	f         rules.Fight
	walls     rules.MapSides // nil: open country, no walls between the blocks
	b         rules.Board    // the board with the units placed so far
	free      [2][gridRings][gridRows][gridCols]bool
	teamDir   [2]int
	place     [2]rules.BoardPos // placeX, placeY
	teamRows  [2]int
	solo      bool // isSoloCombat: not ported, no BR1 script sets it
	wildernes bool // interfaceState = WILDERNESS: no walls matter
	original  int  // originalFacing: the party's facing when placing began
	last      int  // lastFacing: the facing InitOccupancy laid last, noFacing before the first
}

const (
	noFacing         = -1   // lastFacing := 255
	iconSizeLowBits  = 0x07 // combatst.pas ICON_SIZE_LOW_BITS: the bits of iconDimensions PlaceAllCreatures takes
	approachQuarters = 4    // the quarters monsterApproach's bits 0..3 name
)

func (p *placer) wallBetween(x, y, dir int) int {
	return wallBetween(p.walls, p.f.Terms.PartyAt.Y, x, y, dir)
}

// The 3D map is 16 by 16 squares (geoMap wallsNE); the directions are C_Dir's.
const (
	arenaMapSide = 16
	dirNorth     = 0
	dirEast      = 2
	dirWest      = 6
)

// initOccupancy is PlaceAllCreatures' InitOccupancy for the party facing facing: unless that is the facing laid last,
// team 1's origin a monsterDistance off the party along it (1 when the distance is 0 and the facing is not the
// original), each team's facing quarter and the open cells of each ring's grid.
func (p *placer) initOccupancy(facing int) {
	distance := int(p.f.Terms.MonsterDist)
	if distance == 0 && facing != p.original {
		distance = 1
	}
	if facing == p.last {
		return
	}
	p.last = facing
	p.teamDir[teamParty] = facing / 2
	p.place[teamHostile] = dirStep[facing].Mul(distance).Add(p.place[teamParty])
	p.teamDir[teamHostile] = (facing + 4) % 8 / 2
	for team := teamParty; team <= teamHostile; team++ {
		for ring := range gridRings {
			set := p.teamDir[team]
			if ring == 1 {
				set = ringOneRangeSet
			}
			for qy := range gridRows {
				for qx := range gridCols {
					rg := columnRanges[set][qy]
					p.free[team][ring][qy][qx] = qx >= rg.lo && qx <= rg.hi
				}
			}
		}
	}
}

// tryPlaceAt is combatst.pas TryPlaceAt: id on cell (qx, qy) of team's ring grid laid at map square (baseQX, baseQY)
// from the party, when it is open and the unit fits (nobody there, on the arena, passable); then the cell is taken
// unless the unit is disabled in a solo combat. The unit's cell is set even when it does not fit.
func (p *placer) tryPlaceAt(id rules.CharID, team, qx, qy, baseQX, baseQY, ring int) bool {
	if qx < 0 || qy < 0 || qx >= gridCols || qy >= gridRows || !p.free[team][ring][qy][qx] {
		return false
	}
	at := rules.BoardPos{X: qx + baseQX*blockCols + baseQY*blockRows + placeX0, Y: qy + baseQY*blockRows + blockY0}
	if p.b.Straight { // the original's square, on the deskewed board
		at.X = rules.StraightX(at.X, at.Y)
	}
	i := slices.IndexFunc(p.b.Units, func(u rules.BoardUnit) bool { return u.ID == id })
	p.b.Units[i].At = at
	blocked, terrain, _ := p.r.probeDirection(p.b, id, centerDir)
	if blocked || terrain == 0 || terrainRow(terrain).cost >= impassable {
		return false
	}
	if p.r.chars[id].Enabled || !p.solo {
		p.free[team][ring][qy][qx] = false
	}
	return true
}

func (p *placer) offGrid(qx, qy int) bool {
	return qx < 0 || qy < 0 || qx >= gridCols || qy >= gridRows
}

// quadFullyOff is combatst.pas QuadFullyOff: off the grid on both axes.
func quadFullyOff(qx, qy int) bool { return (qx < 0 || qx >= gridCols) && (qy < 0 || qy >= gridRows) }

// findPlacement is combatst.pas FindPlacement: the team's rings walked out from its origin in zigzag legs for a cell
// tryPlaceAt accepts; false when the rings ran out.
func (p *placer) findPlacement(id rules.CharID, team int) bool {
	const (
		legStart  = 1 // the states of the walk: start a leg, step across, step back
		legAcross = 2
		legBack   = 3
	)
	first, state, legShift, ring := true, legStart, 0, 0
	base := p.place[team]
	var legX, legY, legLen, tries, qx, qy int
	for {
		quad := facingDirections[p.teamDir[team]][ring] / 2
		ringSet := 0
		if ring > 0 {
			ringSet = 1
		}
		switch state {
		case legStart:
			step := dirStep[quarterDirections[(quad+2)%4]]
			legX = ringStartColumns[ringSet][quad] + step.X*legShift
			legY = ringStartRows[ringSet][quad] + step.Y*legShift
			qx, qy, legLen, state, tries = legX, legY, 1, legAcross, 1
		case legAcross:
			step := dirStep[quarterDirections[(quad+1)%4]]
			qx, qy = legX+step.X*legLen, legY+step.Y*legLen
			state = legBack
			tries++
		case legBack:
			step := dirStep[quarterDirections[(quad+3)%4]]
			qx, qy = legX+step.X*legLen, legY+step.Y*legLen
			state = legAcross
			legLen++
			tries++
		}
		off := p.offGrid(qx, qy)
		if state > legStart {
			if off && !quadFullyOff(qx, qy) || first && tries >= p.teamRows[team] || !first && tries > laterPassTries {
				legShift++
				if team == teamParty && p.teamDir[team]%2 == 1 && ring == 0 && legShift == 1 {
					open := false
					for k := 1; k < len(ringBlockDirections[0]); k++ {
						dir := ringBlockDirections[p.teamDir[team]][k]
						if p.wildernes || p.wallBetween(p.place[team].X+p.f.Terms.PartyAt.X, p.place[team].Y+p.f.Terms.PartyAt.Y, dir) == 0 {
							open = true
						}
					}
					if open {
						legShift++
					}
				}
				state, first = legStart, false
			}
		}
		if off && quadFullyOff(qx, qy) {
			state = 0
			for ring < gridRings-1 && state != legStart {
				ring++
				dir := ringBlockDirections[p.teamDir[team]][ring]
				if p.wildernes || p.wallBetween(p.place[team].X+p.f.Terms.PartyAt.X, p.place[team].Y+p.f.Terms.PartyAt.Y, dir) != 1 {
					base = p.place[team].Add(dirStep[dir])
					legShift, state = 0, legStart
				}
			}
			if state != legStart {
				return false
			}
		}
		if !off && p.tryPlaceAt(id, team, qx, qy, base.X, base.Y, ring) {
			return true
		}
	}
}

// placeAllCreatures is combatst.pas PlaceAllCreatures (monsterApproach: the hostiles from the sides it names): every creature of the party list placed,
// each team facing the other. A downed creature in play that is placed leaves a corpse; one out of play (past the
// game's party count) that finds no place Left the list.
// A solo combat (FightTerms.Duel) places only the enabled.
func (r *Rules) placeAllCreatures(f rules.Fight, walls rules.MapSides) rules.Outcome {
	var o rules.Outcome
	p := &placer{r: r, f: f, walls: walls, b: r.board(rules.Board{}), wildernes: f.Terms.Area == rules.AreaWilderness}
	var counts [2]int // CountTeams: the enabled of each side
	for _, id := range r.fightParty(f) {
		if c := r.chars[id]; c != nil && c.Enabled {
			counts[sideTeam(c)]++
		}
	}
	p.teamRows[teamParty] = (counts[teamParty] + 1) / 2
	approach := f.Terms.Approach
	routes := 1 // no side named: one direction
	if approach&rules.ApproachSides != 0 {
		routes = bits.OnesCount8(uint8(approach & rules.ApproachSides))
	}
	p.teamRows[teamHostile] = (counts[teamHostile] + routes*2 - 1) / (routes * 2) // two grids a side
	p.original, p.last = int(f.Terms.PartyFacing)&7, noFacing
	p.initOccupancy(p.original)
	approachQuarter := 0
	for n, id := range r.fightParty(f) {
		c := r.chars[id]
		if c == nil {
			continue
		}
		team, outOfPlay := sideTeam(c), n+1 > int(r.memberCount)
		facing := p.original
		if approach&rules.ApproachSides != 0 { // each hostile from the next side named, the party from where it stands
			if team == teamHostile {
				for ok := false; !ok; ok = approach&(1<<approachQuarter) != 0 {
					approachQuarter = (approachQuarter + 1) % approachQuarters
				}
				facing = approachQuarter * 2
			}
			p.initOccupancy(facing)
		}
		face := facing
		if team == teamHostile {
			face = (facing + 4) % 8
		}
		p.b.Units = append(p.b.Units, rules.BoardUnit{ID: id, Size: int32(c.IconDimensions & iconSizeLowBits), Facing: p.facingDir(face), Side: c.Side})
		if !(f.Terms.Duel && !c.Enabled) && p.findPlacement(id, team) { // a solo combat places only who is enabled (combatst.pas:948)
			if !c.Enabled && !outOfPlay {
				u := p.b.Units[len(p.b.Units)-1]
				r.layCorpses(&p.b, []rules.CharID{id}, func(rules.CharID) (rules.BoardPos, bool) { return u.At, true }, &o)
			}
			continue
		}
		last := len(p.b.Units) - 1
		p.b.Units = slices.Delete(p.b.Units, last, last+1) // size 0: off the arena
		if outOfPlay {
			o.Effects = append(o.Effects, rules.Left{Unit: id})
		}
	}
	r.ground.Units = p.b.Units
	return o
}

// facingDir is the unit's facing for the party's facing face: the original's table leans the north-south moves into its
// perspective; deskewed, north is north (rules1e's facingDir).
func (p *placer) facingDir(face int) rules.Facing {
	if p.b.Straight {
		return rules.Facing(face &^ 1) // the quarter's own direction, 0 2 4 6
	}
	return rules.Facing(quarterDirections[face/2])
}

// sideTeam is C_Side: the party's side, steered or not, is 0 (TEAM), the hostile 1.
func sideTeam(c *Character) int {
	if c.team() {
		return teamParty
	}
	return teamHostile
}
