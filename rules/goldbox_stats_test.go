package rules

import "testing"

// Pinned on the games' tables: dexterity 16 gives AC 2 and missile 1; con 12 saves +3 (low), 19 +1.
func TestStatsPinned(t *testing.T) {
	if DexAC(16) != 2 || DexMissile(16) != 1 || DexAC(2) != -4 || DexMissile(1) != -4 {
		t.Error("dex")
	}
	if ConSaveLow(12) != 3 || ConSaveHigh(19) != 1 || ConSaveHigh(12) != 0 {
		t.Error("con saves")
	}
}
