package rules

// The cues of the sound pack's events a rule or the engine makes.
const (
	CueStep       SoundCue = ""             // a step: the pack has none, the original's sound plays
	CueBump       SoundCue = "hitwall"      // walking into a wall (PlaySound 7)
	CueFireball   SoundCue = "breathe_fire" // PlaySound 0xB (DS:262A)
	CueSpell      SoundCue = "cast_spell"   // a spell flies
	CueHit        SoundCue = "hit"          // a weapon hits
	CueMagicHit   SoundCue = "hit_good"     // a magical weapon hits
	CueFall       SoundCue = "kill"         // a combatant falls
	CueMissile    SoundCue = "shoot"        // a missile flies
	CueLightning  SoundCue = "res_elec"     // Lightning Bolt
	CueMiss       SoundCue = "miss"         // an attack hits nobody
	CueMissileEnd SoundCue = "shoot_hit"    // a missile arrives
)
