// Package core is the types the engine, a ruleset and an adventure share:
// the two interfaces (Rules, Adventure), the neutral ids, and the plain data
// that crosses them. It imports nothing of ours and holds no behaviour but
// the load check (Mismatches).
//
// DRAFT for review, nothing uses it yet. See ../ARCHITECTURE.md.
package core

// Neutral ids are typed string constants: readable in a load error and stable
// in a save, whatever order anyone's tables have. A ruleset that does not
// handle one says so in its Catalog, and the adventure fails at load.

// Ability is one of a character's scores.
type Ability string

const (
	Strength     Ability = "strength"
	Intelligence Ability = "intelligence"
	Wisdom       Ability = "wisdom"
	Dexterity    Ability = "dexterity"
	Constitution Ability = "constitution"
	Charisma     Ability = "charisma"
	Tech         Ability = "tech" // XXVc's seventh (eclbr.OffStats)
)

// Race is what a character is born as.
type Race string

const (
	RaceMonster  Race = "monster" // por raceMonster: no playable race
	RaceDwarf    Race = "dwarf"
	RaceElf      Race = "elf"
	RaceGnome    Race = "gnome"
	RaceHalfElf  Race = "half-elf"
	RaceHalfling Race = "halfling"
	RaceHalfOrc  Race = "half-orc"
	RaceHuman    Race = "human"
	RaceTerran   Race = "terran" // the XXVc names come from START.EXE (tab.RaceNames): the rest are to be listed at migration
)

// Class is one class or career.
type Class string

const (
	ClassNone  Class = ""
	Cleric     Class = "cleric"
	Druid      Class = "druid"
	Fighter    Class = "fighter"
	Paladin    Class = "paladin"
	Ranger     Class = "ranger"
	Mage       Class = "mage"
	Thief      Class = "thief"
	Monk       Class = "monk"
	Warrior    Class = "warrior" // XXVc careers
	Rogue      Class = "rogue"
	Medic      Class = "medic"
	Engineer   Class = "engineer"
	RocketJock Class = "rocket-jock"
)

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

const (
	LawfulGood     Alignment = "lawful-good"
	LawfulNeutral  Alignment = "lawful-neutral"
	LawfulEvil     Alignment = "lawful-evil"
	NeutralGood    Alignment = "neutral-good"
	TrueNeutral    Alignment = "true-neutral"
	NeutralEvil    Alignment = "neutral-evil"
	ChaoticGood    Alignment = "chaotic-good"
	ChaoticNeutral Alignment = "chaotic-neutral"
	ChaoticEvil    Alignment = "chaotic-evil"
)

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
)

// Side is who a combatant fights for (por Character.Side, Auto; the Side word
// scripts write at 268).
type Side string

const (
	SideParty    Side = "party" // the player steers it
	SideAuto     Side = "auto"  // on the party's side, the computer steers it
	SideMonsters Side = "monsters"
)

// Skill is a named skill, thief or XXVc. XXVc has 85 (eclbr.OffSkills+s);
// only those scripts use are listed so far, the others come at migration.
type Skill string

const (
	SkillPickPockets   Skill = "pick-pockets" // por Thief[0..7]
	SkillOpenLocks     Skill = "open-locks"
	SkillFindTraps     Skill = "find-traps"
	SkillMoveQuietly   Skill = "move-quietly"
	SkillHide          Skill = "hide"
	SkillHear          Skill = "hear"
	SkillClimb         Skill = "climb"
	SkillReadLang      Skill = "read-languages"
	SkillBattleTactics Skill = "battle-tactics" // BR skill 31 (combat.go tactics)
)

// Save is a kind of saving throw. 1e has five, XXVc eight (eclbr.OffSaves);
// Catalog says which a ruleset rolls.
type Save string

const (
	SaveParalysis Save = "paralysis"
	SavePetrify   Save = "petrification"
	SaveRod       Save = "rod"
	SaveBreath    Save = "breath"
	SaveSpell     Save = "spell"
)

// Condition is a lasting effect a script may ask about. Only the ones a
// script does ask about are here: por, cotab and pod name 125-142 affects of
// their own, which stay private to their rulesets (ARCHITECTURE.md).
type Condition string

const (
	CondFindTraps     Condition = "find-traps"       // por CheckParty 19
	CondResistFire    Condition = "resist-fire"      // cotab FindAffect 20
	CondSticksSnakes  Condition = "sticks-to-snakes" // sotsb CheckParty 3
	CondInvisible     Condition = "invisible"        // pod FindAffect 25
	CondInvisibleRing Condition = "invisible-ring"   // pod FindAffect 56
	CondMirrorImage   Condition = "mirror-image"     // pod FindAffect 72
	CondGazeImmune    Condition = "gaze-immune"      // pod FindAffect 126
	CondDetectMagic   Condition = "detect-magic"     // pod gameMagicCheck
)

// Feature is a whole subsystem an adventure may need of a ruleset.
type Feature string

const (
	FeatureSpaceCombat Feature = "space-combat" // BR OpSpaceFight (not yet in this interface: open question)
)

// Handles. A ruleset hands out CharID, the engine stores it. CreatureRef,
// ItemSetRef, MapRef and AreaID are the adventure's own.
type (
	CharID  int32 // a character or monster the ruleset holds; survives its save
	AreaID  int32 // an area (the ECL block number today)
	DataSet int32 // the original's numbered file sets (por Set: ECL3, GEO3, ...)
)

// NoChar is no character.
const NoChar CharID = 0

// CreatureRef names a monster or NPC of the adventure (por LoadMonster: MON<set>CHA, index).
type CreatureRef struct {
	Set   DataSet
	Index int32
}

// ItemSetRef names a list of items of the adventure (por opTreasure: ITEM<set>, block).
type ItemSetRef struct {
	Set   DataSet
	Block int32
}

// MapRef names a walking map (por geo.go:75: GEO<set>, block).
type MapRef struct {
	Set   DataSet
	Block int32
}
