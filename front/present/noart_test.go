package present

import (
	"gobox/front"
	"testing"
)

// Without doll or atlas art the inventory loads and answers, no panic.
func TestNoArt(t *testing.T) {
	d, a := dollPNG, atlasPNG
	defer func() { dollPNG, atlasPNG = d, a }()
	dollPNG, atlasPNG = nil, nil
	var inv inventory
	inv.load()
	if inv.doll.Bounds().Empty() || inv.icon(&front.InvItem{}) >= 0 && inv.atlas == nil {
		t.Fatal("no-art images should be a clear pixel")
	}
}
