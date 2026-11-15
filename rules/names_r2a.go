package rules

// Names every Gold Box ruleset shares, declared here so the engine can name
// them without a ruleset (REFACTOR.md R2a). The strings are the ids the
// rulesets key their tables by.

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
