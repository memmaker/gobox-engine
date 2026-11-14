package engine

import (
	"testing"

	"gobox/game"
	"gobox/rules1e"
)

// TestPickSpell: Cast out of combat is refused with gobox's spell pick
// among the member's memorized ones, each once; a wrong one is refused
// with it again; Cure Light Wounds then with the pick of whom it goes on.
func TestPickSpell(t *testing.T) {
	e := quickStarted(t)
	c := e.member(e.selected())
	c.Memorized[0], c.Memorized[1] = rules1e.SpCureLightWounds, rules1e.SpCureLightWounds
	call := func(spell int) *game.Ask {
		t.Helper()
		err := e.runCall(pendingCall{kind: callPower, member: -1, power: game.PowerCast, spell: spell, target: -1})
		r, ok := err.(*game.Refusal)
		if !ok || r.Ask == nil {
			t.Fatalf("spell %d: %v, want a pick", spell, err)
		}
		return r.Ask
	}
	if a := call(0); a.Kind != game.PickSpell || len(a.Options) != 1 || a.Options[0].Do.(game.UsePower).Spell != rules1e.SpCureLightWounds {
		t.Fatalf("got %+v, want Cure Light Wounds to pick", a)
	}
	if a := call(99); a.Kind != game.PickSpell {
		t.Errorf("a wrong spell: %+v", a)
	}
	if a := call(rules1e.SpCureLightWounds); a.Kind != game.PickMember {
		t.Fatalf("no target asked: %+v", a)
	}
}
