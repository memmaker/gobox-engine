package main

import (
	"gobox/pool"
	"gobox/ui"
	"testing"

	"gobox/game"
)

// TestTourPausesTimed: Rolf's tour (area 0 CityTour) walks itself: each
// DELAY reaches the presenter as a timed pause (a Continue with its Ms),
// not a bare story page that waits for a key. game.Question had no Ms, so
// the pause's time was dropped between pool and the front.
func TestTourPausesTimed(t *testing.T) {
	eng, err := pool.New("../../data/por", "../../data/por/sources")
	if err != nil {
		t.Skip(err)
	}
	eng.Cfg.SaveDir = t.TempDir()
	seed := uint32(7)
	eng.FixedSeed = &seed
	g := ui.NewGame(eng, eng.Data())
	g.Open("")
	for i := 0; i < 5; i++ {
		g.TitleDone()
	}
	if r := g.Send(game.QuickStart{}); r.Err != nil { // as the party menu's: the state follows
		t.Fatal(r.Err)
	}
	if r := g.Send(game.Begin{}); r.Err != nil {
		t.Fatal(r.Err)
	}
	pauses := 0
	for range 40 { // the greeting's two Approaches, then the walk's first stops
		w, pg, a, at := g.Window(), g.Page(), g.Asked(), g.Where()
		if w == game.WindowText && len(pg.View) == 0 || w == game.WindowPicture && len(pg.Layers) == 0 && game.CommonOf(eng.State()).Portrait == nil { // a speaker's portrait is gobox's to compose: no layer of the game's
			t.Fatalf("at %d,%d window %v with nothing in it", at.X, at.Y, w)
		}
		switch {
		case a == nil:
			t.Fatalf("at %d,%d nothing asked: a story page waiting for a key", at.X, at.Y)
		case a.Kind == game.Continue && a.Ms > 0: // a timed pause
			pauses++
			g.Send(game.ContinueAction{})
		default:
			g.Send(game.Reply{N: 0})
		}
	}
	if pauses == 0 {
		t.Fatal("no timed pause in the tour")
	}
}
