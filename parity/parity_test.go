// Package parity checks gobox's classic front against a game's own View:
// each screen classic builds from game.State must equal the one the game
// published, and the game's goldens must come out byte-identical through
// classic.
package parity

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"gobox/front"
	"gobox/front/classic"
	"gobox/front/present"
	"por/rules"
)

// porDir is por's checkout (the replace in go.mod).
func porDir(t *testing.T) string {
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "por").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

// engine is por started headless on its data (dev: a -dev scene, "" none).
func engine(t *testing.T, explore bool, dev string) *rules.Engine {
	dir := filepath.Join(porDir(t), "data")
	if _, err := os.Stat(filepath.Join(dir, "dos", "TITLE.DAX")); err != nil {
		t.Skip("no game files")
	}
	e, err := rules.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Cfg.SaveDir = t.TempDir()
	e.Explore = explore
	if dev != "" {
		e.DevAction = e.DevScene(dev)
	}
	e.Init()
	e.Start(e.Play)
	return e
}

// settled applies as, each until the engine stops pacing.
func settled(e *rules.Engine, as ...front.Action) front.View {
	v := e.View()
	for _, a := range as {
		v = e.Apply(a)
		for t0 := time.Now(); v.Pace != front.PaceNone && time.Since(t0) < 300*time.Millisecond; {
			time.Sleep(time.Millisecond)
			e.Step()
			v = e.Apply()
		}
	}
	return v
}

// same says classic's party menu and creation sheet equal por's.
func same(t *testing.T, name string, e *rules.Engine, cols, rows int) {
	t.Helper()
	v, s := e.View(), e.Rules().State()
	c := classic.View(v, s)
	if e.CreationView() == nil { // por's View keeps the menu under the sheet: the presenter drops it (ownCreation)
		if !reflect.DeepEqual(c.MainMenu, v.MainMenu) || c.Hall != v.Hall || c.Logo != v.Logo || c.Backdrop != v.Backdrop {
			t.Errorf("%s: menu %v hall %v logo %v, por %v %v %v", name, c.MainMenu, c.Hall, c.Logo != nil, v.MainMenu, v.Hall, v.Logo != nil)
		}
	} else if c.MainMenu != nil {
		t.Errorf("%s: menu %v under the sheet", name, c.MainMenu)
	}
	if got, want := classic.Creation(s, cols, rows), e.CreationView(); !reflect.DeepEqual(got, want) {
		t.Errorf("%s: creation\n got %+v\nwant %+v", name, got, want)
	}
}

// partyMenu is por past the title at the party menu.
func partyMenu(t *testing.T) *rules.Engine {
	e := engine(t, false, "")
	v := settled(e)
	for i := 0; v.MainMenu == nil && i < 20; i++ {
		v = settled(e, front.Action{Kind: front.ActOK})
	}
	if v.MainMenu == nil {
		t.Fatal("no party menu")
	}
	return e
}

func TestParity(t *testing.T) {
	e := partyMenu(t)
	same(t, "empty party", e, 40, 25)
	if err := e.CreateCharacter(); err != nil {
		t.Fatal(err)
	}
	settled(e, front.Action{})
	same(t, "creation", e, 40, 25)
	e.SetStat(0, "18/50") // the human fighter it starts as
	settled(e, front.Action{})
	if e.Rules().State().Creation.Draft.Str00 != 50 {
		t.Fatal("no 18/50")
	}
	same(t, "creation 18/50", e, 40, 25)
	for _, f := range [][2]int{{front.FieldSex, 1}, {front.FieldRace, 2}, {front.FieldClass, 3}, {front.FieldAlign, 4}} {
		e.SetField(f[0], f[1])
		settled(e, front.Action{})
		same(t, "creation field change", e, 40, 25)
	}
	settled(e, front.Action{Kind: front.ActResize, X: 83, Y: 80})
	same(t, "creation 83x80", e, 83, 80)
	settled(e, front.Action{Kind: front.ActResize, X: 40, Y: 25})
	if a := e.Rules().Pending(); a == nil || a.Prompt != e.CreationView().Abandon {
		t.Fatalf("abandon: %+v", a)
	}
	e.EndCreation(false)
	settled(e, front.Action{})
	same(t, "abandoned", e, 40, 25)
	e.QuickStart()
	settled(e, front.Action{})
	same(t, "party", e, 40, 25)

	x := engine(t, true, "")
	settled(x)
	same(t, "exploring", x, 40, 25)

	h := engine(t, true, "hall")
	settled(h, front.Action{})
	if !h.View().Hall {
		t.Fatal("no hall")
	}
	same(t, "hall", h, 40, 25)
}

// TestGoldens: por's creation goldens through classic (present.WithRules).
func TestGoldens(t *testing.T) {
	e := partyMenu(t)
	p := &present.Presenter{}
	g := present.WithRules(e, p)
	p.SetEngine(g)
	shown := func(as ...front.Action) front.View {
		for _, a := range as {
			p.Decide(g.View(), a)
			for _, f := range p.Calls() {
				f(g)
			}
			settled(e, append(p.Actions(), front.Action{})...)
		}
		settled(e)
		return p.Own(g.View())
	}
	v := shown(front.Action{Kind: front.ActCreate})
	if v.Creation == nil {
		t.Fatal("no creation")
	}
	golden(t, "creation", v)
	golden(t, "pickhead", shown(front.Action{Kind: front.ActPicker, N: front.PickHead}))
}

// golden compares the screen's dump with por's presentdata/testdata/name.txt.
func golden(t *testing.T, name string, v front.View) {
	t.Helper()
	want, err := os.ReadFile(filepath.Join(porDir(t), "presentdata", "testdata", name+".txt"))
	if err != nil {
		t.Fatal(err)
	}
	if got := present.ScreenDump(v); got != string(want) {
		gl, wl := strings.Split(got, "\n"), strings.Split(string(want), "\n")
		for i := range min(len(gl), len(wl)) {
			if gl[i] != wl[i] {
				t.Fatalf("%s line %d:\n got %s\nwant %s", name, i+1, gl[i], wl[i])
			}
		}
		t.Fatalf("%s: %d lines, want %d", name, len(gl), len(wl))
	}
}
