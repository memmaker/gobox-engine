package engine

import (
	"gobox/ecl/script"
	"gobox/rules"
)

// sotsbCodecs are sotsb's encodings of canonical fields (its layout's enc=).
var sotsbCodecs = map[script.Encoding]codec{
	// SpecialReadZ1 $100 (sotsb-re eclmenu.pas) reads the status 1..7; SpecialWriteZ1 $100 acts
	// only on 1 (okay and enabled: the canonical revive); any other write only stores.
	"SotsbSelStatus": {
		toGame: func(e *Engine, canonical int64) int64 {
			if canonical == 1 {
				return sotsbStatus[rules.StatusOkay]
			}
			return sotsbStatus[e.ruleset().Sheet(e.selected()).Status] // a status with no row reads 0, as the Pascal's else
		},
		toCanon: func(_ *Engine, g int64) int64 {
			if g == sotsbStatus[rules.StatusOkay] {
				return canonStatusOkay
			}
			return canonStatusStore
		},
	},
	// SpecialWriteZ1 $1B..$65: memorizedSpells[address-$1B], slot 0 included; POR's field starts at slot 1.
	"SotsbSelMemorized": {element: func(i int) int { return i - 1 }},
	// SpecialWriteZ1 $312: DaxIo.PushState(v) keeps disk (v+1) div 2, which $312 reads back.
	"SotsbSymbolSet": {toCanon: func(_ *Engine, v int64) int64 { return (v + 1) / 2 }},
}

// sotsbStatus is SpecialReadZ1 $100's number for each status.
var sotsbStatus = map[rules.Status]int64{
	rules.StatusOkay: 1, rules.StatusGone: 2, rules.StatusDead: 3, rules.StatusDying: 4,
	rules.StatusUnconscious: 5, rules.StatusRunning: 6, rules.StatusStoned: 7,
}
