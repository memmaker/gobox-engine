package engine

import (
	"gobox/adventure"
	"gobox/rules"
	"testing"
)

// TestPodCodecs round-trips pod's encodings on the original's values.
func TestPodCodecs(t *testing.T) {
	e := &Engine{}
	seen := map[int64]bool{}
	for _, n := range adventure.PodStatus { // the read is the member's own status: one C_Status number each
		if seen[n] {
			t.Errorf("C_Status %d twice", n)
		}
		seen[n] = true
	}
	sd := codecs["PodSelSide"]
	for _, g := range []int64{0, 1} {
		if got := sd.toGame(e, adventure.Canonical("PodSelSide", g, e.set)); got != g {
			t.Errorf("side %d -> %d", g, got)
		}
	}
	coin := func(i int) rules.Fact { f, _ := selCoin(i); return f }
	if c := codecs["PodSelCoins"].element; coin(c(0)) != rules.Fact(rules.CoinFact{Coin: rules.CoinGold}) || coin(c(2)) != rules.Fact(rules.PiecesFact{}) || coin(c(4)) != rules.Fact(rules.PiecesFact{Jewel: true}) || c(1)%2 != 1 {
		t.Error("coins: gold (the import's), gems, jewelry")
	}
	if codecs["PodSelMemorized"].element(0) != -1 {
		t.Error("memorized: slot 0")
	}
	for g, c := range map[int64]int64{0: 0, 3: 3, 4: 0, 9: 0} {
		if got := adventure.Canonical("PodSymbolSet", g, e.set); got != c {
			t.Errorf("symbol set %d: %d, want %d", g, got, c)
		}
	}
}
