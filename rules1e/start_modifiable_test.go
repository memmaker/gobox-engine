package rules1e

import "testing"

// TestModifiable: only the start's own experience for the class count.
func TestModifiable(t *testing.T) {
	s := &RuleState{}
	tb := POR
	tb.Start.Experience = []int32{25000, 12500}
	s.Tables = &tb
	c := &Character{}
	c.Levels[ClsFighter] = 1
	for _, k := range []struct {
		exp  uint32
		want bool
	}{{25000, true}, {0, false}, {25001, false}} {
		c.Exp = k.exp
		if got := s.Modifiable(c); got != k.want {
			t.Errorf("exp %d: %v", k.exp, got)
		}
	}
	c.Levels[ClsThief], c.Exp = 1, 12500
	if !s.Modifiable(c) {
		t.Error("two classes at 12500")
	}
}
