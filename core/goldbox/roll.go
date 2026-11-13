package goldbox

// GameRNG is the game's own random source (its original-GameRNG implementation).
type GameRNG interface {
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
func RollCounted(rv *Roll, rng GameRNG, n, sides int) int {
	rv.DiceN = byte(n)
	return rng.Dice(n, sides)
}
