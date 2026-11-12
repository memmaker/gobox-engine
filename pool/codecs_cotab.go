package pool

import "gobox/ecl/script"

// cotabCodecs are cotab's encodings of canonical fields (its layout's enc=).
var cotabCodecs = map[script.Encoding]codec{
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
