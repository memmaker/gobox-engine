package ui

import (
	"slices"
	"testing"

	"gobox/engine"
	"gobox/game"
	"gobox/pool"
)

// hallFront is POR's Front in the training hall (the engine's "hall" dev
// scene): its party menu up over the 3D view (skipped when POR is not
// imported). Walking there is not this test's: auto-walk goes only over
// squares seen (TestWalkKnownWays), a fresh town has three.
func hallFront(t *testing.T) (*Front, *engine.Engine) {
	eng, err := pool.New("../data/por", "../data/por/sources")
	if err != nil {
		t.Skip(err)
	}
	eng.Cfg.SaveDir = t.TempDir()
	seed := uint32(7)
	eng.FixedSeed, eng.Explore = &seed, true
	eng.FullIntro.Store(false)
	eng.DevAction = eng.DevScene("hall")
	data := eng.Data()
	f := NewFront(NewGame(eng, data), data)
	f.Start("")
	f.Frame(flowW, flowH)
	f.Game.SequenceDone("intro")
	f.Game.TitleDone()
	f.Game.Settle()
	for range 10 {
		f.Frame(flowW, flowH)
		f.Tick()
	}
	if as, ok := eng.State().(game.Assembling); !ok || !as.Where.Hall {
		t.Fatalf("no hall: %T", eng.State())
	}
	return f, eng
}

// The hall's party menu (class: a presenter rule call naming "the selected
// member" as -1, which no State offers; a gobox screen opened under a menu
// the State still asks): every function acts for the selected member, a card
// or a member key selects, View's sheet shows and closes.
func TestHallPartyMenu(t *testing.T) {
	f, eng := hallFront(t)
	p := &f.P
	selected := func() int { return game.CommonOf(eng.State()).Selected }
	menu := func(what string) {
		t.Helper()
		if !p.main || p.sheetShown() || p.creating || p.gridPick != nil {
			t.Fatalf("%s: the hall's menu is not back (main %v, sheet %v, creating %v, picker %v, note %q)", what, p.main, p.sheetShown(), p.creating, p.gridPick != nil, p.note().Text)
		}
	}
	menu("entered")
	flowKey(f, 0x3C00) // F2
	if selected() != 1 {
		t.Fatalf("F2 selected member %d (%q)", selected(), p.note().Text)
	}
	r := p.mm.cards[2]
	if f.Hover(r.Min.X+4, r.Min.Y+4, false); !f.Click(r.Min.X+4, r.Min.Y+4) {
		t.Fatal("the third card took no click")
	}
	f.Frame(flowW, flowH)
	if selected() != 2 {
		t.Fatalf("the third card's click selected member %d (%q)", selected(), p.note().Text)
	}
	name := p.party()[2].Name
	flowKey(f, 'v')
	if sp := p.sheetPage(); p.main || sp == nil || sp.Name != name {
		t.Fatalf("View: main menu up %v, sheet %+v, want %s's", p.main, sp, name)
	}
	flowKey(f, 27)
	menu("View, Esc")
	flowKey(f, 'm')
	if !p.creating || p.creation() == nil || f.Game.Subject().Name != name {
		t.Fatalf("Modify: sheet up %v for %q, want %s (%q)", p.creating, f.Game.Subject().Name, name, p.note().Text)
	}
	flowKey(f, 27)
	flowKey(f, 'y') // abandon the changes
	menu("Modify, Esc")
	flowKey(f, 'c')
	if !p.creating {
		t.Fatalf("Create: no sheet (%q)", p.note().Text)
	}
	flowKey(f, 'r') // Roll
	flowKey(f, 27)
	flowKey(f, 'y')
	menu("Create, Esc")
	flowKey(f, 'r')
	flowKey(f, 'y')
	if n := len(p.party()); n != 5 || slicesHas(p.party(), name) {
		t.Fatalf("Remove: party of %d, %s in it %v (%q)", n, name, slicesHas(p.party(), name), p.note().Text)
	}
	flowKey(f, 'a')
	if pk := p.picker(p.ink()); pk == nil || !slices.Contains(pk.Labels, name) {
		t.Fatalf("Add: no picker with %s: %+v", name, pk)
	}
	flowKey(f, 27)
	menu("Add, Esc")
	name = p.party()[selected()].Name
	flowKey(f, 'd') // the hall has no Drop
	if n := len(p.party()); n != 5 || !slicesHas(p.party(), name) || p.mm.ask != askNone {
		t.Fatalf("D in the hall: party of %d, %s in it %v, asking %v", n, name, slicesHas(p.party(), name), p.mm.ask)
	}
	menu("D")
	flowKey(f, 'l')
	if _, ok := eng.State().(game.Exploring); !ok {
		t.Fatalf("Leave the Hall: state %T", eng.State())
	}
}

func slicesHas(ms []game.Member, name string) bool {
	return slices.ContainsFunc(ms, func(m game.Member) bool { return m.Name == name })
}

// The title's party menu: View's sheet shows over it and Modify is the
// selected member's.
func TestTitlePartyMenu(t *testing.T) {
	f, _ := flowFront(t, "")
	p := &f.P
	flowKey(f, 'e') // out of the Party submenu
	flowKey(f, 'q') // Quick Start
	for range 20 {
		f.Frame(flowW, flowH)
	}
	flowKey(f, 'p')
	if len(p.party()) < 2 || !p.mm.sub {
		t.Fatalf("party of %d, submenu %v", len(p.party()), p.mm.sub)
	}
	flowKey(f, 0x3C00) // F2
	name := p.party()[1].Name
	flowKey(f, 'v')
	if sp := p.sheetPage(); p.main || sp == nil || sp.Name != name {
		t.Fatalf("View: main menu up %v, sheet %+v, want %s's", p.main, sp, name)
	}
	flowKey(f, 27)
	flowKey(f, 'm')
	if !p.creating || f.Game.Subject().Name != name {
		t.Fatalf("Modify: sheet up %v for %q, want %s (%q)", p.creating, f.Game.Subject().Name, name, p.note().Text)
	}
}
