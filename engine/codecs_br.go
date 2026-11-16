package engine

import (
	"fmt"

	"gobox/adventure"
	"gobox/ecl"
	"gobox/rules"
)

// brCareerFirst is the number of the catalog's first career: 0 is none.
const brCareerFirst = 1

// brRaceFirst is the number of the catalog's first race: 0 is RaceMonster (types.pas TRace).
const brRaceFirst = 1

// zone 0's shipCombatOutcome (shipcbt.pas 392, 396, 1106, 1272) and postCombatMode (c_postcb.pas) values.
const (
	brOutcomeFoeGone   = 0 // shipCombatOutcome: the foe destroyed or boarded
	brOutcomeTeamLost  = 1 // the team's hull lost, or the foe out of range
	brPostCombatNormal = 0 // C_PostCb.NORMAL
	brPostCombatBoard  = 1 // C_PostCb.BOARDED
)

// brCodecs are BR's encodings of canonical fields (its layout's enc=).
var brCodecs = map[ecl.Encoding]codec{
	// Party[256] (ecl.go charVar): r[212] the status, 0 while no character is selected (noChar: the last
	// SELECT MEMBER matched nobody, until the next one that does).
	"BrSelStatus": {toGame: func(e *Engine, c int64) int64 {
		if e.es.memberMissing {
			return 0
		}
		st := adventure.StatusKinds[c]
		n, ok := adventure.BRStatus[st]
		if !ok {
			panic(fmt.Sprintf("br: SelStatus: no TStatus for %s", st))
		}
		return n
	}},
	// zone0 areaKind: the canonical rules.PlaceTraits as BR keeps them (the stored byte is never rebuilt from them).
	"BrTravelKind": {toGame: func(_ *Engine, c int64) int64 { return int64(adventure.BRAreaKind(rules.PlaceTraits(c))) }},
	// zone0 shipCombatOutcome: the canonical value is the typed rules.ShipEnd.
	"BrShipEnd": {toGame: func(_ *Engine, c int64) int64 {
		switch rules.ShipEnd(c) {
		case rules.ShipFoeDestroyed, rules.ShipBoarded:
			return brOutcomeFoeGone
		case rules.ShipHullLost, rules.ShipReceded:
			return brOutcomeTeamLost
		}
		panic(fmt.Sprintf("br: ShipOutcome of ship end %d", c))
	}},
	// zone0 postCombatMode as ShipCombat leaves it: NORMAL, or BOARDED once the team boarded.
	"BrPostCombat": {toGame: func(_ *Engine, c int64) int64 {
		if rules.ShipEnd(c) == rules.ShipBoarded {
			return brPostCombatBoard
		}
		return brPostCombatNormal
	}},
	// Party[40] (charVar r[40]): TCareer, the career's place in the catalog from 1, 0 none; the canonical
	// SelCharClass is the place from 0 (the Pool's record number), -1 none.
	"BrSelCharClass": {toGame: func(_ *Engine, c int64) int64 { return c + brCareerFirst }},
	// Party[39] (charVar r[39]): TRace, RaceMonster 0 then the races in the catalog's order from 1; the canonical
	// SelRace is the place from 0, none below it.
	"BrSelRace": {toGame: func(_ *Engine, c int64) int64 { return int64(int16(c)) + brRaceFirst }},
	// Party[76] (SpecialReadZ1 $4C): 128 while the character is a companion.
	"BrSelCompanion": {toGame: func(_ *Engine, c int64) int64 {
		if c != 0 {
			return adventure.BRCompanionRead
		}
		return 0
	}},
}
