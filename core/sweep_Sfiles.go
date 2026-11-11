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

// AreaGeo asks the current area's GEO block number (Value).
type AreaGeo struct{}

func (InFight) check() {}
func (AreaGeo) check() {}

type (
	// InitRules readies the rules' tables for a new session (por Init:
	// SpellTableInit, SetupAffectHandlers).
	InitRules struct{}
	// NewAdventure puts the rules' adventure state back to a new game's: the
	// game and area records, an empty purse, loot and pool.
	NewAdventure struct{}
	// EndSession closes what a session had open: camp, training, learning.
	EndSession struct{}
	// EmptyParty sets the party's member count to 0 (the roster stays).
	EmptyParty struct{}
)

func (InitRules) action()    {}
func (NewAdventure) action() {}
func (EndSession) action()   {}
func (EmptyParty) action()   {}

// Slots is the save slot screen: the saves, newest first.
type Slots struct {
	Names, Times []string
	Levels       []int  // the highest level of a member of each slot's party (0 unknown)
	CanSave      bool   // Save is taken (else loading only, or not now)
	LoadOnly     bool   // the party menu's Load
	Current      string // the last save ("" none): the list starts on it
}
