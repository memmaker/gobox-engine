// invdraw.go draws the inventory screen: the menu row's words, the doll, the member panel, tooltips and borders.

package present

import (
	"gobox/front"
	"gobox/game"
	"gobox/gpu"
	"gobox/ui"
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
	act  front.ActKind
	on   bool // it does something now
	r    image.Rectangle
}

// itemAt is the inventory's item with Index i, nil if none.
func itemAt(in *front.Inventory, i int) *front.InvItem {
	for _, it := range in.Doll {
		if it != nil && it.Index == i {
			return it
		}
	}
	for k := range in.Pool {
		if in.Pool[k].Index == i {
			return &in.Pool[k]
		}
	}
	return nil
}

// menu is the menu row for the highlighted item.
func (inv *inventory) menu(in *front.Inventory) []word {
	it := itemAt(in, inv.sel)
	if inv.ask != nil { // the question has its Yes and No
		return nil
	}
	if in.Loot { // the loot: take a piece of treasure
		t := in.Treasure
		if t == nil {
			t = &front.Loot{}
		}
		w := []word{{text: "Take", act: front.ActBuy, on: wareAt(in, inv.ware) != nil}, {text: "All", act: front.ActTakeAll, on: len(in.Wares) > 0 || len(t.Money) > 0}}
		for _, o := range []struct {
			on   bool
			text string
			act  front.ActKind
		}{{len(t.Money) > 0, "Money", front.ActTakeMoney}, {t.Detect, "Detect", front.ActLootDetect}, {t.Share, "Share", front.ActLootShare}, {t.Pool, "Pool", front.ActLootPool}, {t.Status, "Status", front.ActLootStatus}} {
			if o.on {
				w = append(w, word{text: o.text, act: o.act, on: true})
			}
		}
		return append(w, word{text: "Exit", act: front.ActClose, on: true})
	}
	if in.Wares != nil { // the shop: buy a ware, sell an item
		w := []word{{text: "Buy", act: front.ActBuy, on: wareAt(in, inv.ware) != nil}, {text: "Sell", act: front.ActSell, on: it != nil && in.Sell}}
		if in.Reload {
			w = append(w, word{text: "Reload", act: front.ActReload, on: it != nil})
		}
		w = append(w, word{text: "View", act: front.ActSheet, on: true}) // look-only: the member's sheet
		if in.PoolCoins {
			w = append(w, word{text: "Pool", act: front.ActLootPool, on: true})
		}
		return append(w, word{text: "Exit", act: front.ActClose, on: true})
	}
	w := []word{{text: "Ready", act: front.ActReady, on: it != nil && it.Slot}}
	if in.Use {
		w = append(w, word{text: "Use", act: front.ActUse, on: it != nil && it.Use})
	}
	for _, o := range []struct {
		on   bool
		text string
		act  front.ActKind
	}{{in.Trade, "Trade", front.ActTrade}, {in.Halve, "Halve", front.ActHalve}, {in.Join, "Join", front.ActJoin}} {
		if o.on {
			w = append(w, word{text: o.text, act: o.act, on: it != nil})
		}
	}
	w = append(w, word{text: "Drop", act: front.ActDrop, on: it != nil})
	if in.Sell {
		w = append(w, word{text: "Sell", act: front.ActSell, on: it != nil})
	}
	for _, pw := range in.Powers {
		w = append(w, word{text: pw.Name, act: pw.Kind, on: true})
	}
	return w // Tab is the sheet, Esc leaves
}

// Inventory draws the inventory screen when the game shows it, and
// reports whether it did.
func (p *Presenter) Inventory(dst *gpu.Image, note Note, v front.View) bool {
	inv := &p.inv
	inv.shown = false
	in := v.Inventory
	if p.own(v).Ask != nil && p.q == nil { // the game's question (Heal whom?, Use it?): Draw asks it, never hidden under the items; the presenter's own (Trade to whom?) is asked over the items
		return false
	}
	if p.playing() != nil { // an event to show (Draw plays it); this also drops the played ones, else Playing holds every key
		return false
	}
	if in == nil {
		inv.ask, inv.view = nil, nil
		return false
	}
	p.shownView = p.own(v) // the decisions are made on the inventory, not on the screen under it
	inv.load()
	if inv.view == nil || in.Member != inv.member {
		inv.sel, inv.list.top, inv.carry, inv.keep.row = -1, 0, nil, -1 // a new member: nothing highlighted or on the cursor
	}
	if inv.view == nil || (inv.view.Wares == nil) != (in.Wares == nil) {
		inv.ware, inv.wares.top = -1, 0
	}
	inv.shown, inv.member, inv.count, inv.glyph, inv.view = true, in.Member, len(v.Party), note.Glyph, in
	if shown, k := inv.list.shown(in.Pool), inv.keep; k.row >= 0 && (len(shown) != k.n || shown[k.row].Index != k.i) { // the sold, dropped or readied item is gone (a swap puts another in its place)
		inv.sel = -1
		if len(shown) > 0 {
			inv.sel = shown[min(inv.keep.row, len(shown)-1)].Index
		}
		inv.keep.row = -1
	}
	if itemAt(in, inv.sel) == nil {
		inv.sel = -1
	}
	if wareAt(in, inv.ware) == nil {
		inv.ware = -1
	}
	if o := inv.order(); inv.ware < 0 && inv.sel < 0 && len(o) > 0 { // something is always highlighted: the keys reach every item
		inv.choose(o[0].ware, o[0].i)
	}
	shop := in.Wares != nil
	split := leftCells // the left box's width in cells, with its border
	busy := inv.ask != nil
	p.setup(note, v)
	t := p.border.tiles

	b := dst.Bounds()
	inv.k = max(1, min(b.Dx()/invW, b.Dy()/invH))
	cw, ch := max(invW, b.Dx()/inv.k)/8, max(invH, b.Dy()/inv.k)/8
	if shop { // the goods and the party inventory share the screen
		split = (cw + 1) / 2
	}
	if w, h := cw*8, ch*8; gpu.Sized(&inv.img, w, h) {
		inv.w, inv.h, inv.px = w, h, image.NewRGBA(image.Rect(0, 0, w, h))
	}
	inv.fill(inv.px.Rect, color.RGBA{A: 255})
	inv.lines = inv.lines[:0]
	bottom := (ch - 2) * 8 // the boxes' inner bottom: under them the menu row, the toast in its place while it shows

	// the member: name, portrait between the arrows, hit points (the shop: its goods)
	lx, lw := 8, (leftCells-2)*8
	inv.dollAt = image.Rectangle{}
	inv.facePNG = nil
	if shop {
		inv.drawTabs(&inv.wares, lx, busy)
		inv.drawPane(&inv.wares, inv.wares.shown(in.Wares), inv.ware, true, !in.Loot, busy, lx, (split-1)*8, bottom) // the loot has no prices; a shop's goods do
	} else {
		inv.wares.tabs = [len(tabNames)]image.Rectangle{} // no shop: no tabs to click
		p.drawMember(in, v, lx, lw, bottom, busy)
	}

	// the party inventory: its title and fill, the lines, the scroll bar
	rx, rr := split*8, (cw-1)*8 // a cell inside the shared divider
	inv.drawTabs(&inv.list, rx, busy)
	fillc := green
	if in.Cap > 0 && len(in.Pool) >= in.Cap {
		fillc = red
	}
	fill := strconv.Itoa(len(in.Pool))
	if in.Cap > 0 { // 0: no limit (game.State.PartyCap), the count alone
		fill += "/" + strconv.Itoa(in.Cap)
	}
	inv.drawPane(&inv.list, inv.list.shown(in.Pool), inv.sel, false, shop && !in.Loot, busy, rx, rr, bottom) // the shop prices the party's items (it buys them); the loot does not

	// an item dragged, or the stack picked up, under the pointer
	if inv.carry != nil && inv.carriedItem() == nil { // gone (used, dropped, another member)
		inv.carry = nil
	}
	if inv.dragging && inv.press != nil {
		inv.drawIcon(&inv.press.item, image.Rectangle{inv.hover, inv.hover}.Inset(-rowH/2))
	} else if c := inv.carry; c != nil {
		inv.drawIcon(&c.icon, image.Rectangle{inv.hover, inv.hover}.Inset(-rowH/2))
		if c.n > 1 {
			inv.text(strconv.Itoa(c.n), inv.hover.X+rowH/2-8*len(strconv.Itoa(c.n))+2, inv.hover.Y+rowH/2-8, white)
		}
	}

	// the menu row under the boxes; the toast (noteLine) takes its place while it shows
	rowY := (ch - 1) * 8
	inv.words = inv.menu(in)
	toast := note.Show && note.Text != ""
	if toast {
		noteLine(inv.px, note, rowY)
	}
	x := 0
	for _, w := range inv.words {
		x += 8 * (len(w.text) + 1)
	}
	x = (inv.w - x + 8) / 2
	if shop && !toast && in.Member >= 0 && in.Member < len(v.Party) { // the shop: the choices left, the buyer right
		x = 8
		m := v.Party[in.Member]
		s := m.Name + " " + m.Class
		inv.text(m.Name, inv.w-8-8*len(s), rowY, white)
		inv.text(m.Class, inv.w-8-8*len(m.Class), rowY, gray)
	}
	inv.focus = min(inv.focus, len(inv.words))
	for i := range inv.words {
		w := &inv.words[i]
		if toast { // no words to see, none to click
			w.r = image.Rectangle{}
			continue
		}
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
	switch it := itemAt(in, inv.sel); {
	case inv.ask != nil && it != nil && inv.ask.Kind == front.ActDrop:
		box = strings.Split(confirmText(confirmOf(itemConfirm(front.ActDrop, it), game.Confirm{Asks: []string{"Your %s\nwill be gone forever.\n\nDrop it?"}}).Asks[0], strings.TrimSpace(it.Name), ""), "\n")
	case inv.ask != nil && inv.ask.Kind == front.ActBuy:
		if w := wareAt(in, inv.ware); w != nil {
			box = []string{strings.TrimSpace(w.Name), "costs " + price(w.Price) + ".", "", "Buy it?"}
		}
	case inv.ask != nil && it != nil:
		box = strings.Split(confirmText(confirmOf(itemConfirm(front.ActSell, it), game.Confirm{Asks: []string{"I'll give you %p\nfor your %s.\n\nIs it a deal?"}}).Asks[0], strings.TrimSpace(it.Name), price(it.Price)), "\n")
	}
	if it := itemAt(in, inv.sel); box != nil && it != nil && it.Scribe && inv.ask.Kind != front.ActBuy && (inv.ask.Kind != front.ActDrop || confirms["DropScribe"].Asks == nil) { // ovr20:0D70's question (a game's own DropScribe words it whole)
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
	case in.Loot:
		onBorder(inv.img, inv.glyph, lx, 0, true, []string{"Loot"}, []color.RGBA{white})
		if t := in.Treasure; t != nil && len(t.Money) > 0 {
			onBorder(inv.img, inv.glyph, rr, (ch-1)*8-8, false, []string{strings.Join(t.Money, ", ")}, []color.RGBA{col(inkYellow)})
		}
	case shop:
		onBorder(inv.img, inv.glyph, lx, 0, true, []string{"Shop"}, []color.RGBA{white})
		onBorder(inv.img, inv.glyph, rr, (ch-1)*8-8, false, []string{"party " + price(front.Money{Gold: in.Gold, Copper: in.Copper})}, []color.RGBA{col(inkYellow)})
	}
	if !shop && in.Member >= 0 && in.Member < len(v.Party) {
		paneName(inv.img, inv.glyph, v.Party[in.Member].Name)
		if in.Stats && !in.Loot { // the member's AC and THAC0 under its items, as the sheet's
			onBorder(inv.img, inv.glyph, lx, (ch-1)*8-8, true, []string{invStats(in)}, []color.RGBA{col(inkYellow)})
		}
	}
	head := "Party Items"
	switch {
	case shop && in.Member >= 0 && in.Member < len(v.Party): // the shop shows the shopper's items only
		head = v.Party[in.Member].Name
	case v.Battle != nil: // a fight's Use: the acting member's own items, never the party's
		head = "Items"
	}
	onBorder(inv.img, inv.glyph, rx, 0, true, []string{head}, []color.RGBA{white})
	onBorder(inv.img, inv.glyph, rr, 0, false, []string{fill}, []color.RGBA{fillc})
	if box == nil && !inv.dragging {
		inv.tooltip(t, in)
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
	if q := p.q; q != nil && q.Ask.Kind == front.AskMember { // the presenter's member pick over the items: the members by their number keys
		var names []string
		var keys []byte
		for i, m := range v.Party {
			names, keys = append(names, m.Name), append(keys, byte('1'+i))
		}
		p.keyWindow(dst, q.Ask.Prompt, names, keys, nil, q.Cand, false)
	}
	return true
}

// drawMember draws the member's name, portrait, hit points and paper doll
// with the readied items in the left box.
func (p *Presenter) drawMember(in *front.Inventory, v front.View, lx, lw, bottom int, busy bool) {
	inv := &p.inv
	inv.face, inv.facePNG = image.Rectangle{}, nil
	if in.Member >= 0 && in.Member < len(v.Party) {
		inv.face, inv.facePNG = p.memberPane(inv.px, inv.glyph, v.Party[in.Member], bottom)
	}

	// the doll and the readied items
	db := inv.doll.Bounds()
	inv.dollAt = db.Add(image.Pt(lx+(lw-db.Dx())/2, bottom-db.Dy()-4)) // on the box's bottom, under the portrait
	draw.Draw(inv.px, inv.dollAt, inv.doll, db.Min, draw.Src)
	for s, it := range in.Doll {
		if s >= len(dollBoxes) { // a game without a paper doll (no DollBoxes: Silver Blades, Curse) draws no slots; its readied items are on the list
			break
		}
		box := dollBoxes[s].Add(inv.dollAt.Min)
		if it != nil {
			inv.drawIcon(it, box.Inset(1))
			inv.lines = append(inv.lines, hit{r: box, item: *it, doll: true})
		}
		if it != nil && it.Index == inv.sel || inv.hover.In(box) && !busy {
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
		ui.DrawText(px, glyph, t, at, 0, cols[i])
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
func (inv *inventory) tooltip(t ui.Ribbon, in *front.Inventory) {
	var lines []string
	describe := func(it *front.InvItem, price string) {
		lines = []string{strings.TrimSpace(it.Name), ""}
		for _, l := range it.Look {
			if !strings.HasPrefix(l, "Price ") || price == "" { // the shop's own price instead of the list price
				lines = append(lines, l)
			}
		}
		if price != "" {
			lines = append(lines, price)
		}
	}
	for i, r := range dollBoxes {
		if inv.hover.In(r.Add(inv.dollAt.Min)) {
			lines = []string{slotNames[i]}
			if it := in.Doll[i]; it != nil {
				describe(it, "")
			}
		}
	}
	for _, l := range inv.lines {
		if l.doll || !inv.hover.In(l.r) {
			continue
		}
		switch {
		case l.ware && in.Loot:
			describe(&l.item, "")
		case l.ware:
			describe(&l.item, "Costs "+price(l.item.Price))
		case in.Sell:
			describe(&l.item, "Sells for "+price(l.item.Price))
		default:
			describe(&l.item, "")
		}
	}
	if len(lines) == 0 {
		return
	}
	inv.panel(t, lines, func(w, h int) image.Point { // beside the pointer, inside the screen
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
func (inv *inventory) panel(t ui.Ribbon, lines []string, at func(w, h int) image.Point) {
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
			ui.DrawText(px, inv.glyph, l[j:], 8+8*j, 8+8*i, dc)
			l = l[:j]
		}
		ui.DrawText(px, inv.glyph, l, 8, 8+8*i, c)
	}
	inv.tip.WritePixels(px.Pix)
	window(inv.tip, t, gpu.GeoM{}, cw, rows, nil)
	op := &gpu.DrawImageOptions{}
	op.GeoM.Translate(float64(p.X), float64(p.Y))
	inv.img.DrawImage(inv.tip, op)
	inv.panelAt = image.Rectangle{Min: p, Max: p.Add(image.Pt(w, h))}
}

// invStats is the items page's AC and THAC0 line.
func invStats(in *front.Inventory) string {
	return "AC " + strconv.Itoa(in.AC) + "  THAC0 " + strconv.Itoa(in.THAC0)
}
