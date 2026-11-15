package engine

import (
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"gobox/game"
)

// TestBuildReadsNoDisk (REFACTOR.md G2c): build runs after every action and
// reads no file. Once the state has settled, the adventure and save folders
// are pointed at nothing; the build is the same state as before.
func TestBuildReadsNoDisk(t *testing.T) {
	scenes := map[string]func(t *testing.T) *Engine{
		"party menu": quickStarted,
		"hall":       func(t *testing.T) *Engine { return devPlay(t, "hall") },
		"shop":       func(t *testing.T) *Engine { return devPlay(t, "shop") },
		"loot":       func(t *testing.T) *Engine { return devPlay(t, "loot") },
		"fight":      func(t *testing.T) *Engine { return devPlay(t, "meleefight") },
	}
	for name, open := range scenes {
		t.Run(name, func(t *testing.T) {
			e := open(t)
			want := e.build() // reads what it needs once
			adv, save := e.adv, e.Cfg.SaveDir
			gone := filepath.Join(t.TempDir(), "gone")
			e.adv, e.Cfg.SaveDir = gone, gone
			defer func() { e.adv, e.Cfg.SaveDir = adv, save }()
			if got := e.build(); !reflect.DeepEqual(got, want) {
				t.Errorf("build with no files differs:\n got %+v\nwant %+v", got, want)
			}
		})
	}
}

// The party menu's Roster follows the roster files: Quick Start's kept
// character is listed, marked "* " while in the party; forgotten, it is gone;
// kept again, it is back.
func TestRosterListRefreshed(t *testing.T) {
	e := quickStarted(t)
	roster := func() []string { a, _ := e.build().(game.Assembling); return a.Roster }
	marked := func(l []string) int {
		return len(slices.DeleteFunc(slices.Clone(l), func(n string) bool { return n[:2] != "* " }))
	}
	n := len(roster()) // Quick Start's saves each rewrite the list without the marked: one name is left
	if got := roster(); n == 0 || marked(got) != n {
		t.Fatalf("roster %v after Quick Start", got)
	}
	party := e.state().Party
	e.state().Party = nil
	if got := roster(); len(got) != n || marked(got) != 0 {
		t.Fatalf("roster %v after dropping the party", got)
	}
	e.OpenPicker(PickForget)
	for len(e.BuildCharList()) > 0 {
		e.Pick(0)
		e.OpenPicker(PickForget)
	}
	if got := roster(); len(got) != 0 {
		t.Fatalf("roster %v after forgetting all", got)
	}
	if err := e.SaveCharacter(e.charID(party[0]), ""); err != nil {
		t.Fatal(err)
	}
	if got := roster(); len(got) != 1 {
		t.Fatalf("roster %v after keeping one", got)
	}
}
