package importer_test

import (
	"gobox/engine"
	"gobox/importer"
	"slices"
	"testing"
)

// TestBoundFields: pod's layout names every field pool binds (pool.BoundFields) with the
// canonical name, but these, which the game has no field for (pod-re: no script reads them,
// or the field is POR's own story or engine word).
func TestPodBoundFields(t *testing.T) {
	absent := []string{"AreaKind", "Boss", "CombatEnd", "EffectTable", "FacingOrRaw", "FightState", "FoeControl", "FoeCrew", "FoeEngines", "FoeHull", "FoeLifeSupport", "FoeSensors", "FoeWeaponCount", "LoopCounter", "LoopLimit", "MapPatchA", "MapPatchB", "Mark3", "Mark4", "Mark4Off", "MarkFlags4", "MarkFlags57", "Morale", "PartyCredits", "Portrait", "PriceMultiplier", "ProbeCell", "ProbeWall", "ProbeX", "ProbeY", "RiverOpen", "ScriptBlock", "SelCha", "SelCompanion", "SelExpPerHP", "SelFlag1", "SelFlag2", "SelFlag3", "SelGems", "SelInt", "SelMageLevel", "SelPartyIndex", "SelPartyIndex2", "SelRangerLevel", "SelReaction", "SelSideFlags", "SelSkills", "SelWindowF2", "SetGroundShade", "SetSkyShade", "ShipControl", "ShipEngines", "ShipFuel", "ShipHull", "ShipLifeSupport", "ShipMount0Ammo", "ShipMount0Reloads", "ShipMount0Status", "ShipMount1Ammo", "ShipMount1Reloads", "ShipMount1Status", "ShipMount2Status", "ShipSensors", "Shop", "Sky", "SkyInside", "Sound", "TourDone", "TravelKind", "UnknownD015", "UnusedB8", "UnusedB9", "ViewFlags", "WallSlot1", "WallSlot2", "WallSlot3", "WallSplitA", "WallSplitB", "WildTerrain"}
	got, err := engine.UnboundFields(importer.PodScriptData())
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, absent) {
		t.Errorf("unbound %q,\nlisted %q", got, absent)
	}
}
