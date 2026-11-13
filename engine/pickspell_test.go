package engine

import (
	"gobox/rules1e"
	"slices"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestPickSpell: Cast out of combat asks for a spell among the member's
// memorized ones, each once (the pending question); a wrong one asks
// again; Cure Light Wounds then asks whom it goes on.
func TestPickSpell(t *testing.T) {
	e := quickStarted(t)
	c := e.member(e.selected())
	c.Memorized[0], c.Memorized[1] = rules1e.SpCureLightWounds, rules1e.SpCureLightWounds
	call := func(spell int) *playQuestion {
		t.Helper()
		e.ext.pq = nil
		if err := e.runCall(pendingCall{kind: callPower, member: -1, power: game.PowerCast, spell: spell, target: -1}); err != nil {
			t.Fatal(err)
		}
		return e.ext.pq
	}
	if pq := call(0); pq == nil || pq.of != askSpell || len(pq.q.Replies) != 1 {
		t.Fatalf("got %+v, want one spell to pick", pq)
	}
	if pq := call(99); pq == nil || pq.of != askSpell {
		t.Errorf("a wrong spell: %+v", pq)
	}
	if pq := call(rules1e.SpCureLightWounds); pq == nil || pq.of != askMember || len(pq.q.Replies) != len(e.state().Party) {
		t.Fatalf("no target asked: %+v", pq)
	}
	if !slices.Contains(commonOf(e.build()).Allowed.Plain, play.Action(play.Reply{N: 0})) {
		t.Error("the target question's reply not allowed")
	}
}
