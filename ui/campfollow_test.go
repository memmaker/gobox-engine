package ui

import (
	"testing"

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
	p.campKey('P') // Pool: Buy View Pool Share Exit, the highlight on Pool
	p.inv.acts = nil
	f.st = shopState(play.Take{}, play.PoolCoins{}, play.LootShare{}) // pooled: Take comes in above
	for _, k := range []uint16{keyDown, keyDown, '\r'} {
		p.campKey(k)
	}
	f.sent = nil
	for _, c := range p.Calls() {
		c(g)
	}
	if len(f.sent) != 1 || f.sent[0] != (play.Leave{}) {
		t.Fatalf("Pool, then Down Down Enter with Take in: %+v, want Exit (Leave)", f.sent)
	}
}
