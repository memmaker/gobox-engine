package rules1e

import (
	"gobox/rules"
)

// Types the rules fill and the game package does not have: rules-local
// copies of what the front module once shared.

// BookSpell is a spell with a count, and the slots of its level.
type BookSpell struct {
	ID      int // the spell (powerPlan, powerScribe X)
	Name    string
	Level   string // its group, as "Cleric Level 1"
	Count   int
	Planned int // spells planned of its level
	Slots   int // slots of its level
}

// Slots is the save slot screen (core's).
type Slots = rules.Slots

// Stat is one of a rule set's figures on the sheet (game.Stat).
type Stat = rules.Figure

// The creation's fields (SetField).
const (
	FieldSex = iota
	FieldRace
	FieldAlign
	FieldClass
)

// Policy (extra) is a member's standing answer to one of the rule set's
// reactions, chosen beforehand on the sheet: nothing asks mid-turn.
type Policy uint8

const (
	Always Policy = iota // the zero value: an older save's, a foe's
	Never
	OnHit // only when it turns a hit (the rule set says which way)
)

// Step is one square of a walk; Free: leaving it lets an enemy take a free attack.
type Step = rules.Step

// Maneuver is one of the bar's Maneuver list.
type Maneuver struct {
	ID    int
	Name  string
	Tip   string // what it does, one line a sentence
	Aim   Aim
	Range int // squares, AimFoe, AimAlly and AimArea
	Area  int // AimArea: the burst's squares round the square aimed at
	Move  int // AimWeapon: squares the member may move first; the player aims at the square (Aimed; its own: none)
}

// Square is one map square.
type Square = rules.Square

// Aim is what a maneuver aims at.
type Aim int

const (
	AimWeapon Aim = iota // a foe, with c's next weapon attack (the bar's Attack)
	AimSelf              // no aim
	AimFoe               // a foe within Range
	AimAlly              // a member of c's side within Range (c too)
	AimArea              // a square within Range: who stands within Area squares of it (Aimed, Burst)
)

func (p Policy) String() string { return [...]string{"Always", "Never", "On hit"}[p%3] }
