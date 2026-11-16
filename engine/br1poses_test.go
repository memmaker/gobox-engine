package engine

import (
	"testing"

	"gobox/rules"
)

// brActionBlock is the second frame's block offset (maptiles.pas SECOND_FRAME_RECORD; LoadIconAnim: block n and n+0x80).
const brActionBlock = 0x80

// TestBR1ActionPose: each unit of a BR fight publishes the action pose the original shows while it attacks (cbtmap.pas
// SetCreatureFacing frame 1): the icon's block+0x80, pixel for pixel, in the facing of the ready pose.
func TestBR1ActionPose(t *testing.T) {
	e, f := br1AtTurn(t, br1FarFight, br1FarFacing)
	seen := 0
	for _, u := range f.Field.Units {
		slot := int(e.fact(e.unitID(u.ID), rules.IconFact{}).(rules.Number))
		im := e.icons[slot][0]
		src := e.LoadPic(im.File, im.Block+brActionBlock)
		if src == nil {
			if u.Action.Pixels != nil {
				t.Errorf("unit %d %s: an action pose without a block", u.ID, u.Name)
			}
			continue
		}
		if u.Action.W != src.Width*8 || u.Action.H != src.Height || u.Action.Pixels == nil {
			t.Errorf("unit %d %s: action %dx%d, the block is %dx%d", u.ID, u.Name, u.Action.W, u.Action.H, src.Width*8, src.Height)
			continue
		}
		seen++
		for i, p := range src.Frame(0) {
			x := i % u.Action.W
			if u.Facing > 3 {
				x = u.Action.W - 1 - x
			}
			if (p == brMaskColour) != (int(u.Action.Pixels[i-i%u.Action.W+x]) == u.Action.Clear) {
				t.Fatalf("unit %d %s: action pixel %d differs from the block", u.ID, u.Name, i)
			}
		}
	}
	if seen == 0 {
		t.Error("no unit has an action pose")
	}
}
