package pool

import (
	grules "gobox/rules"
	"slices"
	"testing"

	"gobox/game/play"
)

// TestAddChar: on the party menu with saved characters, AddCharacter{Roster: 0}
// adds the first of them (BUG-CLASSES 30); the Add picker lists them in
// BuildCharList order.
func TestAddChar(t *testing.T) {
	e := testEngine(t)
	e.Open("")
	e.titleDone()
	e.Advance()
	if _, err := e.Do(play.QuickStart{}); err != nil { // a party to save, then dropped
		t.Fatal(err)
	}
	for _, c := range e.state().Party {
		e.SaveCharacter(e.charID(c), grules.MakeFileName(c.Name()))
	}
	e.state().Party = e.state().Party[:0]
	if len(e.BuildCharList()) == 0 {
		t.Fatal("no saved characters")
	}
	n := len(e.state().Party)
	if _, err := e.Do(play.OpenPicker{Kind: PickAdd}); err != nil || e.picker == nil {
		t.Fatalf("OpenPicker PickAdd: %v", err)
	}
	if p := commonOf(e.State()).Picker; p == nil || !slices.Equal(p.Labels, e.BuildCharList()) {
		t.Fatalf("Add picker labels %v, want %v", p, e.BuildCharList())
	}
	if _, err := e.Do(play.AddCharacter{Roster: 0}); err != nil {
		t.Fatalf("AddCharacter refused: %v", err)
	}
	if len(e.state().Party) != n+1 {
		t.Fatalf("party %d, want %d", len(e.state().Party), n+1)
	}
	if _, err := e.Do(play.AddCharacter{Roster: 0}); err == nil { // BUG-CLASSES 31: added twice, nothing changes
		t.Error("the same character added again was accepted")
	}
}

// TestNoSilentNoop (BUG-CLASSES 31): a call the rules can't take is refused,
// not accepted with nothing changed.
func TestNoSilentNoop(t *testing.T) {
	e := quickStarted(t)
	for _, a := range []play.Action{play.Order{From: 0, To: 99}, play.RemoveNote{ID: 99}} {
		if _, err := e.Do(a); err == nil {
			t.Errorf("%T%+v accepted with no effect", a, a)
		}
	}
}

// TestDropMember: the party menu's Drop drops the member at once; the
// original's two questions (ovr17:1F5A, 1F93) are gobox's to ask first.
func TestDropMember(t *testing.T) {
	e := quickStarted(t)
	e.Advance() // the party menu comes up
	n := len(e.state().Party)
	if _, err := e.Do(play.DropMember{Member: 0}); err != nil || len(e.state().Party) != n-1 || commonOf(e.State()).Ask != nil {
		t.Fatalf("Drop: %v, party %d want %d, ask %+v", err, len(e.state().Party), n-1, commonOf(e.State()).Ask)
	}
}
