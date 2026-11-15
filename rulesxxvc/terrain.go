package rulesxxvc

import (
	"slices"

	"gobox/rules"
)

// lineClass is TCombatTerrain.lineClass: how a terrain treats a line (pathing.pas).
type lineClass uint8

const (
	lineOpen    lineClass = 0
	lineStops   lineClass = 1 // LINE_STOPS: stops a line while distMode is 0
	lineBlocked lineClass = 2 // LINE_BLOCKED: sawBlockedTerrain past the grace cells
)

// impassable is C_Terr.IMPASSABLE, a cost no move can pay.
const impassable = 255

// The arena markers c_terr.pas names.
const (
	chaffTerrain       rules.TerrainID = 1
	gasTerrain         rules.TerrainID = 2
	aerosolMistTerrain rules.TerrainID = 4
	corpseTerrain      rules.TerrainID = 5
	openTerrain        rules.TerrainID = 7 // an empty Board.Terrain: combatTerrain[7], cost 1, no line class, no marker
)

type combatTerrainRow struct {
	cost  int
	line  lineClass
	blast rules.TerrainID // blastResult: what a blast leaves, 0 unchanged; 250..253 a rubble roll (scanArea)
}

// combatTerrain is gdconsts.pas combatTerrain[1..LAST_COMBAT_TERRAIN] (cost, lineClass, blastResult).
var combatTerrain = [154]combatTerrainRow{{1, 0, 0}, {1, 0, 0}, {1, 0, 0}, {1, 0, 0}, {1, 0, 0}, {1, 0, 0}, {1, 0, 0}, {255, 1, 0}, {1, 0, 0}, {255, 1, 0}, {255, 1, 0}, {255, 1, 0}, {1, 0, 0}, {255, 1, 0}, {255, 1, 0}, {255, 1, 0}, {255, 1, 0}, {255, 1, 0}, {1, 0, 0}, {255, 1, 0}, {255, 1, 0}, {255, 1, 0}, {255, 1, 0}, {1, 0, 0}, {255, 1, 0}, {1, 0, 0}, {255, 1, 0}, {1, 0, 0}, {255, 1, 0}, {255, 1, 0}, {255, 1, 0}, {255, 1, 0}, {1, 0, 0}, {1, 0, 0}, {255, 1, 0}, {1, 0, 0}, {255, 1, 0}, {1, 0, 0}, {1, 0, 0}, {1, 0, 0}, {255, 1, 0}, {255, 1, 0}, {255, 1, 0}, {255, 1, 0}, {1, 0, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {1, 0, 0}, {1, 0, 0}, {1, 0, 0}, {1, 0, 0}, {2, 0, 0}, {1, 0, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {255, 1, 0}, {255, 1, 0}, {255, 1, 0}, {255, 1, 0}, {255, 1, 0}, {255, 1, 0}, {255, 1, 0}, {255, 1, 0}, {255, 1, 0}, {255, 1, 0}, {255, 2, 0}, {255, 2, 0}, {255, 2, 0}, {4, 2, 0}, {4, 2, 0}, {4, 2, 0}, {4, 2, 0}, {4, 2, 0}, {4, 2, 0}, {4, 2, 0}, {1, 0, 139}, {1, 0, 139}, {2, 0, 0}, {255, 0, 0}, {255, 0, 0}, {3, 2, 131}, {3, 2, 132}, {3, 2, 133}, {3, 2, 134}, {1, 2, 135}, {1, 2, 136}, {255, 2, 137}, {255, 2, 138}, {3, 2, 140}, {3, 2, 141}, {3, 2, 142}, {3, 2, 143}, {3, 2, 144}, {3, 2, 145}, {1, 0, 146}, {1, 0, 146}, {255, 0, 0}, {255, 0, 0}, {2, 0, 0}, {2, 0, 0}, {2, 0, 0}, {3, 2, 147}, {3, 2, 148}, {3, 0, 153}, {255, 2, 250}, {255, 2, 250}, {255, 2, 250}, {255, 2, 250}, {255, 2, 250}, {255, 2, 251}, {255, 2, 251}, {255, 2, 252}, {255, 2, 252}, {1, 2, 151}, {1, 2, 151}, {1, 2, 151}, {1, 2, 151}, {2, 0, 0}, {2, 0, 0}, {2, 0, 0}, {2, 0, 0}, {1, 2, 0}, {1, 2, 0}, {255, 2, 0}, {255, 2, 0}, {1, 0, 0}, {2, 0, 0}, {2, 0, 0}, {2, 2, 0}, {2, 0, 0}, {2, 0, 0}, {2, 0, 0}, {1, 0, 0}, {2, 2, 0}, {2, 0, 0}, {255, 0, 251}, {255, 2, 252}, {1, 2, 0}, {255, 0, 250}, {2, 0, 0}, {2, 0, 0}}

// terrainAt is the arena cell at p (CbtMap.GetArenaContents' terrain), 0 off the board.
func terrainAt(b rules.Board, p rules.BoardPos) rules.TerrainID {
	if p.X < 0 || p.Y < 0 || p.X >= b.Size.X || p.Y >= b.Size.Y {
		return 0
	}
	if len(b.Terrain) == 0 {
		return openTerrain
	}
	return b.Terrain[p.Y*b.Size.X+p.X]
}

// terrainRow is combatTerrain[t]; a value past LAST_COMBAT_TERRAIN reads as 21 (CbtMap).
func terrainRow(t rules.TerrainID) combatTerrainRow {
	if t == 0 {
		return combatTerrainRow{cost: impassable}
	}
	if int(t) > len(combatTerrain) {
		t = 21
	}
	return combatTerrain[t-1]
}

// lineSeen is WalkLine's saw flags.
type lineSeen struct{ chaff, mist, blocked bool }

// graceCells is pathing.pas GRACE_CELLS: cells at a line's start where
// LINE_BLOCKED terrain is not counted.
const graceCells = 2

// walkLine is pathing.pas WalkLine over goldbox.Line (PathFind, NextPathStep):
// the line from a toward b, dist its range in straight steps. With distMode
// 0 (stopsAtWalls) it stops early on a LINE_STOPS cell unless a is b; on
// chaff when stopsAtChaff; past dist*2+1. It answers the cell reached, the
// distance walked (2 a straight step, 3 a diagonal), whether b was reached
// and what it saw.
func walkLine(b rules.Board, from, to rules.BoardPos, dist int, stopsAtWalls, stopsAtChaff bool) (rules.BoardPos, int, bool, lineSeen) {
	var seen lineSeen
	grace := graceCells
	maxDist := dist*2 + 1
	l := rules.NewLine(from.X, from.Y, to.X, to.Y)
	for {
		c := rules.BoardPos{X: l.X, Y: l.Y}
		cell := terrainAt(b, c)
		done := false
		switch {
		case cell == chaffTerrain:
			seen.chaff = true
			done = stopsAtChaff
		case cell == aerosolMistTerrain:
			seen.mist = true
		case grace > 0:
			grace--
		case terrainRow(cell).line == lineBlocked:
			seen.blocked = true
		}
		if stopsAtWalls && terrainRow(cell).line == lineStops && from != to || done || l.Steps > maxDist {
			return c, l.Steps, false, seen
		}
		if !l.Step() {
			return c, l.Steps, true, seen
		}
	}
}

// iconCells is gdconsts.pas iconCells: the cells an icon of size 1..4 covers
// from its corner.
var iconCells = [4][]rules.BoardPos{
	{{}},
	{{}, {X: 0, Y: 1}},
	{{}, {X: 1, Y: 0}},
	{{}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}},
}

func cellsOf(u rules.BoardUnit) []rules.BoardPos {
	size := max(1, min(int(u.Size), len(iconCells)))
	cs := make([]rules.BoardPos, 0, 4)
	for _, d := range iconCells[size-1] {
		cs = append(cs, u.At.Add(d))
	}
	return cs
}

// occupant is cellCombatant: the unit covering p. A downed fighter is off
// the arena arrays (CreatureDowned: size 0, BuildArenaArrays), so it covers
// nothing; its corpse is terrain.
func (r *Rules) occupant(b rules.Board, p rules.BoardPos) (rules.CharID, bool) {
	for _, u := range b.Units {
		if c := r.chars[u.ID]; c != nil && !c.Enabled {
			continue
		}
		if slices.Contains(cellsOf(u), p) {
			return u.ID, true
		}
	}
	return rules.NoChar, false
}

// probeDirection is cbtmap.pas ProbeDirection: over the cells id's icon
// covers moved one step dir, another combatant there, the costliest terrain
// (0 when any cell is off the board) and whether one is gas.
func (r *Rules) probeDirection(b rules.Board, id rules.CharID, dir int) (blocked bool, terrain rules.TerrainID, harmful bool) {
	terrain = 23
	best := 1
	i := slices.IndexFunc(b.Units, func(u rules.BoardUnit) bool { return u.ID == id })
	if i < 0 {
		return false, 0, false
	}
	for _, c := range cellsOf(b.Units[i]) {
		n := c
		if dir != centerDir { // dirStepX/Y[8] = 0
			n = c.Add(dirStep[dir])
		}
		if o, ok := r.occupant(b, n); ok && o != id {
			blocked = true
		}
		t := terrainAt(b, n)
		if t == gasTerrain {
			harmful = true
		}
		if t == 0 || terrain == 0 {
			terrain = 0
		} else if row := terrainRow(t); row.cost >= best {
			best, terrain = row.cost, t
		}
	}
	return blocked, terrain, harmful
}

// stepCost is the movement one step onto terrain t costs (AiCanStep,
// CombatMoveCreature): its cost, times 3 on a diagonal, times 2 straight.
func stepCost(dir int, t rules.TerrainID) int {
	if dir%2 == 1 {
		return terrainRow(t).cost * 3
	}
	return terrainRow(t).cost * 2
}
