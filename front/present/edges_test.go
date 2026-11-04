package present

import "testing"

// An edge sequence keeps its ends and repeats its middle (the cracked borders).
func TestEdgeCell(t *testing.T) {
	var got []int
	for i := range 7 {
		got = append(got, edgeCell(i, 7, 4))
	}
	want := []int{0, 1, 2, 1, 2, 1, 3}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	if edgeCell(39, 40, 40) != 39 || edgeCell(13, 40, 40) != 13 {
		t.Fatal("a full-length edge must take its table as is")
	}
}
