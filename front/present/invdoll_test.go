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
	in := &front.Inventory{Member: -1, Doll: []*front.InvItem{{Item: 2, Name: "Sword"}}}
	p.drawMember(in, 0, 100, 150, false)
}

// The paper doll is there without SetData: NewFront (gameData) is the
// only data path most games take, and the doll pane was blank.
func TestDollDefaults(t *testing.T) {
	if len(dollBoxes) == 0 || len(dollPNG) == 0 || len(slotNames) == 0 {
		t.Fatal("no default doll")
	}
}

func TestReadyMarkDefault(t *testing.T) {
	if readyMark(&front.InvItem{Readied: true}) == "" {
		t.Fatal("a readied item has no mark")
	}
}
