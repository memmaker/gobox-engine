package input

// Repeat reports whether a key or pad button held for held ticks (1 on the
// tick it went down, 0 not held) fires this tick: on the first, then after
// delay ticks, then every rate ticks. The host sets delay and rate (por:
// tps/3 and tps*3/20). Pure: the caller counts the ticks.
func Repeat(held, delay, rate int) bool {
	return held == 1 || held > 1 && rate > 0 && held >= delay && (held-delay)%rate == 0
}
