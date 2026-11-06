package blobber

import "testing"

// A picture scene has no 3D view: asking it about its menu must not panic (BR2 shop click).
func TestNilViewOptions(t *testing.T) {
	var v *View
	if v.OptionsOpen() || v.Moving() {
		t.Fatal("no view has no menu and no move")
	}
}
