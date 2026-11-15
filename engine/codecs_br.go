package engine

import (
	"fmt"

	"gobox/adventure"
	"gobox/ecl"
	"gobox/rules"
)

// brCareerFirst is the number of the catalog's first career: 0 is none.
const brCareerFirst = 1

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
	// Party[40] (charVar r[40]): TCareer, the career's place in the catalog from 1, 0 none; the canonical
	// SelCharClass is the place from 0 (the Pool's record number), -1 none.
	"BrSelCharClass": {toGame: func(_ *Engine, c int64) int64 { return c + brCareerFirst }},
	// Party[76] (SpecialReadZ1 $4C): 128 while the character is a companion.
	"BrSelCompanion": {toGame: func(_ *Engine, c int64) int64 {
		if c != 0 {
			return adventure.BRCompanionRead
		}
		return 0
	}},
}
