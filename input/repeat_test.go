package input

import "testing"

func TestRepeat(t *testing.T) {
	var got []int
	for d := 0; d <= 40; d++ {
		if Repeat(d, 20, 9) {
			got = append(got, d)
		}
	}
	if want := []int{1, 20, 29, 38}; len(got) != len(want) || got[0] != 1 || got[1] != 20 || got[2] != 29 || got[3] != 38 {
		t.Fatal(got)
	}
	if Repeat(5, 0, 0) {
		t.Fatal("rate 0 must not repeat")
	}
}
