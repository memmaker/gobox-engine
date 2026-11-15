// invdraw.go draws the inventory screen: the menu row's words, the doll, the member panel, tooltips and borders.

package ui

import (
	"gobox/game"
	"gobox/gpu"
	"gobox/rules"
	"image"
	"image/color"
	"image/draw"
	"slices"
	"strconv"
	"strings"
)

// word is a choice on the menu row (or a question's Yes and No): its
// text, the action it sends and where it is (canvas).
type word struct {
	text string
	act  ActKind
	on   bool // it does something now
	r    image.Rectangle
}

// menu is the menu row for the highlighted item.
func (inv *inventory) menu() []word {
	in := inv.up
	it := inv.itemAt(inv.sel)
	if inv.ask != nil { // the question has its Yes and No
		return nil
	}
	if in.kind == invLoot { // the loot: take a piece of treasure
		t := in.treasure
		if t == nil {
			t = &FrontLoot{}
		}
		w := []word{{text: "Take", act: ActBuy, on: inv.wareAt(inv.ware) != nil}, {text: "All", act: ActTakeAll, on: len(in.wares) > 0 || len(t.Money) > 0}}
		for _, o := range []struct {
			on   bool
			text string
			act  ActKind
		}{{len(t.Money) > 0, "Money", ActTakeMoney}, {t.Detect, "Detect", ActLootDetect}, {t.Share, "Share", ActLootShare}, {t.Pool, "Pool", ActLootPool}, {t.Status, "Status", ActLootStatus}, {t.Drop, "Drop", ActDrop}} {
			if o.on {
				w = append(w, word{text: o.text, act: o.act, on: o.act != ActDrop || it != nil})
			}
		}
		return append(w, word{text: "Exit", act: ActClose, on: true})
	}
	if in.kind != invOwn { // the shop: buy a ware, sell an item
		w := []word{{text: "Buy", act: ActBuy, on: inv.wareAt(inv.ware) != nil}, {text: "Sell", act: ActSell, on: it != nil && in.sell}}
		if in.reload {
			w = append(w, word{text: "Reload", act: ActReload, on: it != nil})
		}
		w = append(w, word{text: "View", act: ActSheet, on: true}) // look-only: the member's sheet
		if in.pooled {
			w = append(w, word{text: "Pool", act: ActLootPool, on: true})
		}
		return append(w, word{text: "Exit", act: ActClose, on: true})
	}
	w := []word{{text: "Ready", act: ActReady, on: it != nil && slotted(*it)}}
	if in.use {
		w = append(w, word{text: "Use", act: ActUse, on: it != nil && it.Usable})
	}
	for _, o := range []struct {
		on   bool
		text string
		act  ActKind
	}{{in.trade, "Trade", ActTrade}, {in.halve, "Halve", ActHalve}, {in.join, "Join", ActJoin}} {
		if o.on {
			w = append(w, word{text: o.text, act: o.act, on: it != nil})
		}
	}
	w = append(w, word{text: "Drop", act: ActDrop, on: it != nil})
	if in.sell {
		w = append(w, word{text: "Sell", act: ActSell, on: it != nil})
	}
	for _, pw := range in.powers {
		w = append(w, word{text: pw.Name, act: pw.Kind, on: true})
	}
	return w // Tab is the sheet, Esc leaves
}

// Inventory draws the inventory screen when the game shows it, and
// reports whether it did.
func (p *Presenter) Inventory(dst *gpu.Image, note Note) bool {
	p.sync()
	inv := &p.inv
	inv.shown = false
	in := p.inventory()
	party := p.party()
	if p.q == nil && p.askUp() != nil { // the game's question (Heal whom?, Use it?): Draw asks it, never hidden under the items; the presenter's own (Trade to whom?) is asked over the items
		return false
	}
	if p.playing() != nil { // an event to show (Draw plays it); this also drops the played ones, else Playing holds every key
		return false
	}
	if in == nil {
		inv.ask, inv.up = nil, nil
		return false
	}
	inv.load()
	if inv.up == nil || in.member != inv.member {
		inv.sel, inv.list.top, inv.carry, inv.keep.row = rules.NoItem, 0, nil, -1 // a new member: nothing highlighted or on the cursor
	}
	if inv.up == nil || (inv.up.kind == invOwn) != (in.kind == invOwn) {
		inv.ware, inv.wares.top = rules.NoItem, 0
	}
	inv.shown, inv.member, inv.count, inv.glyph, inv.up = true, in.member, len(party), note.Glyph, in
	if shown, k := inv.list.shown(in.pool), inv.keep; k.row >= 0 && (len(shown) != k.n || shown[k.row].Handle != k.i || k.id != "" && itemID(shown[k.row]) != k.id) { // the sold, dropped or readied item is gone (a swap puts another in its place)
		inv.sel = rules.NoItem
		if j := slices.IndexFunc(shown, func(it game.Item) bool { return k.id != "" && itemID(it) == k.id }); j >= 0 { // a readied or unreadied item moved: it stays highlighted
			inv.sel = shown[j].Handle
		} else if len(shown) > 0 {
			inv.sel = shown[min(inv.keep.row, len(shown)-1)].Handle
		}
		inv.keep.row = -1
	}
	if inv.itemAt(inv.sel) == nil {
		inv.sel = rules.NoItem
	}
	if inv.wareAt(inv.ware) == nil {
		inv.ware = rules.NoItem
	}
	if o := inv.order(); inv.ware == rules.NoItem && inv.sel == rules.NoItem && len(o) > 0 { // something is always highlighted: the keys reach every item
		inv.choose(o[0].ware, o[0].i)
	}
	shop := in.kind != invOwn
	split := leftCells // the left box's width in cells, with its border
	busy := inv.ask != nil
	p.setup(note)
	t := p.border.val.tiles

	b := dst.Bounds()
	var cw, ch int
	inv.k, cw, ch = FullPageCells(b)
	if shop { // the goods and the party inventory share the screen
		split = (cw + 1) / 2
	}
	if w, h := cw*8, ch*8; gpu.Sized(&inv.img, w, h) {
		inv.w, inv.h, inv.px = w, h, image.NewRGBA(image.Rect(0, 0, w, h))
	}
	inv.fill(inv.px.Rect, color.RGBA{A: 255})
	inv.lines = inv.lines[:0]
	bottom := (ch - 2) * 8 // the boxes' inner bottom: under them the menu row

	// the member: name, portrait between the arrows, hit points (the shop: its goods)
	lx, lw := 8, (leftCells-2)*8
	inv.dollAt = image.Rectangle{}
	inv.facePNG = nil
	if shop {
		inv.drawTabs(&inv.wares, lx, busy)
		inv.drawPane(&inv.wares, inv.wares.shown(in.wares), inv.ware, true, in.kind != invLoot, busy, lx, (split-1)*8, bottom) // the loot has no prices; a shop's goods do
	} else {
		inv.wares.tabs = [len(tabNames)]image.Rectangle{} // no shop: no tabs to click
		p.drawMember(lx, lw, bottom, busy)
	}

	// the party inventory: its title and fill, the lines, the scroll bar
	rx, rr := split*8, (cw-1)*8 // a cell inside the shared divider
	inv.drawTabs(&inv.list, rx, busy)
	fillc := green
	fill := strconv.Itoa(len(in.pool))
	if in.capacity > 0 { // 0: no limit (play.Common.PartyCap), the count alone; else the pack's slots, readied gear not counted
		fill = strconv.Itoa(in.used) + "/" + strconv.Itoa(in.capacity)
		if in.used >= in.capacity {
			fillc = red
		}
	}
	inv.drawPane(&inv.list, inv.list.shown(in.pool), inv.sel, false, shop && in.kind != invLoot, busy, rx, rr, bottom) // the shop prices the party's items (it buys them); the loot does not

	// an item dragged, or the stack picked up, under the pointer
	if inv.carry != nil && inv.carriedItem() == nil { // gone (used, dropped, another member)
		inv.carry = nil
	}
	if inv.dragging && inv.press != nil {
		inv.drawIcon(inv.press.item.Base, image.Rectangle{inv.hover, inv.hover}.Inset(-rowH/2))
	} else if c := inv.carry; c != nil {
		inv.drawIcon(c.icon.Base, image.Rectangle{inv.hover, inv.hover}.Inset(-rowH/2))
		if c.n > 1 {
			inv.text(strconv.Itoa(c.n), inv.hover.X+rowH/2-8*len(strconv.Itoa(c.n))+2, inv.hover.Y+rowH/2-8, white)
		}
	}

	// the menu row under the boxes
	rowY := (ch - 1) * 8
	inv.words = inv.menu()
	x := 0
	for _, w := range inv.words {
		x += 8 * (len(w.text) + 1)
	}
	x = (inv.w - x + 8) / 2
	if shop && in.member >= 0 && in.member < len(party) { // the shop: the choices left, the buyer right
		x = 8
		m := party[in.member]
		s := m.Name + " " + m.Class
		inv.text(m.Name, inv.w-8-8*len(s), rowY, white)
		inv.text(m.Class, inv.w-8-8*len(m.Class), rowY, gray)
	}
	inv.focus = min(inv.focus, len(inv.words))
	for i := range inv.words {
		w := &inv.words[i]
		w.r = image.Rect(x-4, rowY-2, x+8*len(w.text)+4, rowY+10)
		c, hc := green, white
		switch {
		case !w.on:
			c, hc = dark, dark
		case inv.hover.In(w.r) || inv.focus == i+1:
			if inkBlue >= 0 {
				inv.fill(w.r, col(inkBlue)) // the text selection
			}
		}
		inv.text(w.text[:1], x, rowY, hc)
		inv.text(w.text[1:], x+8, rowY, c)
		x += 8 * (len(w.text) + 1)
	}

	// the question over the list
	var box []string
	switch it := inv.itemAt(inv.sel); {
	case inv.ask != nil && it != nil && inv.ask.Kind == ActDrop:
		box = strings.Split(confirmText(confirmOf(itemConfirm(ActDrop, *it), game.Confirm{Asks: []string{"Your %s\nwill be gone forever.\n\nDrop it?"}}).Asks[0], strings.TrimSpace(it.Name), ""), "\n")
	case inv.ask != nil && inv.ask.Kind == ActBuy:
		if w := inv.wareAt(inv.ware); w != nil {
			box = []string{strings.TrimSpace(w.Name), "costs " + price(warePrice(*w)) + ".", "", "Buy it?"}
		}
	case inv.ask != nil && it != nil:
		box = strings.Split(sellAsk(it), "\n")
	}
	if it := inv.itemAt(inv.sel); box != nil && it != nil && scribing(*it) && inv.ask.Kind != ActBuy && (inv.ask.Kind != ActDrop || confirms["DropScribe"].Asks == nil) { // ovr20:0D70's question (a game's own DropScribe words it whole)
		box = append([]string{"Someone was going to", "scribe from it.", ""}, box...)
	}

	inv.panelAt = image.Rectangle{}
	// onto the GPU: the boxes' borders, then the box's
	inv.img.WritePixels(inv.px.Pix)
	window(inv.img, t, gpu.GeoM{}, split, ch-1, nil)
	var g gpu.GeoM
	g.Translate(float64(split-1)*8, 0) // the boxes share the divider
	window(inv.img, t, g, cw-split+1, ch-1, nil)
	divider(inv.img, t, g, ch-1)
	switch { // the headings on the top border, the party's money on the shop's bottom one
	case in.kind == invLoot:
		onBorder(inv.img, inv.glyph, lx, 0, true, []string{"Loot"}, []color.RGBA{white})
		if t := in.treasure; t != nil && len(t.Money) > 0 {
			onBorder(inv.img, inv.glyph, rr, (ch-1)*8-8, false, []string{strings.Join(t.Money, ", ")}, []color.RGBA{col(inkYellow)})
		}
	case shop:
		onBorder(inv.img, inv.glyph, lx, 0, true, []string{"Shop"}, []color.RGBA{white})
		onBorder(inv.img, inv.glyph, rr, (ch-1)*8-8, false, []string{"party " + price(Money{Gold: in.gold, Copper: in.copper})}, []color.RGBA{col(inkYellow)})
	}
	if !shop && in.member >= 0 && in.member < len(party) {
		paneName(inv.img, inv.glyph, party[in.member].Name)
		if in.stats && in.kind != invLoot { // the member's AC and THAC0 under its items, as the sheet's
			onBorder(inv.img, inv.glyph, lx, (ch-1)*8-8, true, []string{invStats(in)}, []color.RGBA{col(inkYellow)})
		}
	}
	head := "Party Items"
	switch {
	case shop && in.member >= 0 && in.member < len(party): // the shop shows the shopper's items only
		head = party[in.member].Name
	case p.board != nil: // a fight's Use: the acting member's own items, never the party's
		head = "Items"
	}
	onBorder(inv.img, inv.glyph, rx, 0, true, []string{head}, []color.RGBA{white})
	onBorder(inv.img, inv.glyph, rr, 0, false, []string{fill}, []color.RGBA{fillc})
	if box == nil && !inv.dragging && tipShown() {
		inv.tooltip(t)
	}

	// onto the whole screen
	inv.kx, inv.ky = p.fit(dst, inv.img, inv.w, inv.h)
	if inv.facePNG != nil { // the PNG portrait at the window's resolution, as the cards
		op := &gpu.DrawImageOptions{}
		op.GeoM.Translate(float64(inv.face.Min.X)*inv.kx, float64(inv.face.Min.Y)*inv.ky)
		dst.DrawImage(p.hiFace(inv.facePNG, int(float64(inv.face.Dx())*inv.kx)), op)
		if r := inv.panelAt; r.Overlaps(inv.face) { // the tooltip or question stays on top
			op := &gpu.DrawImageOptions{}
			op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
			op.GeoM.Scale(inv.kx, inv.ky)
			dst.DrawImage(inv.img.SubImage(r).(*gpu.Image), op)
		}
	}
	inv.askAt = nil
	if box != nil { // the question: the standard menu, over everything
		p.keyWindow(dst, strings.Join(box, " "), []string{"Yes", "No"}, []byte{'Y', 'N'}, nil, inv.askSel, false)
		inv.askAt = slices.Clone(p.winRects)
	}
	if q := p.q; q != nil && q.Ask.Kind == AskMember { // the presenter's member pick over the items: the members by their number keys
		var names []string
		var keys []byte
		for i, m := range party {
			names, keys = append(names, m.Name), append(keys, byte('1'+i))
		}
		p.keyWindow(dst, q.Ask.Prompt, names, keys, nil, q.Cand, false)
	}
	return true
}

// drawMember draws the member's name, portrait, hit points and paper doll
// with the readied items in the left box.
func (p *Presenter) drawMember(lx, lw, bottom int, busy bool) {
	inv := &p.inv
	in := inv.up
	inv.face, inv.facePNG = image.Rectangle{}, nil
	if party := p.party(); in.member >= 0 && in.member < len(party) {
		inv.face, inv.facePNG = p.memberPane(inv.px, inv.glyph, party[in.member], bottom)
	}

	// the doll and the readied items
	db := inv.doll.Bounds()
	inv.dollAt = db.Add(image.Pt(lx+(lw-db.Dx())/2, bottom-db.Dy()-4)) // on the box's bottom, under the portrait
	draw.Draw(inv.px, inv.dollAt, inv.doll, db.Min, draw.Src)
	for s, it := range in.doll {
		if s >= len(dollBoxes) { // a game without a paper doll (no DollBoxes: Silver Blades, Curse) draws no slots; its readied items are on the list
			break
		}
		box := dollBoxes[s].Add(inv.dollAt.Min)
		if it != nil {
			inv.drawIcon(it.Base, box.Inset(1))
			inv.lines = append(inv.lines, hit{r: box, item: *it, doll: true})
		}
		if it != nil && it.Handle == inv.sel || inv.hover.In(box) && !busy {
			selectOutline(box, 1, inv.fill)
		}
	}
}

// onBorder draws texts in their colours, a space apart on black, over
// the borders with the right end at x (left end when left), at row y.
func onBorder(dst *gpu.Image, glyph func(byte) [8]byte, x, y int, left bool, texts []string, cols []color.RGBA) {
	n := len(texts) + 1
	for _, t := range texts {
		n += len(t)
	}
	w := 8 * n
	px := image.NewRGBA(image.Rect(0, 0, w, 8))
	draw.Draw(px, px.Rect, image.Black, image.Point{}, draw.Src)
	at := 8
	for i, t := range texts {
		DrawText(px, glyph, t, at, 0, cols[i])
		at += 8 * (len(t) + 1)
	}
	m := gpu.NewImage(w, 8)
	defer m.Deallocate()
	m.WritePixels(px.Pix)
	if !left {
		x -= w
	}
	op := &gpu.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	dst.DrawImage(m, op)
}

// tooltip tells what the pointer is on: a doll slot's item (or what goes
// there), an item's name, what it does and what it costs or sells for.
func (inv *inventory) tooltip(t Ribbon) {
	in := inv.up
	var text []string
	describe := func(it *game.Item, price string) {
		text = []string{strings.TrimSpace(it.Name), ""}
		for _, l := range lines(it.Text) {
			if !strings.HasPrefix(l, "Price ") || price == "" { // the shop's own price instead of the list price
				text = append(text, l)
			}
		}
		if price != "" {
			text = append(text, price)
		}
	}
	for i, r := range dollBoxes {
		if inv.hover.In(r.Add(inv.dollAt.Min)) {
			text = []string{slotNames[i]}
			if it := in.doll[i]; it != nil {
				describe(it, "")
			}
		}
	}
	for _, l := range inv.lines {
		if l.doll || !inv.hover.In(l.r) {
			continue
		}
		switch {
		case l.ware && in.kind == invLoot:
			describe(&l.item, "")
		case l.ware:
			describe(&l.item, "Costs "+price(warePrice(l.item)))
		case in.sell:
			describe(&l.item, "Sells for "+price(salePrice(l.item)))
		default:
			describe(&l.item, "")
		}
	}
	if len(text) == 0 {
		return
	}
	inv.panel(t, text, func(w, h int) image.Point { // beside the pointer, inside the screen
		x := max(0, min(inv.hover.X+8, inv.w-w))
		y := inv.hover.Y + 12
		if y+h > inv.h {
			y = max(0, inv.hover.Y-h-4)
		}
		return image.Pt(x, y)
	})
}

// panel draws lines in a box of the game's border over the canvas (over
// the boxes' borders), the first line cyan, at(w, h) its top left.
func (inv *inventory) panel(t Ribbon, lines []string, at func(w, h int) image.Point) {
	cw := 0
	for _, l := range lines {
		cw = max(cw, len(l))
	}
	cw = min(cw, inv.w/8-2) + 2
	rows := len(lines) + 2
	w, h := cw*8, rows*8
	p := at(w, h)
	gpu.Sized(&inv.tip, w, h)
	px := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(px, px.Rect, image.Black, image.Point{}, draw.Src)
	for i, l := range lines {
		c := gray
		if i == 0 {
			c = cyan
		}
		l = upTo(l, cw-2)
		if j := strings.LastIndex(l, " ("); j >= 0 && strings.HasSuffix(l, ")") && (l[j+2] == '+' || l[j+2] == '-') { // the Look's difference to the readied item: green better, red worse
			dc := green
			if l[j+2] == '-' {
				dc = red
			}
			DrawText(px, inv.glyph, l[j:], 8+8*j, 8+8*i, dc)
			l = l[:j]
		}
		DrawText(px, inv.glyph, l, 8, 8+8*i, c)
	}
	inv.tip.WritePixels(px.Pix)
	window(inv.tip, t, gpu.GeoM{}, cw, rows, nil)
	op := &gpu.DrawImageOptions{}
	op.GeoM.Translate(float64(p.X), float64(p.Y))
	inv.img.DrawImage(inv.tip, op)
	inv.panelAt = image.Rectangle{Min: p, Max: p.Add(image.Pt(w, h))}
}

// invStats is the items page's AC and THAC0 line.
func invStats(in *invUp) string {
	return "AC " + strconv.Itoa(in.ac) + "  THAC0 " + strconv.Itoa(in.thac0)
}

// sellAsk is the shop's sale question for it: the offer, then the game's
// ask (Data.Confirms), a game's ask that is the question alone too.
func sellAsk(it *game.Item) string {
	ask := confirmOf(itemConfirm(ActSell, *it), game.Confirm{Asks: []string{"I'll give you %p\nfor your %s.\n\nIs it a deal?"}}).Asks[0]
	if salePrice(*it) != (Money{}) && !strings.Contains(ask, "%p") {
		ask = "I'll give you %p\nfor your %s.\n\n" + ask
	}
	return confirmText(ask, strings.TrimSpace(it.Name), price(salePrice(*it)))
}

// inventory is the inventory screen up (nil: none or no game).
func (p *Presenter) inventory() *invUp {
	if p.eng == nil || p.pickerOver() {
		return nil
	}
	return p.lay.inventory.get(p.uiKey(), func() *invUp { return p.eng.r.inventoryOf(p.state()) })
}

// inventoryKind is which inventory screen is up.
func (p *Presenter) inventoryKind() invKind {
	if p.eng == nil || p.pickerOver() {
		return invNone
	}
	if in := p.inventory(); in != nil {
		return in.kind
	}
	return invNone
}

// inventoryUp: an inventory screen is up.
func (p *Presenter) inventoryUp() bool { return p.inventoryKind() != invNone }

// pickerOver: a picker or the icon editor is up over the screens under it
// (ownPicker puts the sheet, the items and the party menu away).
func (p *Presenter) pickerOver() bool { return p.gridPick != nil || p.icon != nil }
