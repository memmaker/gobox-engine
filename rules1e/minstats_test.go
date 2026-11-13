package rules1e

import "testing"

func TestParseMinStat(t *testing.T) {
	for _, c := range []struct {
		s   string
		str bool
		v   int
		ok  bool
	}{{"15", false, 15, true}, {"18/50", true, 1850, true}, {"18/00", true, 1900, true}, {"17", true, 1700, true}, {"18/50", false, 0, false}, {"x", false, 0, false}} {
		if v, ok := parseMinStat(c.s, c.str); v != c.v || ok != c.ok {
			t.Errorf("%q: %d %v", c.s, v, ok)
		}
	}
	if StrScore(18, 100) <= StrScore(18, 99) || StrScore(18, 1) <= StrScore(18, 0) {
		t.Error("18/xx ordering")
	}
}
