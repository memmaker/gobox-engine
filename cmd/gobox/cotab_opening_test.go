package main

import (
	"testing"

	"gobox/engine"
	"gobox/game"
	"gobox/importer"
	"gobox/ui"
	"slices"
)

// TestCotabArm: the opening's marked sword arm (Picture(10), PIC2 block
// 10) cycles its 3 frames while the menu waits (cotab's Picture marks the
// view drawn), and F12's DOS art shows the DOS frames on the waiting page.
func TestCotabArm(t *testing.T) {
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
	f.Key('b')
	var arm *game.Common
	for i := 0; i < 400 && arm == nil; i++ {
		f.Frame(1280, 800)
		if i == 20 {
			f.Key('\r') // past the waking page
		}
		if x, ok := eng.State().(game.Scene); ok && x.Picture != nil && x.Picture.Block == 10 {
			arm = &x.Common
		}
	}
	if arm == nil {
		t.Fatal("no arm picture")
	}
	if len(arm.Frames) != 3 || len(arm.Delays) != 3 || slices.Equal(arm.Frames[0].Pixels, arm.Frames[1].Pixels) {
		t.Fatalf("arm frames %d, delays %v: want 3 different frames", len(arm.Frames), arm.Delays)
	}
	engine.DOSArt.Store(!engine.DOSArt.Load())
	defer engine.DOSArt.Store(!engine.DOSArt.Load())
	x, _ := eng.State().(game.Scene)
	if x.Picture == nil || slices.Equal(x.Picture.Pixels, arm.Picture.Pixels) {
		t.Error("the other art (F12) leaves the waiting page's picture as it was")
	}
}
