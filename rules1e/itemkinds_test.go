package rules1e

import (
	"testing"

	"gobox/core/dnd"
)

// TestItemKinds: an item's kind is the game's numbering, never POR's
// (SotSB's 0x24 is armour, its Long Sword 18).
func TestItemKinds(t *testing.T) {
	tb := POR
	e := &RuleState{Tables: &tb}
	if !e.isKind(ItemLongSword, dnd.ItemLongSword) || e.typeOf(dnd.ItemLongSword) != ItemLongSword {
		t.Fatal("POR's long sword")
	}
	tb.Kinds = SotsbItemKinds
	if e.isKind(0x24, dnd.ItemLongSword) || e.typeOf(dnd.ItemLongSword) != 18 || !e.isBow(26) || e.typeOf(dnd.ItemArrow) != 30 {
		t.Fatal("SotSB's numbering")
	}
}

// TestGroundIDs: every game's grounds fit its table (SotSB's 31 records).
func TestGroundIDs(t *testing.T) {
	for name, g := range map[string]GroundIDs{"por": PorGroundIDs, "cotab": CotabGroundIDs, "sotsb": SotsbGroundIDs, "pod": PodGroundIDs} {
		for _, v := range []Ground{g.Floor, g.PoisonCloud, g.StinkingCloud, g.Body} {
			if v == 0 || name == "sotsb" && v > 30 {
				t.Errorf("%s ground %#x", name, v)
			}
		}
	}
}
