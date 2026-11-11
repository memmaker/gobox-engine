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
	// SetScore sets ability a to v as typed on the sheet (POR setStatTyped,
	// limitStat). gobox parses "18/50" into Strength 18 and StrengthPercent 50.
	// The error says v is outside the race and class range.
	SetScore(d Draft, a Ability, v int32) (Draft, error)
	// Make keeps d as a new character (POR SaveCharacter).
	Make(d Draft) (CharID, error)
	// Draft opens an existing character for modifying.
	Draft(c CharID) (Draft, error)
}

// FieldKind is how a field is picked.
type FieldKind int

const (
	FieldChoice FieldKind = iota // one of Options (Draft.Picks)
	FieldSkills                  // points spread over Options up to Budget (Draft.Points)
)

// FieldID names one creation field; the ruleset lists them in Fields.
type FieldID int

// FieldDef is one field: its id, what the sheet shows, its options.
type FieldDef struct {
	ID      FieldID
	Kind    FieldKind
	Label   string
	Options []string
	Budget  int32 // FieldSkills: the points to spend
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
	Need               map[Ability]int32 // rerolled until these are met (POR quickPick.good)
}
