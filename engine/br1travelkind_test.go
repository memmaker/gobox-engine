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

// TestBR1NewECLNotesTheBlockLeft: NewECL saves the block it leaves in ScriptBlock before the new block's load entry runs
// (eclvm.pas EclNewEcl: zone0's savedEclId := eclId), whether or not the block's move entry ever ran. Block 64's load entry
// calls NewECL(65) from its talk scene, and 65's `if ScriptBlock == 65 { exit }` (a loaded game) must not take it: it did,
// and the scripted place it came from stood on (sweep finding "loop without asking", replay 64/2/0,0/0).
func TestBR1NewECLNotesTheBlockLeft(t *testing.T) {
	e := br1Begun(t)
	br1OwnBlock(t, e, 202, "    NewGameBytes = ScriptBlock") // the block's own file for the next call to enter
	br1OwnBlock(t, e, 201, "    NewECL(202)")                // 201's load entry leaves for 202 before any move entry ran
	if got := br1Cell(t, e, "NewGameBytes"); got != 201 {
		t.Fatalf("ScriptBlock in block 202's load entry is %d, want the block left, 201", got)
	}
}
