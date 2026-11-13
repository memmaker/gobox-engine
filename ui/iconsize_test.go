package ui

import (
	"image"
	"testing"

	"gobox/game"
)

// TestDaxIconSize: a big monster's DAX icon (48x24) is a 2x1 figure, its size
// the decoded picture's, never a one-square card.
func TestDaxIconSize(t *testing.T) {
	ref := Picture{File: "CBODY.DAX", Block: 77}
	daxPics[daxKey{ref.File, ref.Block, ref.Frame, refMask(ref)}] = image.NewRGBA(image.Rect(0, 0, 48, 24))
	defer delete(daxPics, daxKey{ref.File, ref.Block, ref.Frame, refMask(ref)})
	c := &game.Combat{Units: []game.Unit{{ID: 1, Member: -1, Foe: true, Art: game.Picture{File: ref.File, Block: ref.Block}}}}
	if got := boardIn(c, nil).Units[0].IconSize; got != image.Pt(48, 24) {
		t.Fatalf("IconSize %v, want 48x24", got)
	}
}
