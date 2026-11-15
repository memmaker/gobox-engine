package engine

import "gobox/ecl"

// cotabScriptFlagRead is what SpecialReadZ1 $E4 keeps of scriptFlags[0]: bit 0, imported from Hillsfar.
const cotabScriptFlagRead = 1

// cotabCodecs are cotab's encodings of canonical fields (its layout's enc=).
var cotabCodecs = map[ecl.Encoding]codec{
	// SelFlag1: the read keeps bit 0; SpecialWriteZ1 has no case for $E4.
	"CotabSelFlag1": {toGame: func(_ *Engine, v int64) int64 { return v & cotabScriptFlagRead }, ignoreWrites: true},
	// SelStatus: as POR's, but the read after a SELECT MEMBER that matched nobody is 0, once
	// (cotab-re eclvm.pas characterNotFound, eclmenu.pas SpecialReadZ1 $100).
	"CotabSelStatus": {toGame: func(e *Engine, v int64) int64 {
		if e.es.memberMissing {
			e.es.memberMissing = false
			return 0
		}
		return v
	}},
}
