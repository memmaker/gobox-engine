package ui

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"gobox/game"
)

// pngOf is im as PNG bytes.
func pngOf(t *testing.T, im image.Image) []byte {
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// TestOwnPNGsReachTheFront: a member's own PNG portrait (Member.Dropped) is
// its card's Portrait, and its own PNG combat icon (Unit.IconPNG) is its
// figure on the board, in the pose the blow shows. Both were lost on the way to the front: the card
// showed a race/class pool face, the board the pool icon.
func TestOwnPNGsReachTheFront(t *testing.T) {
	face, icon := pngOf(t, image.NewRGBA(image.Rect(0, 0, 32, 40))), pngOf(t, image.NewRGBA(image.Rect(0, 0, 128, 64)))
	ms := partyOf(game.Exploring{Common: game.Common{Party: []game.Member{{Name: "A", Race: "Human", Class: "Fighter", Dropped: face}}}})
	if len(ms) != 1 || ownPNG(ms[0].Dropped) != ownPNG(face) {
		t.Errorf("card Portrait is not the dropped PNG")
	}
	st := game.Fighting{Common: game.Common{Party: []game.Member{{Name: "A"}}},
		Field: game.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, Member: 0, IconPNG: icon, Pose: 1}}}}
	b := liveBoard(NewGame(&fake{st: st}, game.Data{}))
	if b == nil || len(b.Units) != 1 || b.Units[0].Icon != ownPose(icon, 1, false) {
		t.Fatal("the board unit in its action pose does not show its own PNG's action icon")
	}
}

// A square PNG is both poses; one twice as wide is cut in the middle, the
// left half ready, the right action, mirrored to face left; each cut is
// made once.
func TestOwnPose(t *testing.T) {
	wide := image.NewRGBA(image.Rect(0, 0, 4, 2))
	wide.Set(0, 0, color.White) // ready: top left
	wide.Set(3, 1, color.White) // action: bottom right
	b := pngOf(t, wide)
	at := func(im image.Image, x, y int) bool { _, _, _, a := im.At(x, y).RGBA(); return a != 0 }
	if r := ownPose(b, 0, false); r.Bounds().Dx() != 2 || !at(r, 0, 0) || at(r, 1, 1) {
		t.Fatal("ready is not the left half")
	}
	if a := ownPose(b, 1, false); !at(a, 1, 1) || at(a, 0, 0) {
		t.Fatal("action is not the right half")
	}
	if a := ownPose(b, 1, true); !at(a, 0, 1) {
		t.Fatal("the flip did not mirror")
	}
	sq := pngOf(t, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	if ownPose(sq, 0, false) != ownPose(sq, 1, false) || ownPose(nil, 0, false) != nil {
		t.Fatal("a square is not both poses, cached")
	}
}
