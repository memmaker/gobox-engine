package present

import "testing"

// PoD's VGA icons recolour the six parts by index, as EGA does (R1C3).
func TestPoolIconRecoloursVGA(t *testing.T) {
	plain, ok := poolIcon("combat_head_026", "combat_body_131", nil, 0, false)
	if !ok || plain.Palette == nil {
		t.Skip("no VGA pool icon")
	}
	all := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff} // every part to index 15
	got, _ := poolIcon("combat_head_026", "combat_body_131", all, 0, false)
	for i, v := range plain.Pix {
		if plain.Palette[v].A == 0 {
			continue // transparent stays
		}
		part := v == 1 || v == 2 || v == 3 || v == 4 || v == 6 || v == 7 || v >= 9 && v <= 12 || v == 14 || v == 15
		if part && got.Pix[i] != 15 || !part && got.Pix[i] != v {
			t.Fatalf("pixel %d: index %d became %d", i, v, got.Pix[i])
		}
	}
}
