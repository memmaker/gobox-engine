//go:build devviews

package engine

import "testing"

// TestDebugStarts (BUG-CLASSES 15): every debug start the host passes goes
// through Engine.Open and lands where it says: -explore adventuring with a
// party, -dev's DevAction at the first prompt, -slot in that save.
func TestDebugStarts(t *testing.T) {
	e := testEngine(t)
	e.Explore = true
	ran := false
	e.DevAction = func() { ran = true }
	e.Open("")
	e.titleDone()
	e.Advance()
	if !e.exploring() || len(e.state().Party) == 0 || !ran {
		t.Fatalf("-explore -dev: exploring %v, party %d, dev ran %v", e.exploring(), len(e.state().Party), ran)
	}
	e.posX ^= 1 // a square a Quick Start would not give
	x, y := e.posX, e.posY
	if err := e.SaveNamed("T"); err != nil {
		t.Fatal(err)
	}
	l := testEngine(t)
	l.Cfg.SaveDir = e.Cfg.SaveDir
	l.Explore = true
	ran = false
	l.DevAction = func() { ran = true }
	l.Open("T")
	l.titleDone()
	l.Advance()
	if !l.exploring() || !ran || l.posX != x || l.posY != y {
		t.Fatalf("-slot: exploring %v, dev ran %v, at %d,%d want %d,%d", l.exploring(), ran, l.posX, l.posY, x, y)
	}
}
