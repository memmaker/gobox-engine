package core

// The cues the engine makes outside the rules: the sound pack's events
// (rules1e DaspFx names the same).
const (
	CueStep     SoundCue = ""             // a step: the pack has none, the original's sound plays
	CueBump     SoundCue = "hitwall"      // walking into a wall (PlaySound 7)
	CueFireball SoundCue = "breathe_fire" // PlaySound 0xB (DS:262A)
)
