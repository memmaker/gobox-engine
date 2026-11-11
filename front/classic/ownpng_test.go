package classic

import (
	"image"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestOwnPNGsReachTheFront: a member's own PNG portrait (Member.Dropped) is
// its card's Portrait, and its own PNG combat icon (Unit.Icon) is its
// figure on the board. Both were lost on the way to the front: the card
// showed a race/class pool face, the board the pool icon.
func TestOwnPNGsReachTheFront(t *testing.T) {
	face, icon := image.NewRGBA(image.Rect(0, 0, 32, 40)), image.NewRGBA(image.Rect(0, 0, 64, 64))
	ms := Party([]game.Member{{Name: "A", Race: "Human", Class: "Fighter", Dropped: face}}, nil, 0)
	if len(ms) != 1 || ms[0].Portrait != face {
		t.Errorf("card Portrait is not the dropped PNG")
	}
	st := play.Fighting{Common: play.Common{Party: []game.Member{{Name: "A"}}},
		Field: play.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, Member: 0, Icon: icon}}}}
	b := NewGame(&fake{st: st}, game.Data{}).View().Battle
	if b == nil || len(b.Units) != 1 || b.Units[0].Icon != icon {
		t.Fatal("the board unit's Icon is not its own PNG")
	}
}
