package ui

import (
	"testing"
)

// An edge sequence keeps its ends and repeats its middle (the cracked borders).
func TestEdgeCell(t *testing.T) {
	var got []int
	for i := range 7 {
		got = append(got, EdgeCell(i, 7, 4))
	}
	want := []int{0, 1, 2, 1, 2, 1, 3}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	if EdgeCell(39, 40, 40) != 39 || EdgeCell(13, 40, 40) != 13 {
		t.Fatal("a full-length edge must take its table as is")
	}
	if EdgeCell(17, 20, 40) != 37 || EdgeCell(2, 20, 40) != 2 {
		t.Fatal("a short edge must keep its table's head and tail")
	}
}
