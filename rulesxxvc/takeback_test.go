package rulesxxvc

import (
	"testing"

	"gobox/rules"
)

// TestTakeBack: combatai.pas CheckKeystroke's Space clears the quick flag of every team member in play that is no
// companion, hands the turn back when the actor was one of them, and leaves a companion and a monster alone.
func TestTakeBack(t *testing.T) {
	r := New()
	a, b, comp := r.Add(member("a", 10, 0, 0, 0)), r.Add(member("b", 10, 0, 0, 0)), r.Add(member("c", 10, 0, 0, 0))
	r.chars[comp].Companion = true
	for _, id := range []rules.CharID{a, b, comp} {
		r.chars[id].Side = rules.SideAuto
	}
	f := rules.Fight{Party: rules.Party{a, b, comp}}
	rng := rules.RNG{State: 1}
	r.BeginFight(f, &rng)
	r.BeginRound(f, &rng)
	f.Actor = a
	o, err := r.Resolve(f, rules.TakeBack{}, &rng)
	if err != nil {
		t.Fatal(err)
	}
	if r.chars[a].Side != rules.SideParty || r.chars[b].Side != rules.SideParty || r.chars[comp].Side != rules.SideAuto {
		t.Errorf("sides a %v b %v companion %v", r.chars[a].Side, r.chars[b].Side, r.chars[comp].Side)
	}
	if !o.TurnOver || r.fight[a].initiative != handBackInitiative {
		t.Errorf("the actor was not handed back: %+v initiative %d", o, r.fight[a].initiative)
	}
	f.Actor = comp // a companion acting stays on the computer: its turn goes on
	if o, err := r.Resolve(f, rules.TakeBack{}, &rng); err != nil || o.TurnOver {
		t.Errorf("the companion's turn: %+v %v", o, err)
	}
}
