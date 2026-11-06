package present

import (
	"image"
	"image/color"
	"testing"

	"gobox/front"
)

// TestDaxIconSize: a big monster's DAX icon (48x24) is a 2x1 figure, its size
// the decoded picture's, never a one-square card.
func TestDaxIconSize(t *testing.T) {
	ref := front.Picture{File: "CBODY.DAX", Block: 77}
	daxPics[daxKey{ref.File, ref.Block, ref.Frame, refMask(ref)}] = image.NewRGBA(image.Rect(0, 0, 48, 24))
	defer delete(daxPics, daxKey{ref.File, ref.Block, ref.Frame, refMask(ref)})
	b := &front.Battle{Units: []front.Unit{{ID: 1}}, IconPix: []front.Picture{ref}}
	if got := pixBattle(b, [16]color.RGBA{}).Units[0].IconSize; got != image.Pt(48, 24) {
		t.Fatalf("IconSize %v, want 48x24", got)
	}
}
