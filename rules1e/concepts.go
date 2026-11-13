package rules1e

import (
	"gobox/core"
)

// The record's class, race, save and thief skill numbers, as the concepts the
// ruleset is asked about. Save files keep the numbers until the savefile move.
var (
	PorClass = [8]core.Class{ClsCleric: CoreClassCleric, ClsDruid: CoreClassDruid, ClsFighter: CoreClassFighter, ClsPaladin: CoreClassPaladin,
		ClsRanger: CoreClassRanger, ClsMage: CoreClassMage, ClsThief: CoreClassThief, ClsMonk: CoreClassMonk}
	porSave  = SaveOrder
	porSkill = [8]core.Skill{PickPockets, OpenLocks, FindTraps, MoveQuietly, Hide, Hear, Climb, ReadLanguages}
)

// Abilities are the core Abilities in POR's stat order (STR .. CHA).
var Abilities = [6]core.Ability{core.Strength, core.Intelligence, core.Wisdom, core.Dexterity, core.Constitution, core.Charisma}
