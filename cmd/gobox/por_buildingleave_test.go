package main

import (
	"gobox/ui"
	"testing"

	"gobox/engine"
	"gobox/game"
)

// TestBuildingLeaveStays (bug class: a building's cell event re-runs after
// its exit): every shop, temple and hall reached walking New Phlan is left
// with game.Leave; settling after it must not open the same place again on
// the square the party stands on.
// The party may stand on the shop's square after it (OpenShop, then
// ResetVisits) or be put back on the square before (the temples' MustLeave:
// PartyX = PrevX), as the original; a walk that crosses the square again
// enters again, which is the original's too.
func TestBuildingLeaveStays(t *testing.T) {
	g, eng := porWalker(t)
	settle := func() (game.Trading, bool) { return settleBuilding(t, g, eng) }
	seen := 0
	for y := range 16 {
		for x := range 16 {
			if !g.WalkTo(x, y) {
				continue
			}
			for i := 0; i < 100 && g.Pace() == game.PaceWalk; i++ {
				g.WalkStep()
			}
			tr, open := settle()
			if !open {
				continue
			}
			seen++
			at := g.Where()
			if r := g.Send(game.Leave{}); r.Err != nil {
				t.Fatalf("%v at %d,%d: Leave: %v", tr.Kind, at.X, at.Y, r.Err)
			}
			if _, again := settle(); again {
				w := g.Where()
				t.Errorf("%v at %d,%d opened again after Leave, at %d,%d", tr.Kind, at.X, at.Y, w.X, w.Y)
			}
		}
	}
	if seen == 0 {
		t.Fatal("no building reached")
	}
}

// porWalker is POR started in New Phlan for a scripted walk.
func porWalker(t *testing.T) (*ui.Game, *engine.Engine) {
	eng, err := engine.NewPool("../../data/por", "../../data/por/sources")
	if err != nil {
		t.Skip(err)
	}
	eng.Cfg.SaveDir = t.TempDir()
	seed := uint32(7)
	eng.FixedSeed = &seed
	eng.Explore = true
	g := ui.NewGame(eng, eng.Data())
	g.Open("")
	for i := 0; g.Mode() == game.ModeTitle && i < 5; i++ {
		g.TitleDone()
	}
	return g, eng
}

// settleBuilding answers every question with its Esc reply, else its last
// (no, exit), until the party explores or a building is open.
func settleBuilding(t *testing.T, g *ui.Game, eng *engine.Engine) (game.Trading, bool) {
	t.Helper()
	for range 300 {
		g.Settle()
		switch s := eng.State().(type) {
		case game.Trading:
			return s, true
		case game.Assembling: // the training hall's party menu or a fall
			if g.Send(game.Begin{}).Err == nil {
				continue
			}
		case game.Exploring:
			if s.Ask == nil {
				return game.Trading{}, false
			}
		}
		q := game.CommonOf(eng.State()).Ask
		if q != nil && (q.Number || q.Typed) {
			g.Send(game.Reply{Text: "0"})
		} else if i, ok := escOf(q); ok {
			g.Send(game.Reply{N: i})
		} else if q != nil && len(q.Replies) > 0 {
			g.Send(game.Reply{N: len(q.Replies) - 1})
		} else if g.Send(game.ContinueAction{}).Err != nil {
			g.Send(game.Leave{})
		}
	}
	t.Fatalf("settle stuck in %T", eng.State())
	return game.Trading{}, false
}

func escOf(q *game.Question) (int, bool) {
	if q == nil {
		return 0, false
	}
	return q.Esc.Reply()
}

// TestTavernGambleEnds: New Phlan's gambler (10,8; the native TavernGamble,
// the script's WagerPrompt copy never runs) takes a wager, and Esc on
// WILL YOU GO AGAIN? (its No) ends the game: the party explores again.
// A walker answering Yes (Reply 0) gambles on, as the original.
func TestTavernGambleEnds(t *testing.T) {
	g, eng := porWalker(t)
	wagered, gambling := false, false
	for range 300 {
		if w := g.Where(); game.CommonOf(eng.State()).Ask == nil && (w.X != 10 || w.Y != 8) {
			if !g.WalkTo(10, 8) {
				t.Fatalf("no walk from %d,%d", w.X, w.Y)
			}
			for i := 0; i < 100 && g.Pace() == game.PaceWalk; i++ {
				g.WalkStep()
			}
			continue
		}
		g.Settle()
		q := game.CommonOf(eng.State()).Ask
		switch _, exploring := eng.State().(game.Exploring); {
		case exploring && q == nil && wagered:
			return
		case exploring && q == nil:
			t.Fatal("at the tavern, no gamble")
		case q.Number:
			wagered = true
			g.Send(game.Reply{Text: "10"})
		case len(q.Replies) == 2 && !gambling: // DOES ANYONE WANT TO GAMBLE? yes
			gambling = true
			g.Send(game.Reply{N: 0})
		case len(q.Replies) == 2 && wagered: // WILL YOU GO AGAIN? Esc
			if r := g.Send(game.Cancel{}); r.Err != nil {
				t.Fatal("Esc:", r.Err)
			}
		default:
			g.Send(game.Reply{N: 0})
		}
	}
	t.Fatal("the gamble did not end")
}
