// Package dnd is the Forgotten Realms / 1st edition concepts: what POR, COTAB,
// SotSB and POD adventures and the 1e ruleset name. The numbers behind each
// are the ruleset's. MonsterID and ItemID rosters are generated at import.
package dnd

import "gobox/core"

const (
	Dwarf    core.Race = "dwarf"
	Elf      core.Race = "elf"
	Gnome    core.Race = "gnome"
	HalfElf  core.Race = "half-elf"
	Halfling core.Race = "halfling"
	HalfOrc  core.Race = "half-orc"
	Human    core.Race = "human"
)

const (
	Cleric  core.Class = "cleric"
	Druid   core.Class = "druid"
	Fighter core.Class = "fighter"
	Paladin core.Class = "paladin"
	Ranger  core.Class = "ranger"
	Mage    core.Class = "mage"
	Thief   core.Class = "thief"
	Monk    core.Class = "monk"
)

const (
	LawfulGood     core.Alignment = "lawful-good"
	LawfulNeutral  core.Alignment = "lawful-neutral"
	LawfulEvil     core.Alignment = "lawful-evil"
	NeutralGood    core.Alignment = "neutral-good"
	TrueNeutral    core.Alignment = "true-neutral"
	NeutralEvil    core.Alignment = "neutral-evil"
	ChaoticGood    core.Alignment = "chaotic-good"
	ChaoticNeutral core.Alignment = "chaotic-neutral"
	ChaoticEvil    core.Alignment = "chaotic-evil"
)

// por Thief[0..7].
const (
	PickPockets   core.Skill = "pick-pockets"
	OpenLocks     core.Skill = "open-locks"
	FindTraps     core.Skill = "find-traps"
	MoveQuietly   core.Skill = "move-quietly"
	Hide          core.Skill = "hide"
	Hear          core.Skill = "hear"
	Climb         core.Skill = "climb"
	ReadLanguages core.Skill = "read-languages"
)

const (
	SaveParalysis core.Save = "paralysis"
	SavePetrify   core.Save = "petrification"
	SaveRod       core.Save = "rod"
	SaveBreath    core.Save = "breath"
	SaveSpell     core.Save = "spell"
)

// The conditions scripts ask about; the 125-142 others stay private to the ruleset.
const (
	CondFindTraps     core.Condition = "find-traps"       // por CheckParty 19
	CondResistFire    core.Condition = "resist-fire"      // cotab FindAffect 20
	CondSticksSnakes  core.Condition = "sticks-to-snakes" // sotsb CheckParty 3
	CondInvisible     core.Condition = "invisible"        // pod FindAffect 25
	CondInvisibleRing core.Condition = "invisible-ring"   // pod FindAffect 56
	CondMirrorImage   core.Condition = "mirror-image"     // pod FindAffect 72
	CondGazeImmune    core.Condition = "gaze-immune"      // pod FindAffect 126
	CondDetectMagic   core.Condition = "detect-magic"     // pod gameMagicCheck
	CondFeatherFall   core.Condition = "feather-fall"     // por effectTable 0x4900 (eclops.go)
)

const (
	Copper   core.Coin = "copper"
	Silver   core.Coin = "silver"
	Electrum core.Coin = "electrum"
	Gold     core.Coin = "gold"
	Platinum core.Coin = "platinum"
	Gems     core.Coin = "gems"
	Jewelry  core.Coin = "jewelry"
)

// A sample of the rosters the importer will generate (MON*CHA names; a name
// with differing stat blocks gets one id per block: scan in ARCHITECTURE.md).
const (
	Hobgoblin core.MonsterID = "hobgoblin"
	Kobold    core.MonsterID = "kobold"
	LongSword core.ItemID    = "long-sword"
)
