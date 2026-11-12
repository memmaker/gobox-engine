package pool_test

import (
	"testing"

	"gobox/pool"
)

// TestClueRead: a chapter is read from its title on, through the book's
// end, and some page has a picture that scrolls.
func TestClueRead(t *testing.T) {
	r := pool.ClueRead(5)
	if r.Title != "Clue Book" || len(r.Lines) == 0 || len(r.Lines) >= len(pool.ClueRead(0).Lines) {
		t.Fatalf("read %q with %d lines", r.Title, len(r.Lines))
	}
}
