// Package dnd is the Forgotten Realms / 1st edition concepts: what POR, COTAB,
// SotSB and POD adventures and the 1e ruleset name. The numbers behind each
// are the ruleset's. MonsterID and ItemID rosters are generated at import.
package rules1e

import "gobox/rules"

const (
	CoreRaceDwarf    rules.Race = "dwarf"
	CoreRaceElf      rules.Race = "elf"
	CoreRaceGnome    rules.Race = "gnome"
	CoreRaceHalfElf  rules.Race = "half-elf"
	CoreRaceHalfling rules.Race = "halfling"
	CoreRaceHalfOrc  rules.Race = "half-orc"
	CoreRaceHuman    rules.Race = "human"
	CoreRaceMonster  rules.Race = "monster" // por Race 0: a record without a player race
)

const (
	CoreClassCleric  rules.Class = "cleric"
	CoreClassDruid   rules.Class = "druid"
	CoreClassFighter rules.Class = "fighter"
	CoreClassPaladin rules.Class = "paladin"
	CoreClassRanger  rules.Class = "ranger"
	CoreClassMage    rules.Class = "mage"
	CoreClassThief   rules.Class = "thief"
	CoreClassMonk    rules.Class = "monk"
)

const (
	LawfulGood     rules.Alignment = "lawful-good"
	LawfulNeutral  rules.Alignment = "lawful-neutral"
	LawfulEvil     rules.Alignment = "lawful-evil"
	NeutralGood    rules.Alignment = "neutral-good"
	TrueNeutral    rules.Alignment = "true-neutral"
	NeutralEvil    rules.Alignment = "neutral-evil"
	ChaoticGood    rules.Alignment = "chaotic-good"
	ChaoticNeutral rules.Alignment = "chaotic-neutral"
	ChaoticEvil    rules.Alignment = "chaotic-evil"
)

// por Thief[0..7].
const (
	PickPockets   rules.Skill = "pick-pockets"
	OpenLocks     rules.Skill = "open-locks"
	FindTraps     rules.Skill = "find-traps"
	MoveQuietly   rules.Skill = "move-quietly"
	Hide          rules.Skill = "hide"
	Hear          rules.Skill = "hear"
	Climb         rules.Skill = "climb"
	ReadLanguages rules.Skill = "read-languages"
)

const (
	SaveParalysis rules.Save = "paralysis"
	SavePetrify   rules.Save = "petrification"
	SaveRod       rules.Save = "rod"
	SaveBreath    rules.Save = "breath"
	SaveSpell     rules.Save = "spell"
)

// The conditions scripts ask about; the 125-142 others stay private to the ruleset.
const (
	CondFindTraps     rules.Condition = "find-traps"       // por CheckParty 19
	CondResistFire    rules.Condition = "resist-fire"      // cotab FindAffect 20
	CondSticksSnakes  rules.Condition = "sticks-to-snakes" // sotsb CheckParty 3
	CondInvisible     rules.Condition = "invisible"        // pod FindAffect 25
	CondInvisibleRing rules.Condition = "invisible-ring"   // pod FindAffect 56
	CondMirrorImage   rules.Condition = "mirror-image"     // pod FindAffect 72
	CondGazeImmune    rules.Condition = "gaze-immune"      // pod FindAffect 126
	CondDetectMagic   rules.Condition = "detect-magic"     // pod gameMagicCheck
	CondDancing       rules.Condition = "dancing"          // AC -4, every save fails, die still rolled: pod affects.go:1606, rules.go:461
	CondFeatherFall   rules.Condition = "feather-fall"     // por effectTable 0x4900 (eclops.go)
)

const (
	Copper = rules.CoinCopper
	Gold   = rules.CoinGold
)

// A sample of the rosters the importer generates, numbers assigned once
// (MON*CHA; a name with differing stat blocks gets one id per block).
const (
	Hobgoblin rules.MonsterID = 1
	Kobold    rules.MonsterID = 2
	LongSword rules.ItemID    = 1
)
