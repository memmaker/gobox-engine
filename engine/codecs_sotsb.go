package engine

import (
	"gobox/adventure"
	"gobox/ecl"
	"gobox/rules"
)

// sotsbCodecs are sotsb's encodings of canonical fields (its layout's enc=).
var sotsbCodecs = map[ecl.Encoding]codec{
	// SpecialReadZ1 $100 (sotsb-re eclmenu.pas) reads the status 1..7; SpecialWriteZ1 $100 acts
	// only on 1 (okay and enabled: the canonical revive); any other write only stores.
	"SotsbSelStatus": {
		toGame: func(e *Engine, canonical int64) int64 {
			if canonical == 1 {
				return adventure.SotsbStatus[rules.StatusOkay]
			}
			return adventure.SotsbStatus[e.ruleset().Sheet(e.selected()).Status] // a status with no row reads 0, as the Pascal's else
		},
	},
	// SpecialWriteZ1 $1B..$65: memorizedSpells[address-$1B], slot 0 included; POR's field starts at slot 1.
	"SotsbSelMemorized": {element: func(i int) int { return i - 1 }},
	"SotsbSymbolSet":    {}, // decoded only (adventure.Canonical)
}
