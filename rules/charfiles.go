package rules

// CharacterFiles is one character as saved, each file's bytes (nil: the
// file is not there). Their layout is the ruleset's business.
type CharacterFiles struct {
	Record   []byte // .cha (roster) or .sav (game): the record, then the pool trailer
	Items    []byte // .itm
	Affects  []byte // .spc
	Plan     []byte // .mem
	Dual     []byte // .dual
	Powers   []byte // .pow
	Portrait []byte // .png
	Icon     []byte // .icon.png
}

// SavedRecord is a character's saved record (a .cha or .sav): the ruleset's
// own bytes, then the pool trailer.
type SavedRecord []byte

// FileFormat is the optional capability of a Rules that saves characters as
// files: the engine moves the bytes and asks these typed questions of them.
// The engine reaches it by a type assertion on its Rules (Rules stays
// without it: a ruleset that saves no files need not have it).
type FileFormat interface {
	// Pack is the canonical character of its files (Import's input).
	Pack(CharacterFiles) (CharacterFile, error)
	// Unpack is the canonical character as its files; game: a game save's
	// (the items file carries the shared pack).
	Unpack(c CharacterFile, game bool) (CharacterFiles, error)
	// IsRecord says whether the bytes are one of this ruleset's saved records (a roster file that is not one is skipped).
	IsRecord(record SavedRecord) bool
	// RosterOf is the roster number stored in a saved record (NoRoster: none).
	RosterOf(record SavedRecord) RosterID
	// WithRoster is the saved record with its roster number replaced.
	WithRoster(record SavedRecord, id RosterID) SavedRecord
	// SavedLevels are the class levels shown for a saved record.
	SavedLevels(record SavedRecord) []int
}
