package engine

import "testing"

// TestBR1ItemSlots ports rules TestItemSlots (the quick team's readied kit names its doll slot, so a front can place
// it): every readied weapon or armour of the party says so as its Slot, as its tab; an unreadied one has none.
func TestBR1ItemSlots(t *testing.T) { brItemSlots(t, gameBR1) }

func TestBR2ItemSlots(t *testing.T) { brItemSlots(t, gameBR2) }

func brItemSlots(t *testing.T, g brGame) {
	e := brBegun(t, g)
	seen := map[string]bool{}
	for _, m := range commonOf(e.build()).Party {
		for _, it := range m.Items {
			switch {
			case it.Readied && (it.Kind == "weapon" || it.Kind == "armor"):
				seen[it.Kind] = true
				if it.Slot != it.Kind {
					t.Errorf("%s: readied %s in slot %q", m.Name, it.Name, it.Slot)
				}
			case !it.Readied && it.Slot != "":
				t.Errorf("%s: unreadied %s in slot %q", m.Name, it.Name, it.Slot)
			}
		}
	}
	if !seen["weapon"] || !seen["armor"] {
		t.Errorf("the party readies %v: want a weapon and armour", seen)
	}
}
