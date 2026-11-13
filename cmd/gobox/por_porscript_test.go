package main

import (
	"os"
	"slices"
	"testing"
	"time"

	"gobox/engine"
	"gobox/game"
)

// explore: a headless engine (needs the game files in the repo root) that
// begins adventuring with the Quick Start party once Init and Play run, as
// -explore does; its roster goes to a temp dir, never the game's saves.
func explore(t *testing.T) *engine.Engine {
	t.Helper()
	if _, err := os.Stat("../../data/por/sources/dos/TITLE.DAX"); err != nil {
		t.Skip("no game files")
	}
	e, err := engine.NewPool("../../data/por/sources/assets", "../../data/por/sources")
	if err != nil {
		t.Fatal(err)
	}
	e.Cfg.SaveDir = t.TempDir()
	e.Explore = true
	return e
}

// common is s's Common.
func common(s game.State) game.Common {
	switch s := s.(type) {
	case game.Assembling:
		return s.Common
	case game.Exploring:
		return s.Common
	case game.Camping:
		return s.Common
	case game.Fighting:
		return s.Common
	case game.Scene:
		return s.Common
	case game.Trading:
		return s.Common
	case game.Looting:
		return s.Common
	case game.Over:
		return s.Common
	}
	return game.Common{}
}

// TestScriptInputPrompt: a script's InputString is a typed-line question
// with the text it printed; a typed line ends it.
func TestScriptInputPrompt(t *testing.T) {
	e := explore(t)
	e.DevAction = e.DevScene("password")
	e.Init()
	e.Start(e.Play)
	typed := func() *game.Question {
		if q := common(e.State()).Ask; q != nil && q.Typed {
			return q
		}
		return nil
	}
	for range 20 {
		if typed() != nil {
			break
		}
		pressKey(t, e)
	}
	sc, _ := e.State().(game.Scene) // a Scene's prompt is its last Text
	if q := typed(); q == nil || len(sc.Text) == 0 || sc.Text[len(sc.Text)-1] != "WHAT DO YOU SAY? (TYPE A SINGLE WORD)" {
		t.Fatalf("ask %+v text %q, want the script's typed line", common(e.State()).Ask, sc.Text)
	}
	if _, err := e.Do(game.Reply{Text: "ohlo"}); err != nil {
		t.Fatal(err)
	}
	if typed() != nil {
		t.Fatal("the line did not end")
	}
}

// settle steps the engine, as the presenter does once a frame, while
// something goes on with time, for d at most (default 300 ms: a picture may
// animate for ever), and returns the State.
func settle(e *engine.Engine, d ...time.Duration) game.State {
	max := 300 * time.Millisecond
	if len(d) > 0 {
		max = d[0]
	}
	for t0 := time.Now(); e.Pace() != engine.PaceNone && time.Since(t0) < max; {
		time.Sleep(time.Millisecond)
		e.Step()
	}
	return e.State()
}

// pressKey is the player's key to what waits, through Do: a page's
// Continue, a lone "Press" reply; with nothing to continue the engine
// steps on.
func pressKey(t *testing.T, e *engine.Engine) {
	t.Helper()
	st := e.State()
	cm, q := common(st), common(st).Ask
	if sc, ok := st.(game.Scene); ok {
		q = sc.Ask
	}
	for _, a := range []game.Action{game.ContinueAction{}, game.Reply{N: 0}} {
		if _, isReply := a.(game.Reply); isReply && (q == nil || len(q.Replies) != 1) {
			continue
		}
		if slices.Contains(cm.Allowed.Plain, a) {
			if _, err := e.Do(a); err != nil {
				t.Fatal(err)
			}
			return
		}
	}
	e.Step()
}
