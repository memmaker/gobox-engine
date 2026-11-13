package engine

import "testing"

// TestCodecCotabSelStatus: the read after a select that found nobody is 0, once.
func TestCodecCotabSelStatus(t *testing.T) {
	e := &Engine{}
	c := cotabCodecs["CotabSelStatus"]
	if v := c.toGame(e, 1); v != 1 {
		t.Fatalf("found: %d", v)
	}
	e.es.memberMissing = true
	if v := c.toGame(e, 1); v != 0 {
		t.Fatalf("not found: %d", v)
	}
	if v := c.toGame(e, 1); v != 1 {
		t.Fatalf("second read: %d", v)
	}
}
