package ui

import "testing"

// TestWalk: the free camera stops at a closed side, asks for the step
// through an open one, and backs off a refused one.
func TestWalk(t *testing.T) {
	v := &View{g: &BlobberGame{}}
	at := Camera{5, 5, 0} // facing north
	north := FreeInput{Forward: 1}
	closed, open := [4]bool{}, [4]bool{true, false, false, false}
	for range 120 {
		if c := v.Walk(at, closed, north, true); c != at {
			t.Fatalf("closed: asked for %v", c)
		}
	}
	if z := v.g.free.z; z < 5+freeRadius-1e-9 {
		t.Fatalf("closed: through the wall, z %v", z)
	}
	var c Camera
	for range 60 {
		if c = v.Walk(at, open, north, true); c != at {
			break
		}
	}
	if c != (Camera{5, 4, 0}) {
		t.Fatalf("open: asked for %v", c)
	}
	v.Walk(at, open, north, true) // refused: the host stays at 5,5
	if z := v.g.free.z; z < 5 {
		t.Fatalf("refused: still out at z %v", z)
	}
	if c := v.Walk(Camera{5, 4, 0}, closed, FreeInput{}, true); c != (Camera{5, 4, 0}) { // taken
		t.Fatalf("taken: asks for %v", c)
	}
}
