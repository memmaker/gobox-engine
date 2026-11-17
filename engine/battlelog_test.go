package engine

import (
	"slices"
	"strings"
	"testing"

	"gobox/game"
	"gobox/rules"
	"gobox/rules1e"
)

// TestLogEveryChange: the battle log tells every change of a fight: each
// blow (who, whom, with what, how much), each fall with the rules' word
// for it, and a change no effect names (here hit points healed in place,
// as an affect's handler does) from reconcile. Bug class: a state change
// with no log line.
func TestLogEveryChange(t *testing.T) {
	e := devFirstTurn(t, "combat", false)
	var foe *rules1e.Character // a monster of the dev fight
	for i := 1; i <= e.state().Ncomb; i++ {
		if c := e.state().CombatantSlots[i-1]; c != nil && c.Side != e.state().Party[0].Side {
			foe = c
		}
	}
	if foe == nil {
		t.Fatal("no foe")
	}
	foe.HP -= 2
	e.present(rules.Narrated{Text: "a word"}) // a step with nothing that names the change
	log := strings.Join(e.ext.log.lines, "|")
	if !strings.Contains(log, "loses 2 hp") {
		t.Fatalf("an unnamed change unlogged: %q", e.ext.log.lines)
	}
	if strings.Contains(log, "a word") {
		t.Fatalf("the rules' text in the ledger: %q", e.ext.log.lines)
	}
	for range 300 {
		s := e.build()
		if _, ok := s.(game.Fighting); !ok {
			break
		}
		if myTurn(s) {
			e.doNow(game.Quick{All: true})
		} else {
			var o game.Outcome
			e.stepOn(&o, nil)
		}
	}
	hits := slices.ContainsFunc(e.ext.log.lines, func(l string) bool { return strings.Contains(l, " hits ") && strings.Contains(l, " for ") })
	downs := slices.ContainsFunc(e.ext.log.lines, func(l string) bool { return strings.Contains(l, "goes down") || strings.Contains(l, "Surrenders") })
	if !hits || !downs {
		t.Fatalf("a fight's log without its blows or falls: %q", e.ext.log.lines)
	}
}
