package core

// The world: the rules' dice and the clock.

// RollDice asks N dice of Sides from the rules' own generator (Answer.Value).
type RollDice struct{ N, Sides int32 }

// RollRandom asks Random(N) from the rules' generator (Answer.Value), the script's RANDOM.
type RollRandom struct{ N int32 }

// IsDay asks whether the clock is by day (Yes).
type IsDay struct{}

func (RollDice) check() {}

func (RollRandom) check() {}

func (IsDay) check() {}

// MapTerrain asks the area map's terrain number at map square (X, Y) (Value).
type MapTerrain struct{ X, Y int }

func (MapTerrain) check() {}

// AreaTileIs asks the area map's tile at square (X, Y) (Answer.Value): the
// terrain, or a landmark while it shows (por AreaTile).
type AreaTileIs struct{ X, Y int32 }

func (AreaTileIs) check() {}

// AdvanceClock adds Count units of the area clock (Unit: the clock's word,
// 1 minutes ..) and runs the timed effects down (por AdvanceClock, ovr21:03D5).
type AdvanceClock struct{ Count, Unit int32 }

func (AdvanceClock) action() {}

// DevPlace sets a dev scene's fight up before its first turn: Melee puts a
// foe beside each member, else each member goes to the map's edge.
type DevPlace struct{ Melee bool }

func (DevPlace) action() {}
