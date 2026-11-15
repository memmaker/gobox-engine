package rules

// AD&D names the Pool games' scripts and script commands still read by
// name (a thief's skills, the saving throws, the Ranger of PartySurprise,
// Knock at a door), declared here so the engine names them without
// importing a ruleset. They are 1e's, not every ruleset's: REFACTOR.md R2f
// puts them behind facts the ruleset gives by script field. The strings
// are the ids the rulesets key their tables by.

const ClassRanger Class = "ranger"

// por Thief[0..7] (the three the scripts read).
const (
	PickPockets Skill = "pick-pockets"
	FindTraps   Skill = "find-traps"
	Climb       Skill = "climb"
)

const (
	SaveParalysis Save = "paralysis"
	SavePetrify   Save = "petrification"
	SaveRod       Save = "rod"
	SaveBreath    Save = "breath"
	SaveSpell     Save = "spell"
)

const SpellKnock SpellID = "knock"
