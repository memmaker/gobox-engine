package app

import (
	"slices"
	"testing"
)

// TestTimedKeys: \t<ms> is a timed wait (\t alone stays Tab), \S a mid-run shot.
func TestTimedKeys(t *testing.T) {
	got := ParseKeys(`q\t500\S\t\r`, nil)
	want := []Key{'q', Timed | 30, Snap, 9, 13}
	if !slices.Equal(got, want) {
		t.Fatalf("%x", got)
	}
}
