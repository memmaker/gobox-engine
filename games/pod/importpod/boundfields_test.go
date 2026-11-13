package importpod

import (
	"gobox/engine"
	"slices"
	"testing"
)

// TestBoundFields: pod's layout names every field pool binds (pool.BoundFields) with the
// canonical name, but these, which the game has no field for (pod-re: no script reads them,
// or the field is POR's own story or engine word).
func TestBoundFields(t *testing.T) {
	absent := []string{"AreaKind", "Boss", "CombatEnd", "EffectTable", "MapPatchA", "MapPatchB", "Mark3", "Mark4", "Mark4Off", "MarkFlags4", "MarkFlags57", "Morale", "Portrait", "ProbeCell", "ProbeWall", "ProbeX", "ProbeY", "RiverOpen", "Scratch6", "SelExpPerHP", "SelPartyIndex", "SelPartyIndex2", "SelReaction", "SetGroundShade", "SetSkyShade", "Shop", "Sky", "SkyInside", "Sound", "TourDone", "UnknownD015", "UnusedB8", "UnusedB9", "ViewFlags", "WallSlot1", "WallSlot2", "WallSlot3", "WallSplitA", "WallSplitB"}
	got, err := engine.UnboundFields(scriptData())
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, absent) {
		t.Errorf("unbound %q,\nlisted %q", got, absent)
	}
}
