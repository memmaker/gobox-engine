package present

import (
	"testing"

	"gobox/front"
)

// TestSlotFooterNoLoad: no hint key for an act not allowed: NoLoad drops
// "L=load".
func TestSlotFooterNoLoad(t *testing.T) {
	for _, w := range slotFooter(&front.Slots{CanSave: true, NoLoad: true}, 60) {
		if w.key == 'L' {
			t.Fatalf("L=load shown with NoLoad")
		}
	}
	if ws := slotFooter(&front.Slots{CanSave: true}, 60); ws[1].key != 'L' {
		t.Fatalf("L=load gone: %+v", ws)
	}
}
