package engine

import (
	"os"
	"slices"
	"testing"
)

// TestBoundFieldsPOR: POR's layouts name every field the engine binds but the VM words POR
// keeps in the ruleset, and the shops' price factor (POR's shops have none).
func TestBoundFieldsPOR(t *testing.T) {
	got, err := UnboundFields(os.DirFS("../data/por/scripts"))
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{fieldRemainder, "FacingOrRaw", fieldLastBlock, fieldLoopCounter, fieldLoopLimit, "PartyCredits", fieldPriceScale, fieldScriptBlock, "SelCha", "SelCharClass", "SelCharisma", "SelClassLevels", "SelCompanion", "SelFlag1", "SelFlag2", "SelFlag3", "SelGems", "SelGender", "SelHearNoise", "SelHideShadows", "SelHpCur", "SelHpMax", "SelInt", "SelMageLevel", "SelMoveSilently", "SelOldLevels", "SelOpenLocks", "SelRangerLevel", "SelReadLanguages", "SelSideFlags", "SelStats", fieldTravelKind, fieldVault, fieldWildTerrain}; !slices.Equal(got, want) {
		t.Errorf("unbound %v, want %v", got, want)
	}
}
