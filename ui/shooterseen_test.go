package ui

import (
	"testing"

	"gobox/game"
	"gobox/pool"
)

// TestShooterOnBoard: in a real fight every shot flies from a unit the
// moment's board holds on the shot's start square: the shooter is drawn
// where the shot leaves (no arrow from an empty square or the map's corner).
func TestShooterOnBoard(t *testing.T) { shooterOnBoard(t, 0) }

func shooterOnBoard(t *testing.T, monster int) {
	eng, err := pool.New("../data/por", "../data/por/sources")
	if err != nil {
		t.Skip(err)
	}
	eng.Cfg.SaveDir = t.TempDir()
	seed := uint32(3)
	eng.FixedSeed, eng.Explore = &seed, true
	eng.FullIntro.Store(false)
	eng.DevMonster = monster
	eng.DevAction = eng.DevScene("combat")
	data := eng.Data()
	f := NewFront(NewGame(eng, data), data)
	f.Start("")
	f.Frame(flowW, flowH)
	f.Game.SequenceDone("intro")
	f.Game.TitleDone()
	f.Game.Settle()
	quick, shots := false, 0
	for range 4000 {
		_, fighting := eng.State().(game.Fighting)
		if fighting && f.P.board != nil && f.P.board.Turn != nil {
			quick = f.P.turnAct(Action{Kind: ActQuick, N: 1}) || quick // Quick All
		}
		if quick && !fighting && !f.P.Playing() {
			break
		}
		f.Frame(flowW, flowH)
		f.Tick()
		for i := range f.P.replay.q {
			fr := &f.P.replay.q[i]
			if fr.fly == nil || fr.Board == nil || i > 0 && f.P.replay.q[i-1].fly != nil && f.P.replay.q[i-1].Unit == fr.Unit && f.P.replay.q[i-1].To == fr.To {
				continue // a flight's later frames
			}
			shots++
			on := false
			for _, u := range fr.Board.Units {
				on = on || u.ID == fr.Unit && u.At == fr.At
			}
			if !on {
				t.Errorf("shot by unit %d from %v to %v: the board holds no such unit there (%d units)", fr.Unit, fr.At, fr.To, len(fr.Board.Units))
			}
		}
		f.P.replay.q = nil // checked: the fight goes on without playing it out
		f.P.Pace()
		f.Key(13)
	}
	t.Logf("quick %v state %T", quick, eng.State())
	if shots == 0 {
		t.Skip("no shot flew")
	}
	t.Logf("%d shots", shots)
}
