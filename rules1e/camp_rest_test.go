package rules1e

import (
	"gobox/core"
	"testing"
)

// TestElapseRest: a rest's Elapse heals the party every 0x120 steps of five
// minutes and stops at the step a monster interrupts.
func TestElapseRest(t *testing.T) {
	r := &Rules{S: &RuleState{}}
	r.S.Camp.Heal = 0x120 - 2
	o := r.Elapse(nil, 15, nil)
	if len(o.Effects) != 1 || o.Effects[0] != (core.PartyHealed{}) || r.S.Camp.Heal != 1 {
		t.Errorf("heal: %+v, Heal %#x", o.Effects, r.S.Camp.Heal)
	}
	r.S.Game.RestEvery, r.S.Game.RestChance = 2, 100
	o = r.Elapse(nil, 60, nil)
	if len(o.Effects) != 1 || o.Effects[0] != (core.RestInterrupted{}) || r.S.Camp.Heal != 3 {
		t.Errorf("interrupt after two steps: %+v, Heal %#x", o.Effects, r.S.Camp.Heal)
	}
}
