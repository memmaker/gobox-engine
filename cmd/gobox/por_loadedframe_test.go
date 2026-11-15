package main

import (
	"gobox/engine"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"gobox/ui"
)

// A loaded level-5 party (back-for-good) exploring builds State once per
// change, not once per piece the front reads: the cost scales with the
// party's items (the 3D view crawled at 77 fps when classic asked for State
// ~15 times a step). Idle frames build none, a turn exactly one (its Do).
func TestLoadedExploreStateBuilds(t *testing.T) {
	src := "../../data/por/sources/save/saves/back-for-good"
	if _, err := os.Stat(src); err != nil {
		t.Skip(err)
	}
	eng, err := engine.NewPool("../../data/por", "../../data/por/sources")
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
	data := eng.Data()
	f := ui.NewFront(ui.NewGame(eng, data), data)
	f.Start("back-for-good")
	f.Frame(1280, 800)
	f.Game.SequenceDone("intro")
	f.Game.TitleDone()
	f.Game.Settle()
	for i := 0; i < 600 && !f.Game.Flags().Exploring; i++ {
		f.Frame(1280, 800)
		f.Tick()
	}
	if ex, area := f.Game.Flags().Exploring, f.Game.Where().Area; !ex || area != "18" {
		t.Fatalf("exploring %v area %s: the save did not load", ex, area)
	}
	frames := func(n int) int {
		built := eng.StateBuilds
		for range n {
			f.Frame(1280, 800)
			f.Tick()
		}
		return eng.StateBuilds - built
	}
	frames(60)
	at := time.Now()
	if n := frames(120); n != 0 { // nothing changed: the State built for the last change still stands
		t.Errorf("%d State builds in 120 idle frames, want 0", n)
	}
	t.Logf("%.2f ms a frame idle", float64(time.Since(at).Microseconds())/1000/120)
	built := eng.StateBuilds
	f.Key(0x4B00) // a turn: one Do
	n := eng.StateBuilds - built
	if n += frames(1); n != 1 {
		t.Errorf("%d State builds for a turn and its frame, want 1", n)
	}
}
