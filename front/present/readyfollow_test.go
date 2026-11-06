package present

import (
	"testing"

	"gobox/front"
	"gobox/gpu"
)

// TestReadyKeepsItemHighlighted: readying moves the item (readied gear
// first); the highlight follows the item, not its old row.
func TestReadyKeepsItemHighlighted(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	note := Note{Glyph: func(byte) [8]byte { return [8]byte{} }, Font: func(byte, bool) [8]byte { return [8]byte{} }}
	mace := front.InvItem{Index: 0, Name: "Mace", Type: 1, Slot: true}
	sling := front.InvItem{Index: 1, Name: "Sling", Type: 2, Slot: true}
	v := front.View{Party: []front.Member{{Name: "A", Selected: true}}, Inventory: &front.Inventory{Pool: []front.InvItem{mace, sling}}}
	p := &Presenter{}
	p.Inventory(dst, note, v)
	p.inv.choose(false, 1)
	p.inv.removeX(front.ActReady, 1, 0)
	sling.Index, sling.Readied, mace.Index = 0, true, 1 // the rules list readied gear first
	v.Inventory = &front.Inventory{Pool: []front.InvItem{sling, mace}}
	p.Inventory(dst, note, v)
	if it := itemAt(v.Inventory, p.inv.sel); it == nil || it.Name != "Sling" {
		t.Fatalf("highlight on %+v, want the Sling", it)
	}
}
