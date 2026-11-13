package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"gobox/front/classic"
	"gobox/front/present"
	"gobox/game/play"
	"gobox/pool"
	"por/playpor"
	"por/porgame"
)

// countedEngine counts the rules' State builds (each is the whole state:
// every member's item cards).
type countedEngine struct {
	*pool.Engine
	states int
}

func (c *countedEngine) State() play.State { c.states++; return c.Engine.State() }

// A loaded level-5 party (back-for-good) exploring builds State once a
// frame and a few times a step, not once per piece the front reads: the
// cost scales with the party's items, and the 3D view crawled at 77 fps
// (12 ms of draw CPU) when classic asked for State ~15 times a step and
// pool rebuilt each item's card again for its Ready.
func TestLoadedExploreStateBuilds(t *testing.T) {
	if err := os.Chdir("../.."); err != nil {
		t.Skip(err)
	}
	t.Cleanup(func() { os.Chdir("cmd/por") })
	src := "../../data/por/sources/save/saves/back-for-good"
	if _, err := os.Stat(src); err != nil {
		t.Skip(err)
	}
	eng, err := porgame.New("../../data/por/sources/assets", "../../data/por/sources")
	if err != nil {
		t.Skip(err)
	}
	dir := t.TempDir() // a copy: the save is never written
	if err := os.MkdirAll(filepath.Join(dir, "saves"), 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("cp", "-R", src, filepath.Join(dir, "saves")).CombinedOutput(); err != nil {
		t.Fatal(err, string(out))
	}
	eng.Cfg.SaveDir = dir
	eng.Explore = true
	eng.FullIntro.Store(false)
	seed := uint32(1)
	eng.FixedSeed = &seed
	c := &countedEngine{Engine: eng}
	data := playpor.Data(eng)
	f := present.NewFront(classic.NewGame(c, data), data)
	f.Start("back-for-good")
	f.Frame(1280, 800)
	f.Game.SequenceDone("intro")
	f.Game.TitleDone()
	f.Game.Settle()
	for i := 0; i < 600 && !f.Game.View().Exploring; i++ {
		f.Frame(1280, 800)
		f.Tick()
	}
	if v := f.Game.View(); !v.Exploring || v.Area != 18 {
		t.Fatalf("exploring %v area %d: the save did not load", v.Exploring, v.Area)
	}
	frames := func(n int) int {
		c.states = 0
		for range n {
			f.Frame(1280, 800)
			f.Tick()
		}
		return c.states
	}
	frames(60)
	at := time.Now()
	if n := frames(120); n != 0 { // nothing changed: the State built for the last change still stands
		t.Errorf("%d State builds in 120 idle frames, want 0", n)
	}
	t.Logf("%.2f ms a frame idle", float64(time.Since(at).Microseconds())/1000/120)
	c.states = 0
	f.Key(0x4B00) // a turn: a step's decision, the walk and its settle
	n := c.states
	n += frames(1)
	if n > 8 { // the key's, the frame's View, the decision's, its outcome's, the settle's, the Draw's
		t.Errorf("%d State builds for a turn and its frame, want at most 8", n)
	}
}
