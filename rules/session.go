package rules

// The session: init, new game, save slots and what a save records.

import "strings"

// InitRules readies the rules' tables for a new session (por Init:
// SpellTableInit, SetupAffectHandlers).
type InitRules struct{}

// LoadItemTable reads the base item records from the ruleset's item source
// (por's ITEMS file, DS:5582).
type LoadItemTable struct{}

func (LoadItemTable) action() {}

// SetSeed sets the rules' random seed (por Randomize, 0AF8:1247).
type SetSeed struct{ Seed uint32 }

func (SetSeed) action() {}

// NewAdventure puts the rules' adventure state back to a new game's: the
// game and area records, an empty purse, loot and pool.
type NewAdventure struct{}

// EndSession closes what a session had open: camp, training, learning.
type EndSession struct{}

func (InitRules) action() {}

func (NewAdventure) action() {}

func (EndSession) action() {}

// Slots is the save slot screen: the saves, newest first.
type Slots struct {
	Names, Times []string
	Levels       []int  // the highest level of a member of each slot's party (0 unknown)
	CanSave      bool   // Save is taken (else loading only, or not now)
	LoadOnly     bool   // the party menu's Load
	Current      string // the last save ("" none): the list starts on it
}

// Pending is the ruleset's effects queued outside Resolve; taking them empties
// the queue (interim, until every rule call returns an Outcome).
type Pending interface{ TakePending() []Effect }

// LogJoin is a battle log line's next part after s: a space between, msg trimmed.
func LogJoin(s, msg string) string {
	msg = strings.TrimSpace(msg)
	if s != "" && msg != "" {
		return s + " " + msg
	}
	return s + msg
}

// RecordView is the engine's view state the next Save records with the
// area: the running area and the wall slots (a block, a slot; 0xFFFF none).
type RecordView struct {
	AreaKind       uint16
	WallsA, WallsB [3]uint16
}

// TakeSavedWalls reads the wall slots a Load brought and forgets them; the
// Outcome holds SavedWalls.
type TakeSavedWalls struct{}

func (RecordView) action() {}

func (TakeSavedWalls) action() {}

// SavedWalls are the wall slots of a loaded save (block, slot).
type SavedWalls struct{ A, B [3]uint16 }

func (SavedWalls) effect() {}
