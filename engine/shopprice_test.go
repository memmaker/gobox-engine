//go:build devviews

package engine

import (
	"gobox/game"
	"gobox/rules1e"
	"testing"
)

// TestShopZeroPrice (BUG-CLASSES 68): a ware priced 0 is priced 1, as
// ovr07:0068 has it: shown so at the shop's door, before Buy, and charged.
func TestShopZeroPrice(t *testing.T) {
	e := quickStarted(t)
	e.state().BeginFlag = 4
	it := e.state().RandomItem(rules1e.ItemLongSword)
	it.Price = 0
	e.state().Treasure = []*rules1e.Item{it}
	e.state().Purse.Gold = 10
	e.Shop()
	e.Advance()
	s, _ := e.build().(game.Trading)
	if len(s.Goods) != 1 || s.Goods[0].Price != 1 {
		t.Fatalf("goods %+v, want the ware at 1", s.Goods)
	}
	if _, err := e.doNow(game.Buy{Item: e.goods()[0]}); err != nil || e.state().Purse.Gold != 9 {
		t.Fatalf("Buy: %v, gold %d, want 9", err, e.state().Purse.Gold)
	}
}
