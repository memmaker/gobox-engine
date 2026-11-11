package core

// Additions of sweep Sfiles (fold E-R6): what por's file, init, sound and text
// code reads from the rules.

// The 16 EGA text colours a goldbox game draws with.
const (
	ColBlack = iota
	ColBlue
	ColGreen
	ColCyan
	ColRed
	ColMagenta
	ColBrown
	ColLightGrey
	ColDarkGrey
	ColLightBlue
	ColLightGreen
	ColLightCyan
	ColLightRed
	ColLightMagenta
	ColYellow
	ColWhite
)

// The sound numbers the game passes to PlaySound (the table at DS:2612).
const (
	SndStop       = 255 // stop the song
	SndMute       = 0   // stop and mute
	SndUnmute     = 1
	SndSpell      = 2 // a spell flies (the spells but 47 and 51)
	SndHit        = 3 // a weapon hits
	SndMagicHit   = 4 // a magical weapon hits
	SndFall       = 5 // a combatant falls
	SndMissile    = 6 // a missile flies
	SndBump       = 7 // walking into a wall
	SndSpell51    = 8
	SndMiss       = 9 // an attack hits nobody
	SndStep       = 10
	SndSpell47    = 11
	SndMissileEnd = 12 // a missile arrives
)

// DaspFx names the sound pack's event that replaces each sound number; the
// rest (a step) keep the original's sound.
var DaspFx = map[int]string{
	SndSpell:      "cast_spell",
	SndHit:        "hit",
	SndMagicHit:   "hit_good",
	SndFall:       "kill",
	SndMissile:    "shoot",
	SndBump:       string(CueBump),
	SndSpell51:    "res_elec", // Lightning Bolt
	SndMiss:       "miss",
	SndSpell47:    string(CueFireball),
	SndMissileEnd: "shoot_hit",
}

// InFight asks whether a fight is on (Yes).
type InFight struct{}

// PartyPlace asks where the party stands on the map: Low is x, High is y.
type PartyPlace struct{}

// AreaGeo asks the current area's GEO block number (Value).
type AreaGeo struct{}

func (InFight) check()    {}
func (PartyPlace) check() {}
func (AreaGeo) check()    {}
