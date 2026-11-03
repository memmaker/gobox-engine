package input

// RepeatDelay and RepeatRate are the shared key repeat, in seconds: after
// RepeatDelay held, then every RepeatRate. One value for all games.
var RepeatDelay, RepeatRate = 0.3, 0.1

// Repeat reports whether a key or pad button held for held ticks (1 on the
// tick it went down, 0 not held) fires this tick, at tps ticks a second:
// on the first, then after RepeatDelay, then every RepeatRate (at least
// one tick). Pure: the caller counts the ticks.
func Repeat(held, tps int) bool {
	delay, rate := int(RepeatDelay*float64(tps)+.5), max(1, int(RepeatRate*float64(tps)+.5))
	return held == 1 || held > 1 && held >= delay && (held-delay)%rate == 0
}
