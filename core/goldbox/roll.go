package goldbox

// RNG is the game's own random source (its original-RNG implementation).
type RNG interface {
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

// RollDice = Dice that also remembers n (the games' RollDice).
func RollDice(rv *Roll, rng RNG, n, sides int) int {
	rv.DiceN = byte(n)
	return rng.Dice(n, sides)
}
