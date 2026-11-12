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
