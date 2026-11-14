// inventory.go holds the inventory screen's state, its picked-up stack and the decisions it sends.

package ui

import (
	"fmt"
	"gobox/game"
	"gobox/rules"
	"image"
	"image/color"
	"slices"
	"strings"
	"time"

	"gobox/asset"
	"gobox/gpu"
)

// The inventory screen (F1-F8, Items): the member's Eye of the Beholder
// paper doll on the left (the shop's Browse: its goods, half the screen
// each), the party inventory on the right, in the
// original's font, colours and box borders. Everything is drawn 1:1 on one
// canvas that is scaled whole to the window, so all pixels are the same
// size. It draws what the state puts up (inventoryOf) and acts through front.Action: the
// menu row, the tooltips and the Drop and Sell questions are its own.

// Layout, in canvas pixels. The canvas is at least invW x invH; the list
// takes what the window gives beyond that.
const (
	invW, invH = PageW, PageH
	leftCells  = 22                    // the doll box's width in cells, with its border
	rowH       = 20                    // a party inventory line: a whole icon high
	listTop    = 24                    // the first line, under the box's title
	atlasCols  = 24                    // the item atlas's cells a row; a cell is the atlas's width / atlasCols
	iconSlot   = 20                    // an item icon's slot on the 2D screens
	wheelPause = 60 * time.Millisecond // the quickest the wheel scrolls a line
)

func col(i int) color.RGBA {
	if i < 0 { // a role of none (Palette's -1): transparent
		return color.RGBA{}
	}
	c := asset.EGAPalette[i]
	return color.RGBA{c[0], c[1], c[2], 255}
}

var green, white, cyan, gray, dark, red color.RGBA // the roles' colours (applyTheme)

// inventory is the screen's state between frames.
type inventory struct {
	doll, atlas image.Image
	icons       map[rules.ItemID]int // items.rec: base kind → atlas cell
	px          *image.RGBA          // the canvas, drawn on the CPU
	img, tip    *gpu.Image           // the canvas and the tooltip on the GPU
	glyph       func(byte) [8]byte

	shown bool             // the last frame was the inventory
	sel   rules.ItemHandle // the highlighted item, NoItem for none
	keep  struct {
		row, n int
		i      rules.ItemHandle
		id     string
	} // a sold, dropped or readied item's line, the list's length and its Index then: the line after it is highlighted once it is gone (row -1 none)
	ask     *Action           // a Drop or Sell waits for Yes
	askSel  int               // the question's highlighted answer: 0 Yes, 1 No
	askAt   []image.Rectangle // the question's Yes and No on the screen, as last drawn
	up      *invUp            // what the screen shows, read from the state when it was last drawn (nil: not up)
	acts    []Action          // to send (Presenter.Actions)
	words   []word            // the menu row's words
	focus   int               // 1+ the word the keys are on (PgUp/PgDn, the pad's shoulders), 0 none
	k       int               // canvas scale, whole: what the pixel art is drawn at
	kx, ky  float64           // the canvas on screen: k and the stretch to the window's edges (fit)
	w, h    int               // canvas size
	list    pane              // the party inventory
	wares   pane              // the shop's goods (Browse)
	ware    rules.ItemHandle  // the highlighted ware, NoItem for none
	lines   []hit             // the items' rectangles (canvas)
	dollAt  image.Rectangle
	face    image.Rectangle // the portrait on the canvas, when facePNG is drawn over it at the window's resolution
	facePNG image.Image
	panelAt image.Rectangle // the last panel on the canvas, drawn again over facePNG
	hover   image.Point     // the pointer (canvas)
	mouseAt image.Point     // the pointer (screen), as last seen
	member  int             // the member whose items are up
	count   int             // party members

	carry     *carried    // the stack on the cursor (a click picked it up), nil for none
	held      bool        // the left button was down last frame
	press     *hit        // the item the button went down on
	pressAt   image.Point // where
	dragging  bool        // it moved off: a drag
	scrolling *pane       // the button went down on its scroll bar
	wheel     float64     // wheel movement not yet a line
	wheeled   time.Time   // the last line the wheel scrolled
}

// carried is a stack picked up on the cursor: n of the item (by type and
// name words, so it is found again as its count and Index change), from
// the doll or the party inventory.
type carried struct {
	typ   byte
	words []string
	doll  bool
	n     int
	icon  game.Item // as it looked when picked up (the icon under the pointer)
}

// carriedItem is the carried stack's item in the inventory shown, nil when it is gone.
func (inv *inventory) carriedItem() *game.Item {
	c := inv.carry
	if c == nil {
		return nil
	}
	same := func(it *game.Item) bool {
		return it != nil && byte(it.Type) == c.typ && slices.Equal(it.Words, c.words)
	}
	if c.doll {
		for _, it := range inv.up.doll {
			if same(it) {
				return it
			}
		}
		return nil
	}
	for k := range inv.up.pool {
		if same(&inv.up.pool[k]) {
			return &inv.up.pool[k]
		}
	}
	return nil
}

// pickUp puts n of the stack h on the cursor (0: all of it).
func (inv *inventory) pickUp(h *hit, n int) {
	all := max(h.item.Count, 1)
	if n <= 0 || n > all {
		n = all
	}
	inv.choose(false, h.item.Handle)
	inv.carry = &carried{typ: byte(h.item.Type), words: h.item.Words, doll: h.doll, n: n, icon: h.item}
}

// put drops n of the carried stack where the pointer is: onto the doll
// from the party inventory readies them, into the party inventory from
// the doll unreadies them (ActReady with the count), back where they came
// from puts them back; what is left stays on the cursor.
func (inv *inventory) put(n int) {
	c, it := inv.carry, inv.carriedItem()
	if it == nil {
		inv.carry = nil
		return
	}
	n = max(1, min(n, c.n))
	if onDoll := inv.hover.In(inv.dollAt); onDoll != (c.doll || it.Readied) { // a readied item from the list is on the doll already
		x := n
		if n >= max(it.Count, 1) {
			x = 0 // the whole stack
		}
		inv.removeX(ActReady, it.Handle, x)
	}
	if c.n -= n; c.n <= 0 {
		inv.carry = nil
	}
}

type hit struct {
	r     image.Rectangle
	item  game.Item
	doll  bool // on the doll (readied)
	ware  bool // one of the shop's goods
	whole bool // its name shows whole
}

// InventoryShown: the last frame was the inventory screen (its mouse
// input goes to InventoryMouse, not the game's text cells).
func (p *Presenter) InventoryShown() bool { return p.inv.shown }

// Screens names the presenter's own screens up, for the input latch: it
// changes whenever one opens or closes.
func (p *Presenter) Screens() string {
	return fmt.Sprint(p.inv.shown, p.board != nil, p.Options(), p.overview, p.pager != nil, p.fullMap != nil, p.clue != nil, p.training, p.frontScreens(), p.stateScreens())
}

// call queues a rule call on the engine (Calls).
func (p *Presenter) call(f func(*classicGame)) { p.calls = append(p.calls, f) }

// Actions hands over the actions the screens made since the last call.
func (p *Presenter) Actions() []Action {
	for _, a := range p.battle.Actions() {
		a := p.askAction(a) // the turn's are rule calls (pace.go)
		p.sendA(a)
	}
	var as []Action
	for _, a := range p.inv.acts { // the screens' own (the inventory's, a list's) meet the presenter's questions and rule calls too
		if !p.slotAct(a) {
			as = append(as, a)
		}
	}
	p.inv.acts = nil
	return as
}

// do runs a menu choice on the highlighted item.
func (inv *inventory) do(k ActKind) {
	it := inv.itemAt(inv.sel)
	switch {
	case k == ActNone:
	case k == ActClose, k == ActSheet, k == ActTakeAll, k >= ActSpells && k <= ActCure, k >= ActTakeMoney && k <= ActLootStatus:
		inv.acts = append(inv.acts, Action{Kind: k})
	case k == ActBuy:
		if w := inv.wareAt(inv.ware); w != nil {
			inv.acts = append(inv.acts, Action{Kind: k, Item: w.Handle})
		}
	case it == nil:
	case inv.ask == nil && (k == ActDrop && inv.up.kind != invLoot || k == ActSell): // leaving an item in the treasure (Deposit) is no loss: no question
		yes := confirmOf(itemConfirm(k, *it), game.Confirm{Asks: []string{""}, Yes: k == ActSell}).Yes // the game's (Data.Confirms), else a sale on Yes, as the Gold Box's, a drop (gone forever) on No
		inv.ask, inv.askSel = &Action{Kind: k, Item: it.Handle}, map[bool]int{true: 0, false: 1}[yes]
		return
	default:
		inv.remove(k, it.Handle)
	}
	inv.ask = nil
}

// remove sends a Sell, Drop or Ready (or any other action) on item i. The
// engine's Index of every later item shifts and the list is sorted, so the
// highlight is kept on the line, not the Index: selling down a list, or
// readying out of the party inventory, stays in place.
func (inv *inventory) remove(k ActKind, i rules.ItemHandle) {
	inv.removeX(k, i, 0)
}

// removeX is remove with the action's X (Ready's count).
func (inv *inventory) removeX(k ActKind, i rules.ItemHandle, x int) {
	inv.acts = append(inv.acts, Action{Kind: k, Item: i, X: x})
	if k != ActSell && k != ActDrop && k != ActReady {
		return
	}
	shown := inv.list.shown(inv.up.pool) // a readied item on the doll is not in it: row -1
	inv.keep.row, inv.keep.n, inv.keep.i = slices.IndexFunc(shown, func(it game.Item) bool { return it.Handle == i }), len(shown), i
	inv.keep.id = ""
	if k == ActReady && inv.keep.row >= 0 { // readying keeps the item in the list, moved (readied gear first): the highlight follows it
		inv.keep.id = itemID(shown[inv.keep.row])
	}
}

// itemID is what tells an item from the others whatever its row or Index.
func itemID(it game.Item) string {
	return fmt.Sprint(byte(it.Type), it.Words, strings.TrimSpace(it.Name))
}

// choose highlights ware i of the shop or item i of the member's.
func (inv *inventory) choose(ware bool, i rules.ItemHandle) {
	inv.sel, inv.ware, inv.keep.row = rules.NoItem, rules.NoItem, -1
	if ware {
		inv.ware = i
	} else {
		inv.sel = i
	}
}

// invKind is which inventory screen the state puts up.
type invKind int

const (
	invNone invKind = iota
	invOwn          // the member's items and the party's
	invShop         // the shop's goods beside them
	invLoot         // a fight's treasure beside them
)

// invUp is what the inventory screen shows, read from the state: the items
// are the state's own (never copies), the rest facts of Allowed and the
// member.
type invUp struct {
	kind                                                      invKind
	member                                                    int // the party index (the selected, or the acting member in a fight)
	pool                                                      []game.Item
	doll                                                      []*game.Item // the readied items by doll slot (DollSlots, nil: empty)
	wares                                                     []game.Item  // the shop's goods or the treasure
	capacity, used                                            int          // the party pack's slots (0: no limit) and what it holds
	use, sell, reload, switchable, trade, halve, join, pooled bool
	powers                                                    []Power    // what the member can do besides items (the View menu's)
	treasure                                                  *FrontLoot // invLoot: the treasure's money and other options
	gold, copper                                              int        // the party's money, in a shop
	ac, thac0                                                 int
	stats                                                     bool // the game publishes AC and THAC0 (a member with a THAC0)
}

// inventoryKind is the inventory screen ps puts up: the items (own,
// shop's, treasure's), else the treasure when it is up.
func (r *frontState) inventoryKind(ps game.State) invKind {
	c := game.CommonOf(ps)
	mode, loot := r.modeOf(ps), lootOf(ps)
	wares := false
	switch r.itemsUp(ps) {
	case game.ItemsOwn:
		return invOwn
	case game.ItemsShop, game.ItemsLoot:
		wares = true
	case game.ItemsNone: // the treasure is up: its screen is the loot inventory
		wares = mode == game.ModeLoot && loot != nil && r.sheetOf(c) == nil && !announced(ps)
	}
	switch {
	case !wares:
		return invNone
	case loot != nil:
		return invLoot
	}
	return invShop
}

// inventoryOf is the inventory screen ps puts up (nil: none).
func (r *frontState) inventoryOf(ps game.State) *invUp {
	kind := r.inventoryKind(ps)
	if kind == invNone {
		return nil
	}
	c, party, sel := game.CommonOf(ps), partyOf(ps), r.selectedOf(ps)
	u := &invUp{kind: kind, capacity: c.PartyCap, used: c.PoolItems, doll: make([]*game.Item, len(DollSlots))}
	if sel < 0 || sel >= len(party) {
		return u
	}
	m, allowed, shop, loot, mode := party[sel], allowedOf(ps), shopOf(ps), lootOf(ps), r.modeOf(ps)
	u.member, u.pool = sel, m.Items
	u.use, u.sell, u.reload, u.switchable = allowed["UseItem"], allowed["Sell"], shop != nil && shop.Reloads, allowed["Select"]
	u.trade, u.halve, u.join, u.pooled = allowed["Trade"], allowed["Halve"], allowed["Join"], shop != nil && allowed["PoolCoins"]
	u.ac, u.thac0, u.stats = m.AC, m.THAC0, m.THAC0 != 0
	for _, p := range m.Powers {
		if k, ok := powerActs[p.Kind]; ok && (k != ActSpells || !u.sell) { // no spells in a shop
			u.powers = append(u.powers, Power{Name: p.Label, Kind: k})
		}
	}
	for i, it := range m.Items {
		if d := dollSlot(it, u.doll[rules.DollWeapon] != nil); it.Readied && d >= 0 && u.doll[d] == nil {
			u.doll[d] = &m.Items[i]
		}
	}
	if kind != invOwn {
		u.gold, u.copper = c.Purse[Coin], c.Purse[game.Copper]
		switch {
		case loot != nil:
			u.wares, u.treasure = loot.Items, Loot(mode, loot, allowed, party, sel)
		case shop != nil:
			u.wares = shop.Goods
		}
	}
	return u
}

// itemAt is the screen's item h, nil if none.
func (inv *inventory) itemAt(h rules.ItemHandle) *game.Item {
	if h == rules.NoItem {
		return nil
	}
	for _, it := range inv.up.doll {
		if it != nil && it.Handle == h {
			return it
		}
	}
	for k := range inv.up.pool {
		if inv.up.pool[k].Handle == h {
			return &inv.up.pool[k]
		}
	}
	return nil
}

// wareAt is the shop's ware h, nil if none.
func (inv *inventory) wareAt(h rules.ItemHandle) *game.Item {
	if h == rules.NoItem {
		return nil
	}
	for k := range inv.up.wares {
		if inv.up.wares[k].Handle == h {
			return &inv.up.wares[k]
		}
	}
	return nil
}

// salePrice is what a shop pays for it; warePrice what it costs there.
func salePrice(it game.Item) Money {
	return Money{Gold: it.Sale[Coin], Copper: it.Sale[game.Copper]}
}
func warePrice(it game.Item) Money { return Money{Gold: it.Price} }

// scribing: a scroll someone was going to scribe from (the front end asks before it goes).
func scribing(it game.Item) bool {
	return slices.ContainsFunc(it.Scroll, func(s game.ScrollSpell) bool { return s.Scribing })
}
