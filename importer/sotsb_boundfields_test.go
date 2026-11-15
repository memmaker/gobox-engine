package importer_test

import (
	"gobox/engine"
	"gobox/importer"
	"slices"
	"testing"
)

// TestBoundFields: sotsb's layout names every field pool binds (pool.BoundFields) with the
// canonical name, but these, which the game has no field for (sotsb-re: no script reads them,
// or the field is POR's own story or engine word).
func TestSotsbBoundFields(t *testing.T) {
	absent := []string{"AreaKind", "Boss", "EffectTable", "FacingOrRaw", "LoopCounter", "LoopLimit", "MapPatchA", "MapPatchB", "Mark3", "Mark4", "Mark4Off", "MarkFlags4", "MarkFlags57", "MonsterHit", "PartyCredits", "PartyHit", "PartyMove", "RiverOpen", "ScriptBlock", "SelCharClass", "SelCharisma", "SelClassLevels", "SelCompanion", "SelFlag1", "SelFlag2", "SelFlag3", "SelGender", "SelHearNoise", "SelHideShadows", "SelInt", "SelMageLevel", "SelMoveSilently", "SelOldLevels", "SelOpenLocks", "SelReadLanguages", "SelSideFlags", "SelStats", "SelWindowF2", "TourDone", "TravelKind", "UnknownD015", "VaultRequest", "ViewFlags"}
	got, err := engine.UnboundFields(importer.EmbeddedScripts("sotsb"))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, absent) {
		t.Errorf("unbound %q,\nlisted %q", got, absent)
	}
}
