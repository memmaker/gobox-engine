package engine

import (
	"gobox/adventure"
	"gobox/ecl"
)

// sotsbCodecs are sotsb's encodings of canonical fields (its layout's enc=).
var sotsbCodecs = map[ecl.Encoding]codec{
	// SpecialReadZ1 $100 (sotsb-re eclmenu.pas): 0 while charNotFound (the last SELECT MEMBER matched nobody, until
	// the next that does), else the status 1..7, 0 for one with no number; the write is adventure.sotsbStatusWrite.
	"SotsbSelStatus": {
		toGame: func(e *Engine, canonical int64) int64 {
			if e.es.memberMissing {
				return 0
			}
			return adventure.SotsbStatus[adventure.StatusKinds[canonical]]
		},
	},
	// SpecialWriteZ1 $1B..$65: memorizedSpells[address-$1B], slot 0 included; POR's field starts at slot 1.
	"SotsbSelMemorized": {element: func(i int) int { return i - 1 }},
	"SotsbSymbolSet":    {}, // decoded only (adventure.Canonical)
}
