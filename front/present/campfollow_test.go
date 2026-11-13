package present

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// shopState is a shop up for one member, the party allowed these actions.
func shopState(allowed ...play.Action) play.Trading {
	return play.Trading{Common: play.Common{Party: []game.Member{{Name: "A"}}, Allowed: play.Allowed{Plain: allowed}}}
}

// TestShopRowAppears: the shop's Take appears above Pool once credits are
// pooled; the highlight stays on the row it was on (Pool), so Down, Down,
// Enter from there picks what is drawn (Exit), not a row shifted by one.
func TestShopRowAppears(t *testing.T) {
	g, f := playGame(shopState(play.PoolCoins{}, play.LootShare{}))
	p := &Presenter{}
	p.SetEngine(g)
	v := front.View{Shop: &front.Shop{}}
	p.campKey('P', v) // Pool: Buy View Pool Share Exit, the highlight on Pool
	p.inv.acts = nil
	f.st = shopState(play.Take{}, play.PoolCoins{}, play.LootShare{}) // pooled: Take comes in above
	for _, k := range []uint16{keyDown, keyDown, '\r'} {
		p.campKey(k, v)
	}
	if a := p.inv.acts; len(a) != 1 || a[0].Kind != front.ActClose {
		t.Fatalf("Pool, then Down Down Enter with Take in: %+v, want Exit (ActClose)", a)
	}
}
