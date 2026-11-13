package pool

import (
	"gobox/core"
	"slices"
)

// shopBrowse (extra; the original's Buy list, ovr07:0361): the inventory
// screen with the goods beside the party inventory; Buy buys one into
// the party inventory, paid from the party's gold, Sell sells.
func (e *Engine) shopBrowse() {
	e.snap.shop = true
	e.openThen(true, cont{kind: kShopBrowse})
}

// Shop = ovr07:053E: the script waits on the shop (PendShop) until the
// party leaves (Leave).
func (e *Engine) Shop() {
	e.songMusic(songShop, 0) // extra
	greeting := e.paneCopy() // extra: the merchant's greeting comes back after each redraw (the original cleared it)
	e.must(core.EnterShop{})
	e.DrawMainScreen()
	e.keepPane(greeting)
	e.must(core.StockShop{})
	e.pw.greeting = greeting
	e.wait(PendShop)
}

// browse = the shop's Buy (extra: the goods beside member n's inventory,
// n < 0 the selected one).
func (e *Engine) browse(n int) error {
	if !e.waiting(PendShop) {
		return core.ErrNotNow
	}
	e.actor(n)
	e.shopBrowse()
	return nil
}

// identify = the shop's Id (extra: the shop's, not the item menu's): item
// h of member n's (identifiable), the player agreed to the price.
func (e *Engine) identify(n int, h core.ItemHandle) error {
	if !e.waiting(PendShop) || n < 0 || n >= e.partySize() {
		return core.ErrNotNow
	}
	id := e.partyIDs()[n]
	if !slices.Contains(e.listedItems(id), h) || !slices.Contains(e.fact(id, core.IdentifiableFact{}).(core.Handles), h) || e.itemCard(id, h).Identified {
		return core.ErrNotNow
	}
	e.identifyItem(id, h)
	return nil
}

// keepPane (extra) puts the text pane's text back as it was in pane
// (paneCopy).
func (e *Engine) keepPane(pane [TextPaneRows][]textCell) {
	for i, row := range pane {
		for x, c := range row {
			if c.set {
				e.put(x, msgTop+i, int(c.colour), int(c.bg), c.c, c.upcase)
			}
		}
	}
}

// paneCopy is a copy of the text pane's text (keepPane).
func (e *Engine) paneCopy() [TextPaneRows][]textCell {
	var p [TextPaneRows][]textCell
	for i, row := range e.pg.tx.pane {
		p[i] = append([]textCell(nil), row...)
	}
	return p
}
