package rules1e

import (
	"gobox/core"
)

// The Gold Box character record's race and alignment bytes (POR's order,
// the scripts' SelRace/SelAlignment words).
var (
	RecordRace = [8]core.Race{CoreRaceMonster, CoreRaceDwarf, CoreRaceElf, CoreRaceGnome, CoreRaceHalfElf, CoreRaceHalfling, CoreRaceHalfOrc, CoreRaceHuman}
	// RecordAlignment: LG, LN, LE, NG, N, NE, CG, CN, CE.
	RecordAlignment = [9]core.Alignment{LawfulGood, LawfulNeutral, LawfulEvil, NeutralGood, TrueNeutral,
		NeutralEvil, ChaoticGood, ChaoticNeutral, ChaoticEvil}
)

// SaveOrder is the saving throws in the Gold Box games' order: a record's
// save bytes and a script's save number (DAMAGE) index it.
var SaveOrder = [5]core.Save{SaveParalysis, SavePetrify, SaveRod, SaveBreath, SaveSpell}
