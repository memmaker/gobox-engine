package present

import (
	"gobox/front"
	"testing"
)

// Without doll or atlas art the inventory loads and answers, no panic.
func TestNoArt(t *testing.T) {
	d := dollPNG
	defer func() { dollPNG = d }()
	dollPNG = nil
	var inv inventory
	inv.load()
	if inv.doll.Bounds().Empty() || inv.iconImage(&front.InvItem{Type: 0xFF}) != nil {
		t.Fatal("no-art images should be a clear pixel")
	}
}
