package rules

// DiceSource is what rolls dice: the game's own Dice(n, sides).
type DiceSource interface {
	Dice(n, sides int) int
}

// Roll is the rule globals the games keep in rv: the last results, which the
// affect handlers read.
type Roll struct {
	Save     byte
	SaveType byte
	SaveMade bool
	DiceN    byte
	HitRoll  int8
}

// RollCounted = Dice that also remembers n (the games' RollCounted).
func RollCounted(rv *Roll, rng DiceSource, n, sides int) int {
	rv.DiceN = byte(n)
	return rng.Dice(n, sides)
}
