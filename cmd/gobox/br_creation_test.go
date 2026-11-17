package main

import (
	"path/filepath"
	"testing"

	"gobox/buck"
	"gobox/game"
	"gobox/engine"
	"gobox/ui"
)

// brEngine is BR1's engine on a fresh import, saving into a temp dir.
func brEngine(t *testing.T) *engine.Engine {
	f, _ := brLaunch(t, "br1")
	eng, err := buck.New(f.Dir, filepath.Join(f.Dir, "sources"))
	if err != nil {
		t.Fatal(err)
	}
	eng.Cfg.SaveDir = t.TempDir()
	return eng
}

// TestBR1QuickStart: Quick Start makes BR1's team of six (rulesxxvc's QuickParty was nil: the menu offered it and it did
// nothing; then PickKind was not ported, and Gizmo's Need named "tech", not the ability key "tch").
func TestBR1QuickStart(t *testing.T) {
	eng := brEngine(t)
	g := ui.NewGame(eng, eng.Data())
	g.Open("")
	for range 5 {
		g.TitleDone()
	}
	if r := g.Send(game.QuickStart{}); r.Err != nil {
		t.Fatal(r.Err)
	}
	party := game.CommonOf(eng.State()).Party
	if len(party) != 6 {
		t.Fatalf("party of %d, want 6", len(party))
	}
	for _, m := range party {
		if m.IconPoses[0].Pixels == nil {
			t.Errorf("%s has no combat icon", m.Name)
		}
	}
}

// TestBR1IconPick: on the creation sheet the icon editor picks one of CHARS' whole icons (Body, the grid, Enter) and
// Keep takes it (it offered Pool's head, size and colours: Size panicked, and so did Exit, SetIconPics not ported).
// biosRight is the Right arrow as Front.Key takes it (the BIOS scan code).
const biosRight = 0x4D00

func TestBR1IconPick(t *testing.T) {
	eng := brEngine(t)
	data := eng.Data()
	f := ui.NewFront(ui.NewGame(eng, data), data)
	f.Start("")
	for i := 0; i < 400 && !f.TitleDrawn(); i++ { // Enter through the title to the party menu
		f.Frame(1280, 800)
		if i%40 == 39 {
			f.Key('\r')
		}
	}
	draftIcon := func() bool {
		as, _ := eng.State().(game.Assembling)
		if as.Draft == nil {
			t.Fatalf("no creation sheet: state %T", eng.State())
		}
		return as.Draft.IconPoses[0].Pixels != nil
	}
	for _, k := range []uint16{'p', 'c', 'I', 'B', biosRight, '\r', 'K'} {
		if k == 'I' && draftIcon() {
			t.Fatal("the draft has an icon before one is picked (SetDefaultIcon comes with Keep)")
		}
		f.Key(k)
		for range 30 {
			f.Frame(1280, 800)
		}
	}
	if !draftIcon() {
		t.Error("the picked icon is not the draft's")
	}
}
