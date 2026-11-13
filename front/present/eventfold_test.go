package present

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// playedEvents is es as the old player gets them (view-cut: removed by slice 2).
func playedEvents(es []play.Report, fight bool, posed func(play.Report) (*play.Combat, []game.Member), holdMs int) []front.CombatEvent {
	if posed == nil {
		posed = func(play.Report) (*play.Combat, []game.Member) { return nil, nil }
	}
	g := NewGame(&fake{st: play.Exploring{}}, game.Data{})
	return g.legacyEvents(g.moments(es, fight, posed, holdMs), front.View{})
}

// Class 20b: play.Report → front kinds one to one; an EvAttack swing stays a
// swing (never a miss) and EvHit + EvDamage stay one hit and one damage.
func TestEventsFold(t *testing.T) {
	es := []play.Report{{Kind: game.EvAttack, Unit: 1, Target: 2}, {Kind: game.EvHit, Unit: 1, Target: 2},
		{Kind: game.EvDamage, Unit: 2, Target: 2, N: 5}, {Kind: game.EvAttack, Unit: 2, Target: 1}, {Kind: game.EvMiss, Unit: 2, Target: 1}}
	want := []front.EventKind{front.EvAttack, front.EvHit, front.EvDamage, front.EvAttack, front.EvMiss}
	out := playedEvents(es, false, nil, 0)
	if len(out) != len(want) {
		t.Fatalf("%d events, want %d", len(out), len(want))
	}
	for i, k := range want {
		if out[i].Kind != k {
			t.Errorf("event %d: %v, want %v", i, out[i].Kind, k)
		}
	}
	if out[2].Amount != 5 || out[2].Unit != 2 {
		t.Errorf("damage %+v", out[2])
	}
}
