package pool

import "testing"

// TestPodCodecs round-trips pod's encodings on the original's values.
func TestPodCodecs(t *testing.T) {
	e := &Engine{}
	st := codecs["PodSelStatus"]
	for g, c := range map[int64]int64{0: 1, 6: 0, 7: 0x87, 4: 0x80, 8: 0x80} {
		if got := st.toCanon(e, g); got != c {
			t.Errorf("status %d: canon %#x, want %#x", g, got, c)
		}
	}
	if st.toGame(e, 1) != 0 || st.toGame(e, 0x80) == 0 {
		t.Error("status read")
	}
	sd := codecs["PodSelSide"]
	for _, g := range []int64{0, 1} {
		if got := sd.toGame(e, sd.toCanon(e, g)); got != g {
			t.Errorf("side %d -> %d", g, got)
		}
	}
	if c := codecs["PodSelCoins"].element; c(0) != 8 || c(1) != 9 || c(2)%2 != 1 || c(4)%2 != 1 {
		t.Error("coins: platinum is 8, gems and jewelry no coin")
	}
	if codecs["PodSelMemorized"].element(0) != -1 {
		t.Error("memorized: slot 0")
	}
	ss := codecs["PodSymbolSet"]
	for g, c := range map[int64]int64{0: 0, 3: 3, 4: 0, 9: 0} {
		if got := ss.toCanon(e, g); got != c {
			t.Errorf("symbol set %d: %d, want %d", g, got, c)
		}
	}
}
