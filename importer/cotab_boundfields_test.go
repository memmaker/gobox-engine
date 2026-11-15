package importer_test

import (
	"gobox/engine"
	"gobox/importer"
	"slices"
	"testing"
)

// TestBoundFields: cotab's layout names every field pool binds (pool.BoundFields) with the
// canonical name, but these, which the game has no field for (cotab-re: no script reads them,
// or the field is POR's own story or engine word).
func TestCotabBoundFields(t *testing.T) {
	absent := []string{"AreaKind", "Boss", "EffectTable", "FacingOrRaw", "LoopCounter", "LoopLimit", "MapPatchA", "MapPatchB", "Mark3", "Mark4", "Mark4Off", "MarkFlags4", "MarkFlags57", "PartyCredits", "RiverOpen", "ScriptBlock", "SelCareer", "SelCha", "SelCharisma", "SelClassLevels", "SelCompanion", "SelFlag2", "SelFlag3", "SelGems", "SelHearNoise", "SelHideShadows", "SelHitPoints", "SelHitPointsMax", "SelHpCur", "SelHpMax", "SelMoveSilently", "SelOldLevels", "SelOpenLocks", "SelRangerLevel", "SelReadLanguages", "SelSex", "SelSideFlags", "SelStats", "SelStatusKind", "SelWindowF2", "TourDone", "TravelKind", "UnknownD015", "VaultRequest"}
	got, err := engine.UnboundFields(importer.EmbeddedScripts("cotab"))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, absent) {
		t.Errorf("unbound %q,\nlisted %q", got, absent)
	}
}
