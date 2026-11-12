package pool

import (
	"rules1e"
	"testing"
)

// A party member under the computer bandages a dying ally and ends its turn
// (CotAB ovr010 sub_35DB1).
func TestQuickBandage(t *testing.T) {
	e := quickStarted(t)
	a, b := e.state().Party[0], e.state().Party[1]
	for _, c := range e.state().Party {
		c.Combat = &rules1e.CombatInfo{}
		c.Side = rules1e.SideParty
	}
	a.Combat.Delay = 5
	b.Status = rules1e.StatusDying
	if !e.state().AiAttackLoop(a) {
		t.Fatal("the turn goes on after bandaging")
	}
	if b.Status != rules1e.StatusUnconscious || a.Combat.Delay != 0 {
		t.Fatalf("status %d, delay %d", b.Status, a.Combat.Delay)
	}
}
