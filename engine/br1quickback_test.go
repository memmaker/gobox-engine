package engine

import (
	"strings"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// br1PlayStepwise ends the party's turns through TurnChoice and paces the rest with Step, as the presenter does
// (Do would run the computer's turns on to the next player's turn): it stops when stop holds, after each state.
func br1PlayStepwise(t *testing.T, e *Engine, each func(game.State) (stop bool)) {
	t.Helper()
	for range br1FightCap {
		s := e.build()
		if _, ok := s.(game.Fighting); !ok {
			t.Fatalf("the fight ended: %T, log %q", s, e.ext.log.lines)
		}
		if each(s) {
			return
		}
		switch {
		case myTurn(s):
			if err := e.TurnChoice(TurnAct{Verb: VerbEnd}, false); err != nil {
				t.Fatal(err)
			}
		case e.waiting(PendFight):
			e.Step()
		default:
			press(e)
		}
	}
	t.Fatal("no stop")
}

// TestBR1QuickBack: a member on Quick is played by the computer; while the fight waits between turns the state
// offers game.Manual (the original's Space, combatai.pas CheckKeystroke); sent, the member is the player's again.
func TestBR1QuickBack(t *testing.T) {
	e, f := br1AtTurn(t, br1FarFight, br1FarFacing)
	quick := f.Field.Turn
	if err := e.TurnChoice(TurnAct{Verb: VerbQuick}, false); err != nil {
		t.Fatal(err)
	}
	offered, sent := 0, false
	br1PlayStepwise(t, e, func(s game.State) bool {
		if myTurn(s) && s.(game.Fighting).Field.Turn == quick {
			if !sent {
				t.Fatal("the quick member was asked before it was taken back")
			}
			return true // the member is asked again
		}
		if !myTurn(s) && e.waiting(PendFight) && playAllows(s, game.Manual{}) {
			offered++
			if !sent {
				if _, err := e.doNow(game.Manual{}); err != nil {
					t.Fatal(err)
				}
				sent = true
			}
		}
		return false
	})
	if offered == 0 || !sent {
		t.Error("Manual was never offered")
	}
}

// TestBR1FledAndDown: one member flees, the other only ends its turns until the monster downs it: the fight ends as
// a flight (postcbt.pas TEAM_FLED). The runner stays okay in the party, the downed member and the monster are gone
// from the list (the engine reads it through PartyFact, as for the monsters), nobody gets experience and no loot
// page comes up.
func TestBR1FledAndDown(t *testing.T) {
	e, f := br1AtTurn(t, br1FarFight, br1FarFacing)
	runner := f.Field.Turn
	runnerID := e.unitID(runner)
	exp := func(id rules.CharID) int64 { return e.ruleset().Sheet(id).Exp }
	before := exp(runnerID)
	flee := br1FleeChoice(t, e)
	br1Turns(t, e, f, func(f game.Fighting) game.Action {
		if f.Field.Turn == runner {
			return flee(f)
		}
		return game.EndTurn{}
	}, func(s game.State) bool { _, ok := s.(game.Fighting); return !ok })
	for range br1FightCap { // the pages after the fight: no loot page on a flight
		s := e.build()
		if _, ok := s.(game.Looting); ok {
			t.Fatal("a loot page after a full flight")
		}
		if strings.HasPrefix(paneText(commonOf(s)), moveText) {
			break
		}
		press(e)
	}
	if got := e.partyIDs(); len(got) != 1 || got[0] != runnerID || e.partySize() != 1 {
		t.Errorf("party %v size %d after the flight, want only the runner %d", got, e.partySize(), runnerID)
	}
	if st, _ := e.ruleset().Fact(runnerID, rules.FactStatus); st != rules.StatusOkay || exp(runnerID) != before {
		t.Errorf("runner status %v exp %d (was %d)", st, exp(runnerID), before)
	}
}
