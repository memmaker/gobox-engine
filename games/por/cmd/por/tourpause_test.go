package main

import (
	"gobox/front/present"
	"testing"

	"gobox/front"
	"gobox/game/play"

	"por/playpor"
	"por/porgame"
)

// TestTourPausesTimed: Rolf's tour (area 0 CityTour) walks itself: each
// DELAY reaches the presenter as a timed pause (front.AskPause with its Ms),
// not a bare story page that waits for a key. play.Question had no Ms, so
// the pause's time was dropped between pool and the front.
func TestTourPausesTimed(t *testing.T) {
	eng, err := porgame.New("../../../../data/por/sources/assets", "../../../../data/por/sources")
	if err != nil {
		t.Skip(err)
	}
	eng.Cfg.SaveDir = t.TempDir()
	seed := uint32(7)
	eng.FixedSeed = &seed
	g := present.NewGame(eng, playpor.Data(eng))
	g.Open("")
	for i := 0; i < 5; i++ {
		g.TitleDone()
	}
	if err := eng.QuickStart(); err != nil {
		t.Fatal(err)
	}
	if r := g.Send(play.Begin{}); r.Err != nil {
		t.Fatal(r.Err)
	}
	pauses := 0
	for range 40 { // the greeting's two Approaches, then the walk's first stops
		v := g.View()
		if v.Window == front.WinText && len(v.ViewText) == 0 || v.Window == front.WinPicture && len(v.Pictures) == 0 {
			t.Fatalf("at %d,%d window %v with nothing in it", v.X, v.Y, v.Window)
		}
		switch {
		case v.Ask == nil:
			t.Fatalf("at %d,%d nothing asked: a story page waiting for a key", v.X, v.Y)
		case v.Ask.Kind == front.AskPause:
			if v.Ask.Ms <= 0 {
				t.Fatalf("pause of %d ms", v.Ask.Ms)
			}
			pauses++
			g.Send(play.Continue{})
		default:
			g.Send(play.Reply{N: 0})
		}
	}
	if pauses == 0 {
		t.Fatal("no timed pause in the tour")
	}
}
