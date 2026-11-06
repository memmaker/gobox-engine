package rules

import "time"

// Random = Turbo Pascal Random(n): *seed = *seed*0x08088405+1, then the high
// word mod n. The seed advances even for n = 0, which returns 0.
func Random(seed *uint32, n int) int {
	*seed = *seed*0x08088405 + 1
	if uint16(n) == 0 {
		return 0
	}
	return int(uint16(*seed>>16) % uint16(n))
}

// Randomize = Turbo Pascal Randomize: *seed = DOS time (CX = hour:min low
// word, DX = sec:hundredths high word), or *fixed when it is set (trace runs).
func Randomize(seed *uint32, fixed *uint32) {
	if fixed != nil {
		*seed = *fixed
		return
	}
	t := time.Now()
	*seed = uint32(t.Hour())<<8 | uint32(t.Minute()) | (uint32(t.Second())<<8|uint32(t.Nanosecond()/1e7))<<16
}

// Roller is a game's random source: Random(n) in 0..n-1.
type Roller interface {
	Random(n int) int
}

// Dice = the games' Dice(n, sides): sum of n rolls of 1..sides, kept in a
// byte like the original (wraps at 256).
func Dice(r Roller, n, sides int) int {
	t := 0
	for i := 0; i < n; i++ {
		t = (t + r.Random(sides) + 1) & 0xFF
	}
	return t
}
