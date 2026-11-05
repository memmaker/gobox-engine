package present

import (
	"image"
	"testing"

	"gobox/front"
)

// A game without a paper doll (no DollBoxes) still sends its readied
// items in Inventory.Doll: the member's box draws without them, no panic.
func TestDrawMemberNoDoll(t *testing.T) {
	old := dollBoxes
	defer func() { dollBoxes = old }()
	dollBoxes = nil
	var p Presenter
	p.inv.px, p.inv.doll = image.NewRGBA(image.Rect(0, 0, 200, 200)), image.NewRGBA(image.Rect(0, 0, 10, 10))
	in := &front.Inventory{Member: -1, Doll: []*front.InvItem{{Index: 1, Name: "Sword"}}}
	p.drawMember(in, front.View{Inventory: in}, 0, 100, 150, false)
}
