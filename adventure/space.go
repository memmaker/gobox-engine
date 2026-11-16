package adventure

import "gobox/rules"

// SpaceMap is a game's system map, the map the rocket flies (BR1 explore.pas MoveRocket, orbitmap.pas): the bodies on their orbits,
// the sun in the middle, the rules of a move. The bodies' squares are derived from the orbit step (Body), never stored.
type SpaceMap struct {
	Size      rules.Size     // the squares the ship ranges over (22 by 22)
	Sun       rules.MapPos   // the centre body's square: no ship stands on it (C_Orbit.CENTER_X, CENTER_Y)
	SunTile   int            // the sun's picture number (OrbitMap ORBIT_TILES)
	Tiles     SymbolLoad     // the 8x8 symbol set the bodies' and the ship's pictures are frames of (OrbitMap LoadOrbitPictures: data set 0, block 1)
	ShipFrame int            // the frame of the ship pointing at heading 0; heading h is ShipFrame+h (C_Orbit.FIRST_ROCKET_ICON)
	Squares   []SpaceSquare  // every body square, orbit by orbit (GDConsts.orbitPositions)
	Orbits    []SpaceOrbit   // the planets' orbits (orbitFirstPositions, orbitLengths, orbitStartPhases)
	Belt      SpaceOrbit     // the last orbit, which the belt's bodies share
	Offsets   []int          // each belt body's phase on the belt (GDConsts.outerBodyPhases)
	BeltTile  int            // added to a belt body's sprite (OrbitMap OUTER_BODY_TILES); a planet's sprite takes SunTile
	Steps     []rules.MapPos // a heading's step, headings 0.. counterclockwise from north (GDConsts.mapStepX, mapStepY)
	Start     int            // the square of Squares the ship starts on (GameInit START_ORBIT's body at phase 0)
	Cycle     int            // the orbit step counts down through 0..Cycle-1 and wraps (Explore PHASE_CYCLE)
	Move      SpaceMove
}

// SpaceSquare is one body square: where it is and its sprite.
type SpaceSquare struct{ X, Y, Sprite int }

// SpaceOrbit is the squares First.. of Squares, Length of them, a body starting Phase in.
type SpaceOrbit struct{ First, Length, Phase int }

// SpaceMove is what a ship's move costs on the clock (Explore MoveRocket: 3 ticks and 3 days; at 32 the bodies move on).
type SpaceMove struct{ Ticks, TickCycle, Days int }

// SpaceBody is a body as the map shows it: its square and its picture.
type SpaceBody struct {
	At   rules.MapPos
	Tile int
}

// Bodies is the number of bodies: a planet per orbit, then the belt's.
func (s SpaceMap) Bodies() int { return len(s.Orbits) + len(s.Offsets) }

// Body is body i (the planets, then the belt's) with the orbits at step (OrbitMap PaintOrbitMap).
func (s SpaceMap) Body(i, step int) SpaceBody {
	o, off, tile := s.Orbits[len(s.Orbits)-1], 0, s.BeltTile
	if i < len(s.Orbits) {
		o, tile = s.Orbits[i], s.SunTile
	} else {
		off = s.Offsets[i-len(s.Orbits)]
	}
	q := s.Squares[o.First+(o.Phase+off+step)%o.Length]
	return SpaceBody{At: rules.MapPos{X: q.X, Y: q.Y}, Tile: q.Sprite + tile}
}
