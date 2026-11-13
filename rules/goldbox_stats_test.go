package rules

import "testing"

// Pinned on the games' tables: 18/75 strength is index 20, hits +2, damages
// +3; dexterity 16 gives AC 2 and missile 1; con 12 saves +3 (low), 19 +1.
func TestStatsPinned(t *testing.T) {
	v := StrIndex(18, 75)
	if v != 20 || StrHit(v) != 2 || StrDamage(v) != 3 || StrIndex(19, 0) != 24 || StrIndex(18, 101) != 0 {
		t.Errorf("str: index %d hit %d dmg %d", v, StrHit(v), StrDamage(v))
	}
	if DexAC(16) != 2 || DexMissile(16) != 1 || DexAC(2) != -4 || DexMissile(1) != -4 {
		t.Error("dex")
	}
	if ConSaveLow(12) != 3 || ConSaveHigh(19) != 1 || ConSaveHigh(12) != 0 {
		t.Error("con saves")
	}
	if HitDiceWeight(false, 0, 2) != 2 || HitDiceWeight(true, 3, 9) != 2 || HitDiceWeight(false, 0, 9) != 8 {
		t.Error("hit dice weight")
	}
}
