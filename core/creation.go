package core

import "strconv"

// CreationRules is what a ruleset states about making a character, so that
// gobox runs the one creation flow (POR c70cbee's sheet: fields as drop-downs,
// Roll, Minimal, Keep, Quick Start) for every game. ONE-BINARY.md Track C2.
// It joins CharacterRules as Creation() once every ruleset implements it.
type CreationRules interface {
	// Fields are the sheet's choices for d in the game's order (POR: sex, race,
	// alignment, class; BR: race, sex, career, then the skills step). They
	// may depend on earlier picks (4e powers on class, BR career skills on career).
	Fields(d Draft) []FieldDef
	// Fits says whether option o of field f is legal for d (POR classFor, alignFor).
	Fits(d Draft, f FieldID, o int) bool
	// Enforce sets f to o and moves every dependant to its first legal
	// option (POR kind.enforce); a new kind rolls afresh, keeping name and pictures.
	Enforce(d Draft, f FieldID, o int, r *RNG) Draft
	// Roll rolls d's abilities and hit points in place (POR rollCharacter;
	// Minimal: rollMinimal). A minimum it cannot meet is a *MinimumError and
	// leaves d as it was (Felix 2026-11-11).
	Roll(d *Draft, k RollKind, r *RNG) error
	// Minimums are the lowest abilities d's picks allow (POR meetsMinStats).
	Minimums(d Draft) map[Ability]int32
	// QuickParty is the ready-made party the empty party menu offers.
	QuickParty() []Template
	// QuickRerolls is how often Quick Start rerolls a member after its first
	// roll until its Need is met; the last roll stays (POR quickPick.good 2000,
	// cotab and pod one fewer).
	QuickRerolls() int
	// SetScore sets ability a to v as typed on the sheet (POR setStatTyped,
	// limitStat). gobox parses "18/50" into Strength 18 and StrengthPercent 50.
	// The error says v is outside the race and class range.
	SetScore(d Draft, a Ability, v int32) (Draft, error)
	// Make keeps d (POR SaveCharacter): a new character when d.Of is NoChar
	// goes to the roster (its RosterID; CharID NoChar), else it updates d.Of
	// (RosterID NoRoster), keeping what the sheet does not show (items,
	// affects, spells).
	Make(d Draft) (CharID, RosterID, error)
	// Join puts roster character id in the party (POR the Add command's join).
	Join(id RosterID) (CharID, error)
	// Draft opens an existing character for modifying (its Of is c); NoChar
	// is the start of a new one: its Picks only (POR a human male fighter).
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
	RollFull     RollKind = iota // Roll
	RollMinimal                  // Minimal: until the minimums are met
	RollPointBuy                 // Point Buy: Draft.Buy's bases (PointBuySpent), then a roll's rules and hit points
)

// Template is one Quick Start member: its picks, name and pictures (pool names).
type Template struct {
	Picks              map[FieldID]int
	Name               string
	Head, Body         PicID
	IconHead, IconBody PicID
	Need               map[Ability]int32 // rerolled until these are met (POR quickPick.good)
}

// IconSize is the built combat icon's size (POR IconSize 1, 2).
type IconSize int

const (
	IconSmall IconSize = iota // one square
	IconLarge                 // two by two
)

// IconParts is how many colour parts a combat icon has (POR IconColour).
const IconParts = 6

// IconColour is one part's two shades, palette indexes (POR's low and high nibble).
type IconColour struct{ Low, High uint8 }

// Icon is the combat icon editor's result: the pool parts, their colours and
// size, or a dropped PNG that replaces the built icon.
type Icon struct {
	Head, Body PicID // combat icon pool pictures
	Colours    [IconParts]IconColour
	Size       IconSize
	PNG        PNG // a dropped PNG, encoded (nil: the built icon)
}

// PNG is an encoded PNG image (a dropped portrait or icon): the rules carry
// it, the presenter decodes it.
type PNG []byte

// MinimumError is Roll(RollMinimal) refusing: the minimum Want for Ability is
// above Best, the most the draft's picks allow (Tries 0), or was not met in
// Tries rolls. Pct is the strength percent of Want and Best (18/xx).
type MinimumError struct {
	Ability          Ability
	Want, Best       int32
	WantPct, BestPct int32
	Tries            int
}

func (m *MinimumError) Error() string {
	if m.Tries > 0 {
		return "Not met in " + strconv.Itoa(m.Tries) + " rolls"
	}
	if m.Ability == Strength && m.Best == 18 && m.BestPct == 0 { // POR create.go's special
		return "STR 18/xx is for fighters only, max 18"
	}
	name := string(m.Ability)
	if len(name) > 3 {
		name = name[:3]
	}
	b := []byte(name)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 'a' + 'A'
		}
	}
	return string(b) + " " + score(m.Want, m.WantPct) + " impossible here, max " + score(m.Best, m.BestPct)
}

// score is v as the sheet shows it, strength 18/xx with its percent (100: 18/00).
func score(v, pct int32) string {
	s := strconv.Itoa(int(v))
	if pct > 0 {
		p := pct % 100
		s += "/" + strconv.Itoa(int(p/10)) + strconv.Itoa(int(p%10))
	}
	return s
}
