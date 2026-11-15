package engine

import (
	"slices"
	"testing"

	"gobox/adventure"
	"gobox/rules"
)

// TestSotsbCodecs: sotsb's codecs on the original's values (sotsb-re eclmenu.pas $1B, $100, $312).
func TestSotsbCodecs(t *testing.T) {
	st := sotsbCodecs["SotsbSelStatus"]
	e, _ := selStatusEngine(selStatusState{})
	if got := st.toGame(e, int64(slices.Index(adventure.StatusKinds, rules.StatusOkay))); got != 1 {
		t.Errorf("SelStatus okay reads %d, want 1", got)
	}
	for _, v := range []int64{0, 1, 2, 7, 0x80, 0x87} { // 1 sets okay, the rest are ignored
		w := adventure.DecodeStatusWrite(adventure.Canonical("SotsbSelStatus", v, 0))
		want := adventure.StatusWrite{Kind: adventure.StatusWriteNothing}
		if v == 1 {
			want = adventure.StatusWrite{Kind: adventure.StatusWriteSet, Status: rules.StatusOkay}
		}
		if w != want {
			t.Errorf("SelStatus write %#x is %+v, want %+v", v, w, want)
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
