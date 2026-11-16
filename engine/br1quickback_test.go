package engine

import (
	"testing"

	"gobox/game"
)

// br1PlayStepwise ends the party's turns through TurnChoice and paces the rest with Step, as the presenter does
// (Do would run the computer's turns on to the next player's turn): it stops when stop holds, after each state.
func br1PlayStepwise(t *testing.T, e *Engine, each func(game.State) (stop bool)) {
	t.Helper()
	for range br1FightCap {
		s := e.build()
		if _, ok := s.(game.Fighting); !ok {
			t.Fatalf("the fight ended: %T, log %q", s, e.ext.battleLog)
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
