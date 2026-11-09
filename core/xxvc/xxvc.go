// Package xxvc is the Buck Rogers concepts: what BR1, BR2 and the XXVc ruleset
// name. Rosters not listed here (races, 85 skills, monsters, items, ships) are
// generated at import from START.EXE and MON0CHA, once.
package xxvc

import "gobox/core"

const Tech core.Ability = "tech" // the seventh score (eclbr.OffStats)

const Terran core.Race = "terran"

const (
	Warrior    core.Class = "warrior"
	Rogue      core.Class = "rogue"
	Medic      core.Class = "medic"
	Engineer   core.Class = "engineer"
	RocketJock core.Class = "rocket-jock"
)

const BattleTactics core.Skill = "battle-tactics" // BR skill 31 (combat.go tactics)

const Credits core.Coin = "credits"

// A ship's systems, each with a now and a max (eclbr.Ship 22+2i, 34+2i).
// Index 4 is fuel; the other names to confirm at migration.
const Fuel core.ShipSystem = "fuel"
