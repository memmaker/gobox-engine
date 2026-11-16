package engine

import (
	"testing"

	"gobox/game"
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

// TestBR1BigMonsterPicture: a monster whose block is larger than a cell (CPIC4 block 4 is 24x48, block 5 is 48x24: the original
// loads the block as it is and blits it from the unit's first cell, maptiles.pas DrawIconAnimSprite) publishes the block's own size.
func TestBR1BigMonsterPicture(t *testing.T) {
	e, f := br1AtTurn(t, br1FarFight, br1FarFacing)
	var foe game.Unit
	for _, u := range f.Field.Units {
		if u.Foe {
			foe = u
		}
	}
	slot := int(e.fact(e.unitID(foe.ID), rules.IconFact{}).(rules.Number))
	for _, c := range []struct{ block, w, h int }{{4, 24, 48}, {5, 48, 24}} {
		e.icons[slot][0] = e.LoadPicMasked("CPIC4", c.block, brMaskColour)
		if e.icons[slot][0] == nil {
			t.Fatalf("no CPIC4 block %d", c.block)
		}
		var got game.Unit
		for _, u := range e.build().(game.Fighting).Field.Units {
			if u.ID == foe.ID {
				got = u
			}
		}
		if got.Art.W != c.w || got.Art.H != c.h || len(got.Art.Pixels) != c.w*c.h {
			t.Errorf("CPIC4 block %d: published %dx%d (%d pixels), want %dx%d", c.block, got.Art.W, got.Art.H, len(got.Art.Pixels), c.w, c.h)
		}
	}
}

// TestBR1OverlandMark: on the overland drawn as one picture the party's square is marked by CURSOR block 1 masked on 13, at the
// square's own 8x8 cell of the picture (mapcurs.pas DrawCursor draws at wildernessX + 1 on a picture shown at column 1).
func TestBR1OverlandMark(t *testing.T) {
	e := br1Begun(t)
	br1OwnBlock(t, e, 201, "    AreaKind(4, 255)")
	br1Pages(t, e)
	if m := commonOf(e.build()).Mark; m != nil {
		t.Fatalf("a mark without a whole picture: %+v", m)
	}
	e.es.bigPic = e.LoadPic("CURSOR", 1) // any whole picture
	e.world.MapX, e.world.MapY = 10, 5
	m := commonOf(e.build()).Mark
	src := e.LoadPic("CURSOR", 1)
	if m == nil || src == nil || m.At != (rules.MapPos{X: 10, Y: 5}) || m.Picture.W != src.Width*8 || m.Picture.H != src.Height || m.Picture.Clear < 0 {
		t.Fatalf("the mark: %v (block %v)", m != nil, src != nil)
	}
	for i, p := range src.Frame(0) {
		if (p == 13) != (int(m.Picture.Pixels[i]) == m.Picture.Clear) {
			t.Fatalf("mark pixel %d differs from the block", i)
		}
	}
}
