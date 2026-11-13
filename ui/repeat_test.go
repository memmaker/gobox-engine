package ui

import "testing"

func TestRepeat(t *testing.T) {
	var got []int
	for d := 0; d <= 40; d++ {
		if Repeat(d, 60) {
			got = append(got, d)
		}
	}
	if want := []int{1, 18, 24, 30, 36}; len(got) != len(want) || got[1] != 18 || got[2] != 24 || got[4] != 36 {
		t.Fatal(got)
	}
	if Repeat(0, 60) || Repeat(2, 120) {
		t.Fatal("not held or too early must not fire")
	}
}
