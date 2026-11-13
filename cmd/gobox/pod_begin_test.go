package main

import (
	"testing"

	"gobox/engine"
	"gobox/game"
	"gobox/importer"
	"gobox/ui"
)

// TestPodBegin: Quick Start, then Begin Adventuring, reaches exploring (it panicked: scripts keyed without pod's ScriptFixes).
func TestPodBegin(t *testing.T) {
	eng, err := engine.NewPool("../../data/pod", "../../data/pod/sources")
	if err != nil {
		t.Skip(err)
	}
	fo, err := importer.Read("../../data/pod")
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
	data := eng.Data()
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
		if _, ok := eng.State().(game.Exploring); ok {
			return
		}
	}
	x, _ := eng.State().(game.Exploring)
	t.Fatalf("not exploring after Begin: state %T, where %+v", eng.State(), x.Where)
}
