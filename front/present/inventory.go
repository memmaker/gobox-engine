// inventory.go holds the inventory screen's state, its picked-up stack and the decisions it sends.

package present

import (
	"fmt"
	"gobox/front"
	"image"
	"image/color"
	"slices"
	"time"

	"gobox/ega"
	"gobox/gpu"
)

// The inventory screen (F1-F8, Items): the member's Eye of the Beholder
// paper doll on the left (the shop's Browse: its goods, half the screen
// each), the party inventory on the right, in the
// original's font, colours and box borders. Everything is drawn 1:1 on one
// canvas that is scaled whole to the window, so all pixels are the same
// size. It draws front.View.Inventory and acts through front.Action: the
// menu row, the tooltips and the Drop and Sell questions are its own.

// Layout, in canvas pixels. The canvas is at least invW x invH; the list
// takes what the window gives beyond that.
const (
	invW, invH = 400, 250
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
	c := ega.Palette[i]
	return color.RGBA{c[0], c[1], c[2], 255}
}

var green, white, cyan, gray, dark, red color.RGBA // the roles' colours (applyTheme)

// inventory is the screen's state between frames.
type inventory struct {
	doll, atlas image.Image
	icons       map[string]int // items.rec: icon key → atlas cell
	px          *image.RGBA    // the canvas, drawn on the CPU
	img, tip    *gpu.Image     // the canvas and the tooltip on the GPU
	glyph       func(byte) [8]byte

	shown   bool                    // the last frame was the inventory
	sel     int                     // the highlighted item (InvItem.Index), -1 for none
	keep    struct{ row, n, i int } // a sold, dropped or readied item's line, the list's length and its Index then: the line after it is highlighted once it is gone (row -1 none)
	ask     *front.Action           // a Drop or Sell waits for Yes
	askSel  int                     // the question's highlighted answer: 0 Yes, 1 No
	askAt   []image.Rectangle       // the question's Yes and No on the screen, as last drawn
	view    *front.Inventory
	acts    []front.Action // to send (Presenter.Actions)
	words   []word         // the menu row's words
	focus   int            // 1+ the word the keys are on (PgUp/PgDn, the pad's shoulders), 0 none
	k       int            // canvas scale, whole: what the pixel art is drawn at
	kx, ky  float64        // the canvas on screen: k and the stretch to the window's edges (fit)
	w, h    int            // canvas size
	list    pane           // the party inventory
	wares   pane           // the shop's goods (Browse)
	ware    int            // the highlighted ware (InvItem.Index), -1 for none
	lines   []hit          // the items' rectangles (canvas)
	dollAt  image.Rectangle
	face    image.Rectangle // the portrait on the canvas, when facePNG is drawn over it at the window's resolution
	facePNG image.Image
	panelAt image.Rectangle // the last panel on the canvas, drawn again over facePNG
	hover   image.Point     // the pointer (canvas)
	mouseAt image.Point     // the pointer (screen), as last seen
	member  int             // View.Inventory.Member
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
	icon  front.InvItem // as it looked when picked up (the icon under the pointer)
}

// carriedItem is the carried stack's item in the inventory shown, nil when it is gone.
func (inv *inventory) carriedItem() *front.InvItem {
	c := inv.carry
	if c == nil {
		return nil
	}
	same := func(it *front.InvItem) bool { return it != nil && it.Type == c.typ && slices.Equal(it.Words, c.words) }
	if c.doll {
		for _, it := range inv.view.Doll {
			if same(it) {
				return it
			}
		}
		return nil
	}
	for k := range inv.view.Pool {
		if same(&inv.view.Pool[k]) {
			return &inv.view.Pool[k]
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
	inv.choose(false, h.item.Index)
	inv.carry = &carried{typ: h.item.Type, words: h.item.Words, doll: h.doll, n: n, icon: h.item}
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
		inv.removeX(front.ActReady, it.Index, x)
	}
	if c.n -= n; c.n <= 0 {
		inv.carry = nil
	}
}

type hit struct {
	r     image.Rectangle
	item  front.InvItem
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
	return fmt.Sprint(p.inv.shown, p.Options(), p.overview, p.pager != nil, p.fullMap != nil, p.clue != nil, p.training, p.frontScreens())
}

// call queues a rule call on the engine (Calls).
func (p *Presenter) call(f func(front.Game)) { p.calls.Add(f) }

// Actions hands over the actions the screens made since the last call.
func (p *Presenter) Actions() []front.Action {
	for _, a := range p.battle.Actions() {
		a := askAction(p.shownView, a) // the turn's are rule calls (pace.go)
		p.sendA(a)
	}
	var as []front.Action
	for _, a := range p.inv.acts { // the screens' own (the inventory's, a list's) meet the presenter's questions and rule calls too
		if !p.slotAct(a) {
			as = append(as, a)
		}
	}
	p.inv.acts = nil
	return as
}

// do runs a menu choice on the highlighted item.
func (inv *inventory) do(k front.ActKind) {
	it := itemAt(inv.view, inv.sel)
	switch {
	case k == front.ActNone:
	case k == front.ActClose, k == front.ActSheet, k == front.ActTakeAll, k >= front.ActSpells && k <= front.ActCure, k >= front.ActTakeMoney && k <= front.ActLootStatus:
		inv.acts = append(inv.acts, front.Action{Kind: k})
	case k == front.ActBuy:
		if w := wareAt(inv.view, inv.ware); w != nil {
			inv.acts = append(inv.acts, front.Action{Kind: k, N: w.Index})
		}
	case it == nil:
	case inv.ask == nil && (k == front.ActDrop || k == front.ActSell):
		inv.ask, inv.askSel = &front.Action{Kind: k, N: it.Index}, map[bool]int{true: 0, false: 1}[k == front.ActSell] // a sale starts on Yes, as the Gold Box's; a drop (gone forever) on No
		return
	default:
		inv.remove(k, it.Index)
	}
	inv.ask = nil
}

// remove sends a Sell, Drop or Ready (or any other action) on item i. The
// engine's Index of every later item shifts and the list is sorted, so the
// highlight is kept on the line, not the Index: selling down a list, or
// readying out of the party inventory, stays in place.
func (inv *inventory) remove(k front.ActKind, i int) {
	inv.removeX(k, i, 0)
}

// removeX is remove with the action's X (Ready's count).
func (inv *inventory) removeX(k front.ActKind, i, x int) {
	inv.acts = append(inv.acts, front.Action{Kind: k, N: i, X: x})
	if k != front.ActSell && k != front.ActDrop && k != front.ActReady {
		return
	}
	shown := inv.list.shown(inv.view.Pool) // a readied item on the doll is not in it: row -1
	inv.keep.row, inv.keep.n, inv.keep.i = slices.IndexFunc(shown, func(it front.InvItem) bool { return it.Index == i }), len(shown), i
}

// choose highlights ware i of the shop or item i of the member's.
func (inv *inventory) choose(ware bool, i int) {
	inv.sel, inv.ware, inv.keep.row = -1, -1, -1
	if ware {
		inv.ware = i
	} else {
		inv.sel = i
	}
}
