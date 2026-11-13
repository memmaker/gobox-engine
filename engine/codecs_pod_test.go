package engine

import (
	"gobox/core"
	"gobox/rules1e"
	"testing"
)

// TestPodCodecs round-trips pod's encodings on the original's values.
func TestPodCodecs(t *testing.T) {
	e := &Engine{}
	st := codecs["PodSelStatus"]
	for g, c := range map[int64]int64{0: 1, 6: 0, 7: 0x87, 4: 0x80, 8: 0x80} {
		if got := st.toCanon(e, g); got != c {
			t.Errorf("status %d: canon %#x, want %#x", g, got, c)
		}
	}
	seen := map[int64]bool{}
	for _, n := range podStatus { // the read is the member's own status: one C_Status number each
		if seen[n] {
			t.Errorf("C_Status %d twice", n)
		}
		seen[n] = true
	}
	sd := codecs["PodSelSide"]
	for _, g := range []int64{0, 1} {
		if got := sd.toGame(e, sd.toCanon(e, g)); got != g {
			t.Errorf("side %d -> %d", g, got)
		}
	}
	coin := func(i int) core.Fact { f, _ := selCoin(i); return f }
	if c := codecs["PodSelCoins"].element; coin(c(0)) != core.Fact(core.CoinFact{Coin: rules1e.Gold}) || coin(c(2)) != core.Fact(core.PiecesFact{}) || coin(c(4)) != core.Fact(core.PiecesFact{Jewel: true}) || c(1)%2 != 1 {
		t.Error("coins: gold (the import's), gems, jewelry")
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
