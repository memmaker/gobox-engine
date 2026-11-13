package engine

import (
	"testing"

	"gobox/game/play"
)

// TestAdvFight<GAME> (bug class: pool's fight board needed POR's art, so no other
// game's fight ever published play.Fighting): the dev combat scene on the game's
// imported folder reaches a party turn as play.Fighting with its board, and a
// round goes by (members end their turns).
func TestAdvFightCOTAB(t *testing.T) { advFight(t, advEnvCOTAB) }
func TestAdvFightSOTSB(t *testing.T) { advFight(t, advEnvSOTSB) }
func TestAdvFightPOD(t *testing.T)   { advFight(t, advEnvPOD) }

// advFightOn is the dev combat scene on env's game at a party turn.
func advFightOn(t *testing.T, env string) *Engine {
	t.Helper()
	e := partyTurnSceneOn(t, advEngine(t, env))
	f, ok := e.build().(play.Fighting)
	if !ok {
		t.Fatalf("party turn is %T, want play.Fighting", e.build())
	}
	if e.battle() == nil || len(f.Field.Units) == 0 {
		t.Fatalf("no board: battleShown %v, %d units", e.battleShown(), len(f.Field.Units))
	}
	return e
}

func advFight(t *testing.T, env string) {
	e := advFightOn(t, env)
	for range 400 {
		s := e.build()
		if f, ok := s.(play.Fighting); ok && f.Field.Round >= 1 {
			return // a round went by
		} else if !ok {
			t.Fatalf("the fight ended before a round: %T", s)
		}
		switch {
		case myTurn(s):
			if _, err := e.doNow(play.EndTurn{}); err != nil {
				t.Fatal("EndTurn:", err)
			}
		case e.waiting(PendFight):
			e.Step()
		default:
			press(e)
		}
	}
	t.Fatal("no round went by")
}
