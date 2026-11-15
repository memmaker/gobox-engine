package engine

import (
	"gobox/ecl"
	"gobox/rules"
)

// Status values POR and COTAB read (SpecialReadZ1 $100).
const (
	porStatusActive = 1
	porStatusOut    = 0x80
)

// porSelStatusRead is POR's and COTAB's SpecialReadZ1 $100 (por-re, cotab-re eclmenu.pas): 1 if enabled <> 0, else $80.
func porSelStatusRead(e *Engine, _ int64) int64 {
	if e.selValue(rules.FactActive).(rules.Yes) {
		return porStatusActive
	}
	return porStatusOut
}

// porCodecs are POR's encodings of canonical fields (its layout's enc=).
var porCodecs = map[ecl.Encoding]codec{
	"PorSelStatus": {toGame: porSelStatusRead},
}
