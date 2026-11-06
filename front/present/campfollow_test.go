package present

import (
	"testing"

	"gobox/front"
)

// TestShopRowAppears: the shop's Take appears above Pool once credits are
// pooled; the highlight stays on the row it was on (Pool), so Down, Down,
// Enter from there picks what is drawn (Exit), not a row shifted by one.
func TestShopRowAppears(t *testing.T) {
	v := front.View{Shop: &front.Shop{Pool: true, Share: true}, Party: []front.Member{{Name: "A", Selected: true}}}
	p := &Presenter{}
	p.campKey('P', v) // Pool: Buy View Pool Share Exit, the highlight on Pool
	p.inv.acts = nil
	v.Shop = &front.Shop{Take: true, Pool: true, Share: true} // pooled: Take comes in above
	for _, k := range []uint16{keyDown, keyDown, '\r'} {
		p.campKey(k, v)
	}
	if a := p.inv.acts; len(a) != 1 || a[0].Kind != front.ActClose {
		t.Fatalf("Pool, then Down Down Enter with Take in: %+v, want Exit (ActClose)", a)
	}
}
