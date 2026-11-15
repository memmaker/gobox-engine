package engine

import (
	"gobox/game"
	"gobox/rules"
	"gobox/rules1e"
	"image"
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
	departureAtWalls(t, e)
}

// TestDepartureAtWallsCOTAB: the same on cotab's imported folder, in every dungeon map
// its data holds.
func TestDepartureAtWallsCOTAB(t *testing.T) {
	e := quickStartedOn(t, advEngine(t, advEnvCOTAB))
	n := 0
	for s := rules.Set(1); s <= sets; s++ {
		for id := range 256 {
			if len(e.gdata().geoBlock(s.File("GEO"), id)) < geoSize {
				continue
			}
			e.set = s
			e.LoadGeo(id)
			e.gs.areaKind, e.viewMode = byte(id), 1
			departureAtWalls(t, e)
			n++
		}
	}
	if n == 0 {
		t.Fatal("no dungeon maps")
	}
}

// departureAtWalls is TestDepartureAtWalls on e's loaded map.
func departureAtWalls(t *testing.T, e *Engine) {
	t.Helper()
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
		t.Fatalf("no walls in geo %d", int(e.world.Geo))
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
		shown := e.build().(game.Trading).Goods[0].Price
		gold := e.state().Purse.Gold
		if _, err := e.doNow(game.Buy{Item: e.goods()[0]}); err != nil {
			t.Fatalf("%x: Buy: %v", typ, err)
		}
		if paid := int(gold - e.state().Purse.Gold); paid != shown || shown <= 0 {
			t.Errorf("%x: shown %d, charged %d gold", typ, shown, paid)
		}
	}
}
