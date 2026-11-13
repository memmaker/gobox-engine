package rules1e

import (
	"gobox/core"
	"gobox/core/dnd"
	"gobox/core/goldbox"
)

// The record's class, race, save and thief skill numbers, as the concepts the
// ruleset is asked about. Save files keep the numbers until the savefile move.
var (
	PorClass = [8]core.Class{ClsCleric: dnd.Cleric, ClsDruid: dnd.Druid, ClsFighter: dnd.Fighter, ClsPaladin: dnd.Paladin,
		ClsRanger: dnd.Ranger, ClsMage: dnd.Mage, ClsThief: dnd.Thief, ClsMonk: dnd.Monk}
	porSave  = goldbox.SaveOrder
	porSkill = [8]core.Skill{dnd.PickPockets, dnd.OpenLocks, dnd.FindTraps, dnd.MoveQuietly, dnd.Hide, dnd.Hear, dnd.Climb, dnd.ReadLanguages}
)

// Abilities are the core Abilities in POR's stat order (STR .. CHA).
var Abilities = [6]core.Ability{core.Strength, core.Intelligence, core.Wisdom, core.Dexterity, core.Constitution, core.Charisma}
