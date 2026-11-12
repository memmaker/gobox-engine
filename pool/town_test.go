package pool

import (
	"testing"

	"gobox/ecl/script"
	"gobox/game/play"
	"rules1e"
)

// TestShopPriceFactor: the shops' factor over 16 (cotab/sotsb/pod's table of 1 2 4 8 32 64 128);
// the normal 16 and any other leave the price.
func TestShopPriceFactor(t *testing.T) {
	for _, c := range []struct{ price, factor, want int32 }{
		{160, 1, 10}, {160, 2, 20}, {160, 4, 40}, {160, 8, 80}, {160, 16, 160}, {160, 32, 320}, {160, 64, 640}, {160, 128, 1280}, {160, 0, 160}, {160, 5, 160},
	} {
		if got := scalePrice(c.price, c.factor); got != c.want {
			t.Errorf("price %d factor %d: %d, want %d", c.price, c.factor, got, c.want)
		}
	}
	if e := testEngine(t); e.shopPrice(100, 5) != 100 { // POR's tables name no factor
		t.Error("POR's shop scaled")
	}
}

// TestShopBundlesAndWares: with a price factor in the layout and a game whose shops sell bundles at
// their count times the price, the price is factor then count (POR: the plain price); the wares the
// shop and the inventory list (Trading.Goods) ask what Buy takes.
func TestShopBundlesAndWares(t *testing.T) {
	e := quickStarted(t)
	e.vm.Script.Layout = &script.Layout{Fields: []script.Field{{Name: fieldPriceScale, Owner: script.OwnerGame, Width: script.WidthByte, From: 39, Len: 1}}}
	r, _ := e.layoutField(fieldPriceScale)
	e.host.Set(r, 32) // twice the price
	if got := e.shopPrice(100, 3); got != 200 {
		t.Errorf("no bundles: %d, want 200", got)
	}
	tab := *e.state().T()
	tab.Town.Bundles = true
	e.state().Tables = &tab
	if got := e.shopPrice(100, 3); got != 600 {
		t.Errorf("bundle of 3: %d, want 600", got)
	}
	if got := e.shopPrice(100, 1); got != 200 {
		t.Errorf("one piece: %d, want 200", got)
	}
	e.state().Treasure = []*rules1e.Item{{Price: 100, Count: 3}}
	e.push(cont{kind: kNop})
	e.Shop()
	e.Advance()
	tr, ok := e.State().(play.Trading)
	if !ok || len(tr.Goods) != 1 || tr.Goods[0].Price != 600 {
		t.Fatalf("wares %+v (%T): want one at 600", tr.Goods, e.State())
	}
}
