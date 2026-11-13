package ui

import (
	"gobox/ui/blobber"
	"testing"
)

// A picture scene has no 3D view: asking it about its menu must not panic (BR2 shop click).
func TestNilViewOptions(t *testing.T) {
	var v *blobber.View
	if (&Presenter{view: v}).optionsOpen() || v.Moving() {
		t.Fatal("no view has no menu and no move")
	}
}
