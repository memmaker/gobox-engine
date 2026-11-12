package pool

import (
	"gobox/game/play"
	"image"
	"rules1e"
	"testing"
)

// TestDepartureAtWalls (BUG-CLASSES 77): the pre-step script (events[0])
// runs ahead of every step command, into a wall too (the move refuses it
// after), so Departure says so for every side.
func TestDepartureAtWalls(t *testing.T) {
	e := quickStarted(t)
	e.set = 8
	e.LoadGeo(13) // Kobold Caves, as TestMapEdgeNoWrap
	e.gs.areaKind, e.viewMode = 13, 1
	m := walkMap{e}
	n := 0
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			for side := 0; side < 8; side += 2 {
				if e.WallAt(x, y, side) != 0 {
					n++
					if !m.Departure(image.Pt(x, y), side) {
						t.Fatalf("Departure false at %d,%d side %d, a wall", x, y, side)
					}
				}
			}
		}
	}
	if n == 0 {
		t.Fatal("no walls in the Kobold Caves")
	}
}

// TestShopPriceIsGold (BUG-CLASSES 78): each ware's shown Price is the gold
// Buy takes (Item.Price is gold; only the sell value is copper, Sale).
func TestShopPriceIsGold(t *testing.T) {
	for _, typ := range []rules1e.ItemType{rules1e.ItemLongSword, rules1e.ItemPlateMail, rules1e.ItemArrow, rules1e.ItemPotion} {
		e := quickStarted(t)
		e.state().BeginFlag = 4
		e.state().Treasure = []*rules1e.Item{e.state().RandomItem(typ)}
		e.state().Purse.Gold = 1 << 20
		e.Shop()
		e.Advance()
		shown := e.State().(play.Trading).Goods[0].Price
		gold := e.state().Purse.Gold
		if _, err := e.Do(play.Buy{Item: e.goods()[0]}); err != nil {
			t.Fatalf("%x: Buy: %v", typ, err)
		}
		if paid := int(gold - e.state().Purse.Gold); paid != shown || shown <= 0 {
			t.Errorf("%x: shown %d, charged %d gold", typ, shown, paid)
		}
	}
}
