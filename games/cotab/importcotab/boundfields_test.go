package importcotab

import (
	"gobox/pool"
	"slices"
	"testing"
)

// TestBoundFields: cotab's layout names every field pool binds (pool.BoundFields) with the
// canonical name, but these, which the game has no field for (cotab-re: no script reads them,
// or the field is POR's own story or engine word).
func TestBoundFields(t *testing.T) {
	absent := []string{"AreaKind", "Boss", "EffectTable", "MapPatchA", "MapPatchB", "Mark3", "Mark4", "Mark4Off", "MarkFlags4", "MarkFlags57", "RiverOpen", "TourDone", "UnknownD015"}
	got, err := pool.UnboundFields(scriptData())
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, absent) {
		t.Errorf("unbound %q,\nlisted %q", got, absent)
	}
}
