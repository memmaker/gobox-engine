package adventure

import (
	"testing"

	"gobox/rules"
)

// TestBRTravelKind: BR's areaKind byte maps to traits and back: every byte's traits re-encode to a byte with the
// same traits (the stored byte itself is kept whole by the field), and the values BR tests read as named.
func TestBRTravelKind(t *testing.T) {
	for b := 0; b < 256; b++ {
		want := BRPlaceTraits(byte(b))
		if got := BRPlaceTraits(BRAreaKind(want)); got != want {
			t.Errorf("byte %#x: traits %b re-encode to %#x with traits %b", b, want, BRAreaKind(want), got)
		}
		if got := Canonical("BrTravelKind", int64(b), 0); got != int64(want) {
			t.Errorf("Canonical(%#x) = %d, want %d", b, got, want)
		}
	}
	if BRPlaceTraits(0) != 0 {
		t.Error("kind 0 has traits")
	}
	if got := BRPlaceTraits(BRMarsWilderness); !got.Has(rules.TraitWilderness | rules.TraitTerrainShift) {
		t.Errorf("Mars %b", got)
	}
	if got := BRPlaceTraits(BRVenusWilderness); !got.Has(rules.TraitWilderness) || got.Has(rules.TraitTerrainShift) {
		t.Errorf("Venus %b", got)
	}
	if got := BRPlaceTraits(0x10 | 1); !got.Has(rules.TraitZeroGravity) || got.Has(rules.TraitWilderness) {
		t.Errorf("zero gravity %b", got)
	}
	if got := BRPlaceTraits(BRLotekB); !got.Has(rules.TraitLowTech) || got.Has(rules.TraitSecondTileSet) {
		t.Errorf("lotek $82 %b", got)
	}
}
