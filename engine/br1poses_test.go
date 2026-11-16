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

// TestBR1MarkPictures: each marker the ruleset lays (arena values 1..5) has the picture of the original's icon row (cbtmap.pas
// DrawArenaCell: iconFrames[(tile shr 1) + 33, tile and 1], tile = value - 1), COMSPR block row-12 (and +0x80 for frame 1).
func TestBR1MarkPictures(t *testing.T) {
	e, _ := br1AtTurn(t, br1FarFight, br1FarFacing)
	const firstMarkRow = 33 // LoadIconAnim loads COMSPR block n into row n + 12
	for m := rules.SquareMark(1); m <= 5; m++ {
		tile := int(m) - 1
		row, frame := firstMarkRow+tile>>1, tile&1
		block := row - 12
		if frame == 1 {
			block += brActionBlock
		}
		src := e.LoadPic("COMSPR", row-12)
		if frame == 1 {
			src = e.LoadPic("COMSPR", row-12+brActionBlock)
		}
		got := e.markPicture(m)
		if src == nil || got.Pixels == nil || got.W != src.Width*8 || got.H != src.Height {
			t.Fatalf("mark %d: picture %dx%d (pixels %v), COMSPR block %d: %v", m, got.W, got.H, got.Pixels != nil, block, src != nil)
		}
		for i, p := range src.Frame(0) {
			if (p == brMaskColour) != (int(got.Pixels[i]) == got.Clear) {
				t.Fatalf("mark %d: pixel %d differs from the block", m, i)
			}
		}
	}
	if e.markPicture(0).Pixels != nil || e.markPicture(9).Pixels != nil {
		t.Error("a marker the data does not name has a picture")
	}
}
