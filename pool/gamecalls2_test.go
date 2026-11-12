package pool

import (
	"testing"

	"gobox/game/play"
)

// AddNote, RemoveNote and Order through Do.
func TestNoteOrderDecisions(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	e.Advance()
	if _, err := e.Do(play.AddNote{Text: "a clue"}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Do(play.AddNote{Text: "a clue"}); err == nil || err.Error() != "Already in the journal" {
		t.Errorf("duplicate note: %v", err)
	}
	n := len(e.ext.notes)
	e.Do(play.RemoveNote{ID: n - 1})
	if len(e.ext.notes) != n-1 {
		t.Error("note not removed")
	}
	if _, err := e.Do(play.MakeCamp{}); err != nil {
		t.Fatal(err)
	}
	a, b := e.state().Party[0].Name(), e.state().Party[1].Name()
	if _, err := e.Do(play.Order{From: 0, To: 1}); err != nil {
		t.Fatal(err)
	}
	if e.state().Party[0].Name() != b || e.state().Party[1].Name() != a {
		t.Errorf("order: %s, %s", e.state().Party[0].Name(), e.state().Party[1].Name())
	}
}
