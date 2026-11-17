package main

import (
	"gobox/pool"
	"gobox/ui"
	"testing"
	"time"

	"gobox/game"
)

// The classic front on the game.Game : Open starts the engine,
// exploring shows the party, a turn turns it, camp opens and closes.
func TestPlayFront(t *testing.T) {
	eng, err := pool.New("../../data/por", "../../data/por/sources")
	if err != nil {
		t.Skip(err)
	}
	eng.Cfg.SaveDir = t.TempDir()
	seed := uint32(7)
	eng.FixedSeed = &seed
	eng.Explore = true
	data := eng.Data()
	g := ui.NewGame(eng, data)
	g.Open("")
	for i := 0; g.Mode() == game.ModeTitle && i < 5; i++ { // the title, as the front ends it
		g.TitleDone()
	}
	if g.Mode() != game.ModeExplore || len(g.Party()) == 0 {
		t.Fatalf("mode %v, %d members, want exploring with a party", g.Mode(), len(g.Party()))
	}
	if !g.Flags().Exploring {
		t.Fatal("no walk map")
	}
	f := (g.Where().Facing + 2) % 8
	if r := g.Send(game.Face{Facing: f}); r.Err != nil {
		t.Fatal("Face:", r.Err)
	}
	if got := g.Where().Facing; got != f {
		t.Errorf("facing %d, want %d", got, f)
	}
	if r := g.Send(game.MakeCamp{}); r.Err != nil {
		t.Fatal("Camp:", r.Err)
	}
	if m := g.Mode(); m != game.ModeCamp {
		t.Errorf("after Camp mode %v, want camp", m)
	}
	if r := g.Send(game.BreakCamp{}); r.Err != nil {
		t.Fatal("LeaveCamp:", r.Err)
	}
	if m := g.Mode(); m != game.ModeExplore {
		t.Errorf("after LeaveCamp mode %v, want explore", m)
	}
}

// The auto-walk is gobox's (WalkTo: Face, RunEvent, Arrive through Do over
// the play state's Map.Walk): a square at a time to the square, by Step alone.
func TestPlayFrontWalk(t *testing.T) {
	eng, err := pool.New("../../data/por", "../../data/por/sources")
	if err != nil {
		t.Skip(err)
	}
	eng.Cfg.SaveDir = t.TempDir()
	seed := uint32(7)
	eng.FixedSeed = &seed
	eng.Explore = true
	eng.DevAction = eng.DevScene("seen") // auto-walk keeps to seen squares (e9c58085)
	data := eng.Data()
	g := ui.NewGame(eng, data)
	g.Open("")
	for i := 0; g.Mode() == game.ModeTitle && i < 5; i++ {
		g.TitleDone()
	}
	x, y := g.Where().X, g.Where().Y // the start of Phlan: a corridor running west
	if !g.WalkTo(x-3, y) {
		t.Fatal("WalkTo refused")
	}
	for until := time.Now().Add(3 * time.Second); g.Pace() == game.PaceWalk && time.Now().Before(until); { // a square per walkPace
		if r := g.Step(); r.Err != nil {
			t.Fatal("Step:", r.Err)
		}
	}
	if w := g.Where(); w.X != x-3 || w.Y != y {
		t.Fatalf("walk from %d,%d to %d,%d ended at %d,%d", x, y, x-3, y, w.X, w.Y)
	}
}
