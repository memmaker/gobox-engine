package engine

import (
	"errors"
	"gobox/rules1e"
	"io/fs"
	"os"
	"testing"

	"gobox/rules"
)

// TestRosterIDs: a roster save gets its own RosterID, kept across saves; a
// .cha from before (no number) gets one when the roster loads, past the
// highest held, written once; Join finds a character by it.
func TestRosterIDs(t *testing.T) {
	e := testEngine(t)
	c := e.state().Fresh(rules1e.Kind{Race: rules1e.RaceHuman, Class: rules1e.ClsFighter}, nil)
	c.SetName("Alpha")
	e.SaveCharacter(e.charID(c), "")
	if c.PoolIDs.Roster != 1 {
		t.Fatalf("first roster id %d", c.PoolIDs.Roster)
	}
	e.SaveCharacter(e.charID(c), "")
	if c.PoolIDs.Roster != 1 {
		t.Fatalf("resaved: id %d", c.PoolIDs.Roster)
	}
	old := e.state().Fresh(rules1e.Kind{Race: rules1e.RaceElf, Class: rules1e.ClsMage}, nil)
	old.SetName("Beta")
	e.SaveCharacter(e.charID(old), "")
	p := e.rosterFile("Beta")
	b, _ := os.ReadFile(p)
	if err := os.WriteFile(p, b[:charRecSize], 0o644); err != nil { // a save from before the numbers
		t.Fatal(err)
	}
	e.setRosterNext(1) // and a lost counter: the files' own numbers still win
	e.BuildCharList()
	b, _ = os.ReadFile(p)
	if got := rules1e.ReadPoolIDs(b).Roster; got != 2 {
		t.Fatalf("migrated id %d, want 2", got)
	}
	if e.BuildCharList(); e.rosterNext() != 3 {
		t.Fatalf("migrated twice: next %d", e.rosterNext())
	}
	e.charID(&rules1e.Character{}) // bug class: registry ids no longer start at 1, so a position id (i+1) would name the wrong character
	e.charID(&rules1e.Character{})
	id, err := e.Creation().Join(2)
	if err != nil || len(e.state().Party) != 1 || e.state().Party[0].Name() != "Beta" || id != e.partyIDs()[0] {
		t.Fatalf("join: %v %v", id, err)
	}
	if d, err := e.Creation().Draft(id); err != nil || d.Name != "Beta" {
		t.Fatalf("draft of the joined id: %q %v", d.Name, err)
	}
	if n := e.ruleset().Sheet(id); n.Name != "Beta" {
		t.Fatalf("sheet of the joined id: %q", n.Name)
	}
	if _, err := e.Creation().Join(rules.RosterID(9)); err == nil {
		t.Error("joined a number no one holds")
	}
}

// TestQuickStartWriteFails: a roster write that fails (a full disk, a
// read-only folder) is QuickStart's error, never a later "no such roster
// character" from the join that cannot find the file.
func TestQuickStartWriteFails(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Init()
	if err := os.Chmod(e.Cfg.SaveDir, 0o500); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(e.Cfg.SaveDir, 0o700)
	err := e.QuickStart()
	if !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("QuickStart on a read-only folder: %v, want the write's permission error", err)
	}
}
