package engine

import "testing"

// TestBR1TravelKindScreen: a script's write of a TravelKind other than the one held selects the screen (eclmenu.pas
// WriteGbvm, GBVM_AREA_KIND: the 3D view, the wilderness, the orbit map); the same value again changes nothing, so a
// scripted place (AreaKind 3) stays one. Block 32's load entry (TravelKind = 16, no AreaKind when ScriptBlock >= 32)
// relies on it: entered from a scripted place it was a place whose move entry never asked (sweep finding "loop without
// asking", replay 50 0 0 4 10701168).
func TestBR1TravelKindScreen(t *testing.T) {
	e := br1Begun(t)
	br1OwnBlock(t, e, 201, "    TravelKind = 16\n    AreaKind(3, 255)\n    TravelKind = 16") // the same value: stays scripted
	if e.world.Scripted != 1 || e.world.In3D != 1 {
		t.Fatalf("same value: scripted %d In3D %d, want a scripted place", e.world.Scripted, e.world.In3D)
	}
	br1OwnBlock(t, e, 202, "    TravelKind = 16\n    AreaKind(3, 255)\n    TravelKind = 17")
	if e.world.Scripted != 0 || e.world.In3D != 1 {
		t.Fatalf("new value: scripted %d In3D %d, want the 3D view", e.world.Scripted, e.world.In3D)
	}
	br1OwnBlock(t, e, 203, "    TravelKind = 162")
	if e.world.Scripted != 0 || e.world.In3D != 0 || e.world.Flying != 0 {
		t.Fatalf("wilderness: scripted %d In3D %d flying %d, want the overland", e.world.Scripted, e.world.In3D, e.world.Flying)
	}
}
