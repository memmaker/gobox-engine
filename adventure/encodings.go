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

// PodStatus is C_Status's number for each status (pod-re includes/c_status.pas; 2 has no name, 9 is the reserve's, CharFightsAlone).
var PodStatus = map[rules.Status]int64{
	rules.StatusOkay: 0, rules.StatusAnimated: 1, rules.StatusRunning: 3, rules.StatusUnconscious: 4,
	rules.StatusDying: 5, rules.StatusDead: 6, rules.StatusStoned: 7, rules.StatusGone: 8, rules.StatusReserve: 9,
}

// SotsbStatus is sotsb's SpecialReadZ1 $100 number for each status.
var SotsbStatus = map[rules.Status]int64{
	rules.StatusOkay: 1, rules.StatusGone: 2, rules.StatusDead: 3, rules.StatusDying: 4,
	rules.StatusUnconscious: 5, rules.StatusRunning: 6, rules.StatusStoned: 7,
}

// StatusKinds is the canonical SelStatus read: a status is its place in this list (the
// engine's binding turns it back into the typed rules.Status at once); a game's
// encoding maps its own numbers to and from it.
var StatusKinds = []rules.Status{rules.StatusOkay, rules.StatusAnimated, rules.StatusRunning, rules.StatusUnconscious,
	rules.StatusDying, rules.StatusDead, rules.StatusStoned, rules.StatusGone, rules.StatusTempGone, rules.StatusReserve,
	rules.StatusPoisoned, rules.StatusComatose, rules.StatusNotHere}

// BRStatus is types.pas TStatus's number for each rules status BR has one for (StNoStatus 0 and the fight's own
// states 10..14 have none: no BR1 script reads or writes them, 0 reads while no character is selected).
var BRStatus = map[rules.Status]int64{
	rules.StatusOkay: 1, rules.StatusGone: 2, rules.StatusDead: 3, rules.StatusDying: 4,
	rules.StatusUnconscious: 5, rules.StatusRunning: 6, rules.StatusPoisoned: 7, rules.StatusComatose: 8, rules.StatusNotHere: 9,
}

// BRStatusMask is what setChar keeps of a status write (ecl.go: r[212] = v & 127).
const BRStatusMask = 127

// BRCompanionRead is what a read of byte $4C answers while isCompanion is non-zero
// (types.pas: SpecialReadZ1 answers $80); ScriptCompanion is the bit $40 of the side word.
const (
	BRCompanionRead = 0x80
	ScriptCompanion = 0x40
)

// SexOrder is a sex's number for the scripts: its place here (types.pas TGender: male 0, female 1).
var SexOrder = []rules.Sex{rules.Male, rules.Female}

// toCanon is each encoding's decoding that changes the value; set is the
// symbol set in play.
var toCanon = map[ecl.Encoding]func(game int64, set rules.Set) int64{
	// TCharRecord.status, 0..9 (OKAY 0); SpecialWriteZ1 $1ED.
	"PodSelStatus": func(g int64, _ rules.Set) int64 { return CanonStatusWrite(podStatusWrite(g)) },
	// POR and COTAB: SpecialWriteZ1 $100.
	"PorSelStatus":   func(g int64, _ rules.Set) int64 { return CanonStatusWrite(porStatusWrite(g)) },
	"CotabSelStatus": func(g int64, _ rules.Set) int64 { return CanonStatusWrite(porStatusWrite(g)) },
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
	// zone0 areaKind (types.pas): the canonical value is rules.PlaceTraits.
	"BrTravelKind":   func(g int64, _ rules.Set) int64 { return int64(BRPlaceTraits(byte(g))) },
	"SotsbSelStatus": func(g int64, _ rules.Set) int64 { return CanonStatusWrite(sotsbStatusWrite(g)) },
	"BrSelStatus":    func(g int64, _ rules.Set) int64 { return CanonStatusWrite(brStatusWrite(g)) },
	// setChar: r[76] = v, read back as 128 while non-zero; canonical: 0 or 1.
	"BrSelCompanion": func(g int64, _ rules.Set) int64 {
		if g != 0 {
			return 1
		}
		return 0
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

// brPlaceTrait is one row of BR's areaKind byte: the trait and the bits that make it. THE table both directions
// are made from (types.pas areaKind; the tests are in rules.PlaceTraits' comments).
type brPlaceTrait struct {
	trait rules.PlaceTraits
	has   func(kind byte) bool
}

// The areaKind values and bits BR tests (c_area.pas, combatst.pas).
const (
	BRVenusWilderness = 0xA2
	BRMarsWilderness  = 0xA8
	BRZeroGravity     = 0x10
	BRLotekA          = 0x88
	BRLotekB          = 0x82
	BRScatterTiles    = 0x40
	BRSecondTileSet   = 0x08
	BRTerrainNibble   = 0x0F // the low nibble: 8 raises the wilderness terrain by 10
)

var brPlaceTraits = []brPlaceTrait{
	{rules.TraitHasKind, func(k byte) bool { return k != 0 }},
	{rules.TraitWilderness, func(k byte) bool { return k == BRVenusWilderness || k == BRMarsWilderness }},
	{rules.TraitZeroGravity, func(k byte) bool { return k&BRZeroGravity != 0 }},
	{rules.TraitLowTech, func(k byte) bool { return k == BRLotekA || k == BRLotekB }},
	{rules.TraitScatterTiles, func(k byte) bool { return k&BRScatterTiles != 0 }},
	{rules.TraitSecondTileSet, func(k byte) bool { return k&BRSecondTileSet != 0 }},
	{rules.TraitTerrainShift, func(k byte) bool { return k&BRTerrainNibble == BRSecondTileSet }},
}

// BRPlaceTraits is the traits of BR's areaKind byte.
func BRPlaceTraits(kind byte) rules.PlaceTraits {
	var t rules.PlaceTraits
	for _, r := range brPlaceTraits {
		if r.has(kind) {
			t |= r.trait
		}
	}
	return t
}

// BRAreaKind is a byte with exactly the traits t (the canonical value as BR keeps it). The script's own byte is
// stored whole and never rebuilt from the traits: the byte a script wrote reads back as written.
func BRAreaKind(t rules.PlaceTraits) byte {
	var k byte
	switch {
	case t.Has(rules.TraitWilderness) && t.Has(rules.TraitTerrainShift):
		k = BRMarsWilderness
	case t.Has(rules.TraitWilderness):
		k = BRVenusWilderness
	case t.Has(rules.TraitLowTech) && t.Has(rules.TraitSecondTileSet):
		k = BRLotekA
	case t.Has(rules.TraitLowTech):
		k = BRLotekB
	case t.Has(rules.TraitTerrainShift):
		k = BRSecondTileSet
	case t.Has(rules.TraitSecondTileSet):
		k = BRSecondTileSet | 1 // bit 8 with the nibble not 8
	case t.Has(rules.TraitHasKind):
		k = 1
	}
	if t.Has(rules.TraitZeroGravity) {
		k |= BRZeroGravity
	}
	if t.Has(rules.TraitScatterTiles) {
		k |= BRScatterTiles
	}
	return k
}
