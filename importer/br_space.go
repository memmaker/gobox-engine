package importer

import (
	"gobox/adventure"
	"gobox/rules"
)

// The system map's sizes and constants (buckroger-re gdconsts.pas orbitPositions, orbitFirstPositions, orbitLengths,
// outerBodyPhases, orbitStartPhases; orbitmap.pas; explore.pas MoveRocket; gameinit.pas START_ORBIT).
const (
	spacePositions  = 58 // orbitPositions[0..57], 3 bytes each: x, y, sprite
	spaceOrbits     = 5  // orbitFirstPositions, orbitLengths and orbitStartPhases hold one per orbit; the last is the belt's too
	spaceBelt       = 8  // outerBodyPhases
	spaceHeadings   = 8  // mapStepX and mapStepY
	spaceSide       = 22 // the rocket's column and row stop at 21 (explore.pas MoveRocket)
	spaceSunX       = 11 // C_Orbit.CENTER_X, CENTER_Y
	spaceSunY       = 10
	spaceSunTile    = 296     // orbitmap.pas ORBIT_TILES
	spaceBeltTile   = 268     // OUTER_BODY_TILES
	spaceCycle      = 224     // explore.pas PHASE_CYCLE
	spaceStartOrbit = 2       // gameinit.pas START_ORBIT
	spaceTilesFile  = "8X8D0" // orbitmap.pas LoadOrbitPictures: dataSet := 0, Load8x8Set(1, ORBIT_TILE_SLOT)
	spaceTilesBlock = 1
	spaceShipFrame  = 87 // C_Orbit.FIRST_ROCKET_ICON
	spaceTicks      = 3  // MoveRocket: 3 ticks and RestAndHeal(3) a move
	spaceTickCycle  = 32 // the tick that reaches 32 moves the bodies on a phase
	spaceDays       = 3
)

// SpaceTables is the system map from the data segment ds: the position rows at e.Space, then orbitFirstPositions (5), orbitLengths
// (5), outerBodyPhases (8) and orbitStartPhases (5) as the data segment lays them out, and mapStepX, mapStepY (8 each) at
// e.MapSteps. A game with no e.Space has none.
func SpaceTables(ds []byte, e BREXE) *adventure.SpaceMap {
	if e.Space == 0 {
		return nil
	}
	s := &adventure.SpaceMap{
		Size: rules.Size{W: spaceSide, H: spaceSide}, Sun: rules.MapPos{X: spaceSunX, Y: spaceSunY}, SunTile: spaceSunTile, BeltTile: spaceBeltTile,
		Tiles: adventure.SymbolLoad{File: spaceTilesFile, Block: spaceTilesBlock}, ShipFrame: spaceShipFrame,
		Cycle: spaceCycle, Move: adventure.SpaceMove{Ticks: spaceTicks, TickCycle: spaceTickCycle, Days: spaceDays},
	}
	for i := range spacePositions {
		r := ds[e.Space+3*i:]
		s.Squares = append(s.Squares, adventure.SpaceSquare{X: int(r[0]), Y: int(r[1]), Sprite: int(r[2])})
	}
	t := ds[e.Space+3*spacePositions:]
	first, length, belt, phase := t, t[spaceOrbits:], t[2*spaceOrbits:], t[2*spaceOrbits+spaceBelt:]
	for i := range spaceOrbits {
		s.Orbits = append(s.Orbits, adventure.SpaceOrbit{First: int(first[i]), Length: int(length[i]), Phase: int(phase[i])})
	}
	for i := range spaceBelt {
		s.Offsets = append(s.Offsets, int(belt[i]))
	}
	for h := range spaceHeadings {
		s.Steps = append(s.Steps, rules.MapPos{X: int(int8(ds[e.MapSteps+h])), Y: int(int8(ds[e.MapSteps+spaceHeadings+h]))})
	}
	o := s.Orbits[spaceStartOrbit]
	s.Start = o.First + o.Phase
	return s
}
