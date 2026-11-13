package rules1e

// Party menu, ported from ovr17:0161 and the list display 00BA:0767.

// The selection (Felix 2026-11-11): the Engine's sel is only the presenter
// remembering whom to pass; the rules' actor (RuleState.who, DS:5D92) is set
// from it for each call into the rules (as), and read back after the calls
// that move it (follow): joining, leaving, a fight's turns, a script's pick.
// TestActorSeam: every Engine call into rules reading the actor goes through as.

// setActor makes member n the actor.
func (e *RuleState) setActor(n int) {
	if n >= 0 && n < len(e.Party) {
		e.Who = n
	}
}
