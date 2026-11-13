package engine

import (
	"gobox/adventure"
	"testing"
)

// TestSotsbCodecs: sotsb's codecs on the original's values (sotsb-re eclmenu.pas $1B, $100, $312).
func TestSotsbCodecs(t *testing.T) {
	st := sotsbCodecs["SotsbSelStatus"]
	if got := st.toGame(nil, 1); got != 1 {
		t.Errorf("SelStatus okay reads %d, want 1", got)
	}
	for v, want := range map[int64]int64{1: adventure.CanonStatusOkay, 0: adventure.CanonStatusStore, 2: adventure.CanonStatusStore, 7: adventure.CanonStatusStore, 0x80: adventure.CanonStatusStore, 0x87: adventure.CanonStatusStore} {
		if got := adventure.Canonical("SotsbSelStatus", v, 0); got != want {
			t.Errorf("SelStatus write %#x is canonical %#x, want %#x (1 revives, the rest only store)", v, got, want)
		}
	}
	if got := sotsbCodecs["SotsbSelMemorized"].element(0); got != -1 { // slot 0 = canonical slot 0 (element -1 + 1)
		t.Errorf("memorized slot 0 is element %d", got)
	}
	for v, disk := range map[int64]int64{1: 1, 2: 1, 3: 2, 4: 2, 5: 3} {
		if got := adventure.Canonical("SotsbSymbolSet", v, 0); got != disk {
			t.Errorf("PushState(%d) keeps disk %d, want %d", v, got, disk)
		}
	}
}
