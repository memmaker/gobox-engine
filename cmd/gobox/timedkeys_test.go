package main

import (
	"gobox/ui"
	"slices"
	"testing"
)

// TestTimedKeys: \t<ms> is a timed wait (\t alone stays Tab), \S a mid-run shot.
func TestTimedKeys(t *testing.T) {
	got := parseKeys(`q\t500\S\t\r`, nil)
	want := []ui.Key{'q', ui.ScriptTimed | 30, ui.ScriptSnap, 9, 13}
	if !slices.Equal(got, want) {
		t.Fatalf("%x", got)
	}
}
