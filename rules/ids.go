// Package rules is the types the engine, a ruleset and an adventure share:
// the Rules interface, the neutral ids, and the plain data
// that crosses them. It imports nothing of ours and holds no behaviour but
// the RNG. Concepts of one world are in rules1e (dnd.go)
// and rules/xxvc; an adventure and its ruleset share one of them.
//
// DRAFT for review, nothing uses it yet. See ../ARCHITECTURE.md.
package rules

// Neutral ids are typed string constants: readable in a load error and stable
// in a save, whatever order anyone's tables have. A ruleset that does not
// handle one says so in its Catalog,

// Ability is one of a character's scores.
type Ability string

const (
	Strength     Ability = "strength"
	Intelligence Ability = "intelligence"
	Wisdom       Ability = "wisdom"
	Dexterity    Ability = "dexterity"
	Constitution Ability = "constitution"
	Charisma     Ability = "charisma"
	// StrengthPercent is the xx of an exceptional 18/xx strength (0: none, 100: 18/00).
	StrengthPercent Ability = "strength percent"
)

// Race is what a character is born as.
type Race string

const RaceMonster Race = "monster" // por raceMonster: no playable race

// Class is one class or career.
type Class string

const ClassNone Class = ""

// ClassSet is a character's classes at once (por clsFighterMageThief is
// three); unused places are ClassNone. An array, so that it compares.
type ClassSet [3]Class

// Sex is a character's sex.
type Sex string

const (
	Male   Sex = "male"
	Female Sex = "female"
)

// Alignment is a character's alignment.
type Alignment string

// Status is a character's condition as scripts and sheets name it (por
// statusOkay ...; names are the DS:1047 table's).
type Status string

const (
	StatusOkay        Status = "okay"
	StatusAnimated    Status = "animated"
	StatusRunning     Status = "running"
	StatusUnconscious Status = "unconscious"
	StatusDying       Status = "dying"
	StatusDead        Status = "dead"
	StatusStoned      Status = "stoned"
	StatusGone        Status = "gone"
	StatusTempGone    Status = "tempgone" // por statusTempGone: away for now
	// StatusInactive: out of the active members, the record's status kept
	// (por SpecialWriteZ1 0x100 with a value >= 0x80 but 0x87).
	StatusInactive Status = "inactive"
)

// Side is who a combatant fights for (por Character.Side, Auto; the Side word
// scripts write at 268).
type Side string

const (
	SideParty    Side = "party" // the player steers it
	SideAuto     Side = "auto"  // on the party's side, the computer steers it
	SideMonsters Side = "monsters"
)

// Skill is a named skill, thief or XXVc. XXVc has 84 (core/xxvc.Skills);
// 1e names its thief skills, core/xxvc all of XXVc.
type Skill string

// Save is a kind of saving throw. 1e has five, XXVc eight (eclbr.OffSaves);
// Catalog says which a ruleset rolls.
type Save string

// Condition is a lasting effect a script may ask about. Only the ones a
// script does ask about are here: por, cotab and pod name 125-142 affects of
// their own, which stay private to their rulesets (ARCHITECTURE.md).
type Condition string

// Feature is a whole subsystem an adventure may need of a ruleset.
type Feature string

const (
	FeatureSpaceCombat Feature = "space-combat" // BR OpSpaceFight: Rules.Ships
)

// Handles. A ruleset hands out CharID; the engine stores it.
type (
	CharID  int32 // a character or monster the ruleset holds; survives its save
	AreaID  int32 // an area of the adventure
	DataSet int32 // the original's numbered file sets (por Set: ECL3, GEO3, ...)
)

// NoChar is no character: what SelectMember past the party's end selects.
const NoChar CharID = 0

// SpellID is a spell or an item power that casts like one: rules1e lists them (dndspells.go).
type SpellID string

// RosterID is a roster character's own stable number, never derived from its
// name (POR: the .cha trailer's "rosterid").
type RosterID uint32

// NoRoster is no roster character.
const NoRoster RosterID = 0

// PicID is one entry of the shared portrait and combat icon collection
// (gobox/sprites, series pool or buck-rogers): its sprite name, "head_012".
type PicID string

// Things named by the adventure and given numbers by the ruleset. The
// constants are in rules1e and core/xxvc, generated at import. Scripts carry
// these ids, resolved by the decompiler, no set or index (ARCHITECTURE.md).
// Their numbers are stable by assignment (saves hold them), never an original's index.
type (
	MonsterID  int32 // one per distinct stat block
	ItemID     int32 // the base item; plus and charges are ItemRef's
	ShipID     int32
	TreasureID int32 // a list of placed items
)
