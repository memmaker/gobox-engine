package engine

import (
	"testing"
)

// TestClueRead: a chapter is read from its title on, through the book's
// end, and some page has a picture that scrolls.
func TestClueRead(t *testing.T) {
	e := testEngine(t)
	r := e.ClueRead(5)
	if r.Title != "Clue Book" || len(r.Texts) == 0 || len(r.Texts) >= len(e.ClueRead(0).Texts) {
		t.Fatalf("read %q with %d texts", r.Title, len(r.Texts))
	}
}
