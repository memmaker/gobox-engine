package rules

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
	// Rolls are the ways the ruleset rolls abilities for d, the ones a front may offer; the others return an error from Roll.
	Rolls(d Draft) []RollKind
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
	// Spend moves by points (+1 spends one, -1 takes one back) onto option o of the FieldSkills field f and
	// returns the draft after it, the ruleset's rules applied (the limits, the prerequisites, what a change
	// resets). A refusal is a *SpendRefusal and leaves d as it was.
	Spend(d Draft, f FieldID, o int, by int32) (Draft, error)
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
	Budget  int32      // FieldSkills: the points to spend
	Max     int32      // FieldSkills: the most points one option may take
	Rows    []FieldRow // FieldSkills: what each option stands at in the draft, one per option
	// Alignment: the field picks the alignment (the sheet places it on a row of its own; a sheet without one lists its fields one per row).
	Alignment bool
}

// PointsLeft is what the FieldSkills field f still has to spend in d: Left, or its whole Budget before any is spent.
func (d Draft) PointsLeft(f FieldDef) int32 {
	if v, ok := d.Left[f.ID]; ok {
		return v
	}
	return f.Budget
}

// FieldRow is what option of a FieldSkills field stands at in the draft: its group heading, its value (Value, and
// Bonus on it, Total as the sheet counts it), whether its prerequisites are met and what it still wants.
type FieldRow struct {
	Group        string
	Value, Bonus int32
	Total        int32
	Open         bool // the prerequisites are met
	Wants        []Want
}

// Want is a prerequisite not met: Level points in the option named Name.
type Want struct {
	Name  string
	Level int32
}

// SpendReason is why Spend refuses a point.
type SpendReason int

const (
	SpendNoPoints  SpendReason = iota // the field has none left
	SpendWants                        // a prerequisite is not met
	SpendAtMost                       // the option is at the most it may take
	SpendNoNewItem                    // no more options may be started
	SpendNoOption                     // no such option or step
)

// SpendRefusal is Spend refusing: Reason for a front or a test, Text in the ruleset's own words.
type SpendRefusal struct {
	Reason SpendReason
	Text   string
}

func (e *SpendRefusal) Error() string { return e.Text }

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

// The creation sheet's character (por's sheet draws a record of the draft as
// the selected member, outside the party).
type (
	// SheetDraft makes the character d draws as, without dice; Outcome.Created is it.
	SheetDraft struct{ Draft Draft }
	// SeatTemp puts Who at the end of the party for the sheet, not as a member joining.
	SeatTemp struct{ Who CharID }
	// TrimParty drops every member after the first Size (SeatTemp's undo).
	TrimParty struct{ Size int32 }
	// ReplaceMember puts Who in the party's place At.
	ReplaceMember struct {
		At  int32
		Who CharID
	}
	// HoldCopy keeps a copy of Who as it is, for RestoreCopy.
	HoldCopy struct{ Who CharID }
	// RestoreCopy makes Who again as HoldCopy kept it.
	RestoreCopy struct{ Who CharID }
	// SetMinimums takes the minimums typed for Minimal ("STR INT WIS DEX CON
	// CHA"; a field that is not a value keeps the last one).
	SetMinimums struct{ Text string }
)

func (SheetDraft) action()    {}
func (SeatTemp) action()      {}
func (TrimParty) action()     {}
func (ReplaceMember) action() {}
func (HoldCopy) action()      {}
func (RestoreCopy) action()   {}
func (SetMinimums) action()   {}

// DraftIconFact is Who's combat icon as a Draft holds it (Icon).
type DraftIconFact struct{}

func (DraftIconFact) fact() {}
func (Icon) value()         {}
