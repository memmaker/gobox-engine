package main

import (
	"testing"

	"gobox/asset"
	"gobox/engine"
	"gobox/game"
	"gobox/importer"
	"gobox/ui"
)

// TestCotabBeginWalls: Quick Start, then Begin Adventuring, puts cotab in
// its own start set with a 3D view of remastered walls (it began in POR's
// set 3, no walls loaded, and showed a black screen).
func TestCotabBeginWalls(t *testing.T) {
	eng, err := engine.NewPool("../../data/cotab", "../../data/cotab/sources")
	if err != nil {
		t.Skip(err)
	}
	fo, err := importer.Read("../../data/cotab")
	if err != nil {
		t.Fatal(err)
	}
	if err := eng.SetStart(fo.Manifest.Start); err != nil {
		t.Fatal(err)
	}
	if err := eng.SetEnd(fo.Manifest.End); err != nil {
		t.Fatal(err)
	}
	eng.Cfg.SaveDir = t.TempDir()
	seed := uint32(1)
	eng.FixedSeed = &seed
	data := poolData(eng, "cotab")
	f := ui.NewFront(ui.NewGame(eng, data), data)
	f.Start("")
	for i := range 400 { // Enter through the title to the party menu
		f.Frame(1280, 800)
		if _, ok := eng.State().(game.Assembling); ok && i > 50 {
			break
		}
		if i%40 == 0 {
			f.Key('\r')
		}
	}
	f.Key('q') // Quick Start
	for range 30 {
		f.Frame(1280, 800)
	}
	f.Key('b') // Begin Adventuring
	for i := range 1500 {
		f.Frame(1280, 800)
		if _, ok := eng.State().(game.Scene); ok && i%20 == 19 { // the opening's pages
			f.Key('\r')
		}
		if x, ok := eng.State().(game.Exploring); ok && x.Map.Walls.Side != nil {
			for _, tex := range x.Map.Walls.Pool {
				if !asset.Has("pool", tex) {
					t.Errorf("wall texture %q not in the pool", tex)
				}
			}
			return
		}
	}
	x, _ := eng.State().(game.Exploring)
	t.Fatalf("no 3D walls after Begin: state %T, where %+v", eng.State(), x.Where)
}
