package core

// Additions of sweep Sfiles (fold E-R6): what por's file, init, sound and text
// code reads from the rules.

// InFight asks whether a fight is on (Yes).
type InFight struct{}

func (InFight) check() {}

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
