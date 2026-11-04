// Package parity checks gobox's classic front against a game: the State it
// publishes must be what it published (state_test.go), and the game's goldens must come out byte-identical through
// classic.
package parity

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

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
	return engineSeed(t, explore, dev, 7)
}

// engineSeed is engine with the dice seeded seed.
func engineSeed(t *testing.T, explore bool, dev string, seed uint32) *rules.Engine {
	dir := filepath.Join(porDir(t), "data")
	if _, err := os.Stat(filepath.Join(dir, "dos", "TITLE.DAX")); err != nil {
		t.Skip("no game files")
	}
	e, err := rules.New(dir)
	if err != nil {
		t.Fatal(err)
	}
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

// games are the classic.Game each engine is driven through (the tests never ask por for its View).
var games = map[*rules.Engine]*classic.Game{}

func gameOf(e *rules.Engine) *classic.Game {
	if games[e] == nil {
		games[e] = classic.NewGame(e.Rules(), e.Data())
	}
	return games[e]
}

// settled answers as through classic.Game, each followed by the engine's
// paced steps until it stops pacing (a count, not a clock).
func settled(e *rules.Engine, as ...front.Action) front.View {
	g := gameOf(e)
	v := g.View()
	for _, a := range as {
		if a != (front.Action{}) {
			g.Answer(a)
		}
		g.Settle()
		e.Apply() // publish: the rules' State is what they last published
		v = g.View()
		for i := 0; v.Pace != front.PaceNone && v.Pace != front.PaceAnim && i < 20000; i++ {
			e.Step()
			e.Apply()
			v = g.View()
		}
	}
	return v
}

// same says the State at name is what por published (at) and classic's creation sheet equals por's.
func same(t *testing.T, name string, e *rules.Engine, cols, rows int) {
	t.Helper()
	at(t, name, e)
	s := e.Rules().State()
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
	same(t, "hall", h, 40, 25)
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
				t.Errorf("%s line %d:\n got %s\nwant %s", name, i+1, gl[i], wl[i])
				return
			}
		}
		t.Errorf("%s: %d lines, want %d", name, len(gl), len(wl))
	}
}
