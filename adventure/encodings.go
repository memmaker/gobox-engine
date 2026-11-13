package adventure

import (
	"fmt"

	"gobox/ecl"
	"gobox/rules"
)

// A game keeps some canonical script fields in an encoding of its own
// (ecl.Field.Encoding, its layout's enc=). Canonical decodes a game's value:
// the engine's script writes and the importer's original saves both go
// through it. The other way (a read) needs the running game: the engine's.

// The canonical SelStatus values (POR's): a read is Okay or Out.
const (
	CanonStatusDied   = 0    // a write: the member died
	CanonStatusOkay   = 1    // active; a write revives the member
	CanonStatusStore  = 2    // a write that only stores
	CanonStatusOut    = 0x80 // inactive (a write >= 0x80 but Stoned)
	CanonStatusStoned = 0x87 // a write: stoned
)

// ScriptSide is the Side word's number of each side (SpecialReadZ1 268);
// writing any other number changes nothing.
var ScriptSide = map[rules.Side]uint16{rules.SideParty: 0, rules.SideAuto: 0x80, rules.SideMonsters: 0x81}

// pod's values (pod-re C_Status, TCharRecord.side).
const (
	PodSideParty     = 0 // TCharRecord.side: the party's
	PodSideOther     = 1 // the other side
	podSymbolSetZero = 4 // TileSet 4 loads symbol set 0 (WriteGbvm $39)
	podSymbolSets    = 5 // TileSet writes from here on are ignored
)

// PodStatus is C_Status's number for each status (pod-re includes/c_status.pas; 2 has no name).
var PodStatus = map[rules.Status]int64{
	rules.StatusOkay: 0, rules.StatusAnimated: 1, rules.StatusRunning: 3, rules.StatusUnconscious: 4,
	rules.StatusDying: 5, rules.StatusDead: 6, rules.StatusStoned: 7, rules.StatusGone: 8,
}

// SotsbStatus is sotsb's SpecialReadZ1 $100 number for each status.
var SotsbStatus = map[rules.Status]int64{
	rules.StatusOkay: 1, rules.StatusGone: 2, rules.StatusDead: 3, rules.StatusDying: 4,
	rules.StatusUnconscious: 5, rules.StatusRunning: 6, rules.StatusStoned: 7,
}

// toCanon is each encoding's decoding that changes the value; set is the
// symbol set in play.
var toCanon = map[ecl.Encoding]func(game int64, set rules.Set) int64{
	// TCharRecord.status, 0..8 (OKAY 0); SpecialWriteZ1 $1ED: inPlay = status = OKAY.
	"PodSelStatus": func(g int64, _ rules.Set) int64 {
		switch g {
		case PodStatus[rules.StatusOkay]:
			return CanonStatusOkay // the revive
		case PodStatus[rules.StatusDead]:
			return CanonStatusDied
		case PodStatus[rules.StatusStoned]:
			return CanonStatusStoned
		}
		return CanonStatusOut
	},
	// TileSet (WriteGbvm $39): 0..3 the set, 4 set 0, 5 and up no change.
	"PodSymbolSet": func(g int64, set rules.Set) int64 {
		switch {
		case g == podSymbolSetZero:
			return 0
		case g >= podSymbolSets:
			return int64(set)
		}
		return g
	},
	// TCharRecord.side: 0 the party's, 1 the other; the canonical 0 party, 0x81 monsters.
	"PodSelSide": func(g int64, _ rules.Set) int64 {
		switch g {
		case PodSideParty:
			return int64(ScriptSide[rules.SideParty])
		case PodSideOther:
			return int64(ScriptSide[rules.SideMonsters])
		}
		panic(fmt.Sprintf("pod: SelSide %d", g))
	},
	// SpecialWriteZ1 $100 acts only on 1 (okay and enabled: the canonical revive); any other write only stores.
	"SotsbSelStatus": func(g int64, _ rules.Set) int64 {
		if g == SotsbStatus[rules.StatusOkay] {
			return CanonStatusOkay
		}
		return CanonStatusStore
	},
	// SpecialWriteZ1 $312: DaxIo.PushState(v) keeps disk (v+1) div 2, which $312 reads back.
	"SotsbSymbolSet": func(v int64, _ rules.Set) int64 { return (v + 1) / 2 },
}

// Canonical is the value v of a field in encoding enc as the canonical one,
// with set the symbol set in play.
func Canonical(enc ecl.Encoding, v int64, set rules.Set) int64 {
	if f, ok := toCanon[enc]; ok {
		return f(v, set)
	}
	return v
}
