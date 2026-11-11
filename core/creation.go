package core

// CreationRules is what a ruleset states about making a character, so that
// gobox runs the one creation flow (POR c70cbee's sheet: fields as drop-downs,
// Roll, Minimal, Keep, Quick Start) for every game. ONE-BINARY.md Track C2.
// It joins CharacterRules as Creation() once every ruleset implements it.
type CreationRules interface {
	// Fields are the sheet's choices in the game's order (POR: sex, race,
	// alignment, class; BR: race, sex, career, then the skills step).
	Fields() []FieldDef
	// Fits says whether option o of field f is legal for d (POR classFor, alignFor).
	Fits(d Draft, f FieldID, o int) bool
	// Enforce sets f to o and moves every dependant to its first legal
	// option (POR kind.enforce); a new kind rolls afresh, keeping name and pictures.
	Enforce(d Draft, f FieldID, o int, r *RNG) Draft
	// Roll rolls the abilities and hit points (POR rollCharacter; Minimal: rollMinimal).
	Roll(d Draft, k RollKind, r *RNG) Draft
	// Minimums are the lowest abilities d's picks allow (POR meetsMinStats).
	Minimums(d Draft) map[Ability]int32
	// QuickParty is the ready-made party the empty party menu offers.
	QuickParty() []Template
}

// FieldID names one creation field; the ruleset lists them in Fields.
type FieldID int

// FieldDef is one field: its id, what the sheet shows, its options.
type FieldDef struct {
	ID      FieldID
	Label   string
	Options []string
}

// RollKind is how Roll rolls.
type RollKind int

const (
	RollFull    RollKind = iota // Roll
	RollMinimal                 // Minimal: until the minimums are met
)

// Template is one Quick Start member: its picks, name and pictures (pool names).
type Template struct {
	Picks              map[FieldID]int
	Name               string
	Head, Body         string
	IconHead, IconBody string
}
