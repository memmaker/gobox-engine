package engine

import "testing"

// TestCodecCotabSelStatus: the read after a select that found nobody is 0, once.
func TestCodecCotabSelStatus(t *testing.T) {
	e, _ := selStatusEngine(selStatusState{okay: 1})
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

// TestCodecCotabSelFlag1: the read keeps bit 0 (SpecialReadZ1 $E4: scriptFlags[0] and 1), the write is ignored.
func TestCodecCotabSelFlag1(t *testing.T) {
	c := cotabCodecs["CotabSelFlag1"]
	if v := c.toGame(&Engine{}, 3); v != 1 {
		t.Errorf("flags 3 read %d, want 1", v)
	}
	if v := c.toGame(&Engine{}, 2); v != 0 {
		t.Errorf("flags 2 read %d, want 0", v)
	}
	if !c.ignoreWrites {
		t.Error("a write is stored")
	}
}
