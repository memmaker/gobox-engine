package main

import (
	"strings"
	"testing"

	"gobox/engine"
	"gobox/game"
	"gobox/importer"
	"gobox/ui"
)

// TestSotsbOpeningBigPicture: Quick Start, then Begin Adventuring: the
// opening's first page (area17 Picture(119), "YOU HAVE GROWN BORED WITH
// PEACE...") is over BIGPIC block 119, whole-screen (it showed no picture).
func TestSotsbOpeningBigPicture(t *testing.T) {
	eng, err := engine.NewPool("../../data/sotsb", "../../data/sotsb/sources")
	if err != nil {
		t.Skip(err)
	}
	fo, err := importer.Read("../../data/sotsb")
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
	for range 1500 {
		f.Frame(1280, 800)
		sc, ok := eng.State().(game.Scene)
		if !ok || !strings.Contains(strings.ToUpper(sceneText(sc)), "BORED WITH PEACE") {
			continue
		}
		if !sc.Full {
			t.Errorf("the opening's page is not Full")
		}
		if l := sc.Picture; l != nil && strings.HasPrefix(l.File, "BIGPIC") && l.Block == 119 {
			return
		}
		t.Fatalf("the opening's page has no BIGPIC 119: picture %+v", sc.Picture)
	}
	t.Fatalf("no opening page: state %T", eng.State())
}

// sceneText is a Scene's words: its paragraphs and placed lines.
func sceneText(sc game.Scene) string {
	var b strings.Builder
	for _, s := range append(sc.Common.Text, sc.Text...) {
		b.WriteString(s + " ")
	}
	for _, l := range sc.Placed {
		for _, r := range l {
			b.WriteString(r.Text)
		}
		b.WriteString(" ")
	}
	return b.String()
}
