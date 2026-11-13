package engine

import "testing"

// TestSotsbCodecs: sotsb's codecs on the original's values (sotsb-re eclmenu.pas $1B, $100, $312).
func TestSotsbCodecs(t *testing.T) {
	st := sotsbCodecs["SotsbSelStatus"]
	if got := st.toGame(nil, 1); got != 1 {
		t.Errorf("SelStatus okay reads %d, want 1", got)
	}
	for v, want := range map[int64]int64{1: canonStatusOkay, 0: canonStatusStore, 2: canonStatusStore, 7: canonStatusStore, 0x80: canonStatusStore, 0x87: canonStatusStore} {
		if got := st.toCanon(nil, v); got != want {
			t.Errorf("SelStatus write %#x is canonical %#x, want %#x (1 revives, the rest only store)", v, got, want)
		}
	}
	if got := sotsbCodecs["SotsbSelMemorized"].element(0); got != -1 { // slot 0 = canonical slot 0 (element -1 + 1)
		t.Errorf("memorized slot 0 is element %d", got)
	}
	ss := sotsbCodecs["SotsbSymbolSet"]
	for v, disk := range map[int64]int64{1: 1, 2: 1, 3: 2, 4: 2, 5: 3} {
		if got := ss.toCanon(nil, v); got != disk {
			t.Errorf("PushState(%d) keeps disk %d, want %d", v, got, disk)
		}
	}
}
