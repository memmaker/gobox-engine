package rules

// RNG is the one pre-seeded random stream of a game: script RANDOM, every rule
// roll. The engine owns it and saves it, and passes it to each Rules method
// that rolls; a ruleset holds no stream of its own. Plain data, so no callback.
//
// Presentation (voices, por dasp.go:166) draws from a second RNG, also the
// engine's, so that frames never shift a game's rolls.
//
// The sequence is Turbo Pascal's (gobox/rules.Random), which POR, COTAB, SotSB
// and POD traces need: state*0x08088405+1, the high word mod n.
type RNG struct{ State uint32 }

// N is 0..n-1. The state advances even for n = 0, which gives 0.
func (r *RNG) N(n int) int {
	r.State = r.State*0x08088405 + 1
	if uint16(n) == 0 {
		return 0
	}
	return int(uint16(r.State>>16) % uint16(n))
}
