// invpane.go holds the inventory's item lists: tabs, panes, prices, the item icons and the text and fill primitives they draw with.

package present

import (
	"bytes"
	"fmt"
	"gobox/front"
	"gobox/front/classic"
	"gobox/sprites"
	"gobox/ui"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"maps"
	"strconv"
	"strings"
)

// tabNames are the party inventory's tabs: all, then by front.ItemKind.
var tabNames = [...]string{"All", "Weapons", "Armor", "Misc"}

// shown is the part of items the pane's tab shows.
func (pn *pane) shown(items []front.InvItem) []front.InvItem {
	kinds := [...]front.ItemKind{0, front.KindWeapon, front.KindArmor, front.KindMisc}
	if pn.tab == 0 {
		return items
	}
	var out []front.InvItem
	for _, it := range items {
		if it.Kind == kinds[pn.tab] {
			out = append(out, it)
		}
	}
	return out
}

// drawTabs draws the pane's tabs in its title row from x.
func (inv *inventory) drawTabs(pn *pane, x int, busy bool) {
	for i, t := range tabNames {
		pn.tabs[i] = image.Rect(x-2, 6, x+8*len(t)+2, 18)
		switch {
		case i == pn.tab:
			inv.text(t, x, 8, white)
		case inv.hover.In(pn.tabs[i]) && !busy:
			inv.text(t, x, 8, green)
		default:
			inv.text(t, x, 8, gray)
		}
		x += 8*len(t) + 8
	}
}

// pane is a scrolled list of items: the party inventory or the shop's goods.
type pane struct {
	top, rows, n int             // the first line shown, the lines shown, the lines
	seen         int             // the highlight the list last scrolled to
	at           image.Rectangle // the lines' area
	scroll       image.Rectangle // the scroll bar
	tab          int             // its tab (tabNames)
	tabs         [len(tabNames)]image.Rectangle
}

// clamp keeps the lines shown inside the list.
func (pn *pane) clamp() { pn.top = max(0, min(pn.top, pn.n-pn.rows)) }

// drawPane draws items as lines of icon and name (and price) from x0 to x1,
// listTop to bottom, the highlighted one (sel) in blue, with a scroll bar
// when they do not fit.
func (inv *inventory) drawPane(pn *pane, items []front.InvItem, sel int, ware, priced, busy bool, x0, x1, bottom int) {
	pn.n, pn.rows = len(items), max(1, (bottom-listTop)/rowH)
	for i, it := range items { // keep a newly highlighted line in sight (the wheel scrolls freely)
		if it.Index == sel && sel != pn.seen {
			pn.top = min(pn.top, i)
			pn.top = max(pn.top, i-pn.rows+1)
		}
	}
	pn.seen = sel
	pn.clamp()
	tr := x1 // the lines' right edge: the scroll bar's left
	if pn.n > pn.rows {
		tr -= 10
	}
	pn.at = image.Rect(x0, listTop, tr, listTop+pn.rows*rowH)
	pw := 0 // the prices' column
	if priced {
		for _, it := range items {
			pw = max(pw, 8*len(price(it.Price))+8)
		}
	}
	chars := (tr - x0 - rowH - 4 - pw) / 8
	for i := pn.top; i < pn.n && i < pn.top+pn.rows; i++ {
		it := &items[i]
		y := listTop + (i-pn.top)*rowH
		r := image.Rect(x0, y, tr, y+rowH)
		if it.Index == sel || inv.hover.In(r) && !busy {
			if inkBlue >= 0 {
				inv.fill(r, col(inkBlue)) // the text selection
			}
		}
		inv.drawIcon(it, image.Rect(x0, y, x0+rowH, y+rowH))
		tc := green
		if it.Wrong {
			tc = red
		}
		inv.text(upTo(readyMark(it)+it.Name, chars), x0+rowH+4, y+(rowH-8)/2, tc)
		if priced {
			m := price(it.Price)
			inv.text(m, tr-4-8*len(m), y+(rowH-8)/2, col(inkYellow))
		}
		inv.lines = append(inv.lines, hit{r: r, item: *it, ware: ware, whole: len(strings.TrimSpace(it.Name)) <= chars})
	}
	if pn.n == 0 {
		inv.text("Nothing", x0+rowH+4, listTop+(rowH-8)/2, dark)
	}
	pn.scroll = image.Rectangle{}
	if pn.n > pn.rows {
		pn.scroll = image.Rect(x1-6, listTop, x1, listTop+pn.rows*rowH)
		ui.ScrollBar{R: pn.scroll, N: pn.n, Rows: pn.rows, Top: pn.top}.Draw(inv.fill, barColors())
	}
}

// price is every price and purse the screens show: "12g", with " 5c" only
// when there is copper.
func price(m front.Money) string {
	s := strconv.Itoa(m.Gold) + "g"
	if m.Copper != 0 {
		s += " " + strconv.Itoa(m.Copper) + "c"
	}
	return s
}

// withGold is a price list's title with the party's gold on the box's
// bottom border (ui.InkFoot).
func withGold(title string, gold int) string {
	return title + string(rune(ui.InkFoot)) + " Party " + price(front.Money{Gold: gold}) + " "
}

// wareAt is the shop's ware with Index i, nil if none.
func wareAt(in *front.Inventory, i int) *front.InvItem {
	for k := range in.Wares {
		if in.Wares[k].Index == i {
			return &in.Wares[k]
		}
	}
	return nil
}

// decodePNG decodes b; empty or bad data gives a 1x1 clear image.
func decodePNG(b []byte) image.Image {
	im, err := png.Decode(bytes.NewReader(b))
	if err != nil { // no art: a clear pixel, so everything drawn from it is nothing
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	return im
}

func (inv *inventory) load() {
	if inv.doll != nil {
		return
	}
	inv.doll = decodePNG(dollPNG)
	o := sprites.Origins()
	inv.icons = o["por"]["ITEMS"] // the pool's pictures were imported from POR: its types are the default
	if own := o[classic.GameID]["ITEMS"]; own != nil && classic.GameID != "por" {
		inv.icons = maps.Clone(inv.icons)
		maps.Copy(inv.icons, own)
	}
}

// icon is the pool item picture of an item ("": none): origins.json's ITEMS
// key "TT:word" for one of its name words (or a word of one), else "TT" for
// its type, the game's own entries over POR's.
func (inv *inventory) icon(it *front.InvItem) string {
	inv.load()
	t := fmt.Sprintf("%02X", it.Type)
	for _, w := range it.Words {
		if n, ok := inv.icons[t+":"+w]; ok {
			return n
		}
		for _, f := range strings.Fields(w) {
			if n, ok := inv.icons[t+":"+f]; ok {
				return n
			}
		}
	}
	return inv.icons[t]
}

// iconImage is the item's pool picture (nil: none).
func (inv *inventory) iconImage(it *front.InvItem) image.Image {
	if n := inv.icon(it); n != "" {
		return sprites.Image(series, "items", n)
	}
	return nil
}

// drawIcon draws the item's picture, unscaled, in the middle of r (cut to r).
func (inv *inventory) drawIcon(it *front.InvItem, r image.Rectangle) {
	im := inv.iconImage(it)
	if im == nil {
		inv.text("?", r.Min.X+r.Dx()/2-4, r.Min.Y+r.Dy()/2-4, gray)
		return
	}
	sz := im.Bounds().Size()
	at := r.Min.Add(r.Size().Sub(sz).Div(2))
	cell := image.Rectangle{at, at.Add(sz)}.Intersect(r)
	draw.Draw(inv.px, cell, im, im.Bounds().Min.Add(cell.Min.Sub(at)), draw.Over)
}

// text draws s in the game's font at (x, y).
func (inv *inventory) text(s string, x, y int, c color.RGBA) {
	ui.DrawText(inv.px, inv.glyph, s, x, y, c)
}

func (inv *inventory) fill(r image.Rectangle, c color.RGBA) {
	draw.Draw(inv.px, r, image.NewUniform(c), image.Point{}, draw.Src)
}

// upTo is s in at most n characters.
func upTo(s string, n int) string { return s[:min(len(s), max(0, n))] }

// gold is a price in a question's title, in gold (ui.InkGold).
func gold(n int) string {
	return string(rune(ui.InkGold)) + price(front.Money{Gold: n}) + string(rune(ui.InkGold))
}

// readyMark is Data.ReadyMark before a readied item's name ("": none).
func readyMark(it *front.InvItem) string {
	if it.Readied {
		return readiedMark
	}
	return ""
}
