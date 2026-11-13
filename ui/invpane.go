// invpane.go holds the inventory's item lists: tabs, panes, prices, the item icons and the text and fill primitives they draw with.

package ui

import (
	"bufio"
	"bytes"
	"gobox/asset"
	"gobox/game"
	"gobox/rules"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strconv"
	"strings"

	xdraw "golang.org/x/image/draw"
)

// tabNames are the party inventory's tabs: all, then by front.ItemKind.
var tabNames = [...]string{"All", "Weapons", "Armor", "Misc"}

// shown is the part of items the pane's tab shows.
func (pn *pane) shown(items []game.Item) []game.Item {
	kinds := [...]ItemKind{0, KindWeapon, KindArmor, KindMisc}
	if pn.tab == 0 {
		return items
	}
	var out []game.Item
	for _, it := range items {
		if itemKinds[it.Kind] == kinds[pn.tab] {
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
	top, rows, n int              // the first line shown, the lines shown, the lines
	seen         rules.ItemHandle // the highlight the list last scrolled to
	at           image.Rectangle  // the lines' area
	scroll       image.Rectangle  // the scroll bar
	tab          int              // its tab (tabNames)
	tabs         [len(tabNames)]image.Rectangle
}

// clamp keeps the lines shown inside the list.
func (pn *pane) clamp() { pn.top = max(0, min(pn.top, pn.n-pn.rows)) }

// drawPane draws items as lines of icon and name (and price) from x0 to x1,
// listTop to bottom, the highlighted one (sel) in blue, with a scroll bar
// when they do not fit.
func (inv *inventory) drawPane(pn *pane, items []game.Item, sel rules.ItemHandle, ware, priced, busy bool, x0, x1, bottom int) {
	pn.n, pn.rows = len(items), max(1, (bottom-listTop)/rowH)
	for i, it := range items { // keep a newly highlighted line in sight (the wheel scrolls freely)
		if it.Handle == sel && sel != pn.seen {
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
			pw = max(pw, 8*len(price(pricing(it, ware)))+8)
		}
	}
	chars := (tr - x0 - rowH - 4 - pw) / 8
	for i := pn.top; i < pn.n && i < pn.top+pn.rows; i++ {
		it := &items[i]
		y := listTop + (i-pn.top)*rowH
		r := image.Rect(x0, y, tr, y+rowH)
		switch {
		case inkBlue < 0:
		case it.Handle == sel:
			inv.fill(r, col(inkBlue)) // the text selection
		case inv.hover.In(r) && !busy: // the row under the pointer: the selection's ink as an outline, never a second selection
			selectOutline(r, 1, func(r image.Rectangle, _ color.RGBA) { inv.fill(r, col(inkBlue)) })
		}
		inv.drawIcon(it.Name, it.Words, image.Rect(x0, y, x0+rowH, y+rowH))
		tc := green
		if it.Wrong {
			tc = red
		}
		if a, b := wrapName(readyMark(it.Readied)+strings.TrimSpace(it.Name), chars); b == "" {
			inv.text(a, x0+rowH+4, y+(rowH-8)/2, tc)
		} else { // too long for the line: two lines of the row
			inv.text(a, x0+rowH+4, y+(rowH-16)/2, tc)
			inv.text(b, x0+rowH+4, y+(rowH-16)/2+8, tc)
		}
		if priced && pricing(*it, ware) != (Money{}) { // no price given: none shown, not 0g
			m := price(pricing(*it, ware))
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
		ScrollBar{R: pn.scroll, N: pn.n, Rows: pn.rows, Top: pn.top}.Draw(inv.fill, barColors())
	}
}

// price is every price and purse the screens show: "12g", with " 5c" only
// when there is copper.
func price(m Money) string {
	s := strconv.Itoa(m.Gold) + "g"
	if Coin == game.Credits {
		s = strconv.Itoa(m.Gold) + "cr"
	}
	if m.Copper != 0 {
		s += " " + strconv.Itoa(m.Copper) + "c"
	}
	return s
}

// withGold is a price list's title with the party's gold on the box's
// bottom border (ui.InkFoot).
func withGold(title string, gold int) string {
	return title + string(rune(InkFoot)) + " Party " + price(Money{Gold: gold}) + " "
}

// pricing is the price the list shows for it: what a ware costs, or what the shop pays for the party's.
func pricing(it game.Item, ware bool) Money {
	if ware {
		return warePrice(it)
	}
	return salePrice(it)
}

// decodePNG decodes b; empty or bad data gives a 1x1 clear image.
func decodePNG(b []byte) image.Image {
	im, err := png.Decode(bytes.NewReader(b))
	if err != nil { // no art: a clear pixel, so everything drawn from it is nothing
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	return im
}

// readIcons parses items.rec: records of "field: value" lines between
// blank lines, comments starting with #.
func readIcons(rec string) map[string]int {
	m := map[string]int{}
	name := ""
	s := bufio.NewScanner(strings.NewReader(rec))
	for s.Scan() {
		k, v, ok := strings.Cut(s.Text(), ":")
		v = strings.TrimSpace(v)
		switch {
		case !ok || strings.HasPrefix(k, "#"):
		case k == "internal_name":
			name = strings.ToLower(v)
		case k == "icon":
			if n, err := strconv.Atoi(v); err == nil && name != "" {
				m[name] = n
			}
		}
	}
	return m
}

func (inv *inventory) load() {
	if inv.doll != nil {
		return
	}
	atlas, rec := asset.ItemIcons(series)
	inv.doll, inv.atlas = decodePNG(dollPNG), decodePNG(atlas)
	inv.icons = readIcons(rec)
}

// icon is the atlas cell of an item: the longest items.rec name its name
// (or one of its words) contains, ignoring case; -1 for none. Of two as
// long, the later in the name (the noun: a Staff Sling is a sling).
func (inv *inventory) icon(itemName string, words []string) int {
	name := strings.ToLower(itemName + "\n" + strings.Join(words, " "))
	n, best, at, bk := -1, 0, -1, ""
	for k, c := range inv.icons { // the longest; of two as long the later in the name (a Staff Sling is a sling), then the first in order: never the map's order
		i := strings.LastIndex(name, k)
		if i >= 0 && (len(k) > best || len(k) == best && (i > at || i == at && k < bk)) {
			n, best, at, bk = c, len(k), i, k
		}
	}
	return n
}

// iconImage is the item's atlas cell (nil: none), at the atlas's size:
// a cell is the atlas's width / atlasCols, whatever the series draws.
func (inv *inventory) iconImage(name string, words []string) image.Image {
	inv.load()
	n := inv.icon(name, words)
	if n < 0 || inv.atlas == nil {
		return nil
	}
	c := inv.atlas.Bounds().Dx() / atlasCols
	at := image.Pt(n%atlasCols*c, n/atlasCols*c)
	return inv.atlas.(interface {
		SubImage(image.Rectangle) image.Image
	}).SubImage(image.Rectangle{at, at.Add(image.Pt(c, c))})
}

// drawIcon draws the item's atlas cell scaled to iconSlot, in the middle of r (cut to r).
func (inv *inventory) drawIcon(name string, words []string, r image.Rectangle) {
	ic := inv.iconImage(name, words)
	if ic == nil {
		inv.text("?", r.Min.X+r.Dx()/2-4, r.Min.Y+r.Dy()/2-4, gray)
		return
	}
	at := r.Min.Add(image.Pt(r.Dx()-iconSlot, r.Dy()-iconSlot).Div(2))
	cell := image.Rectangle{at, at.Add(image.Pt(iconSlot, iconSlot))}
	m := image.NewRGBA(cell)
	xdraw.NearestNeighbor.Scale(m, cell, ic, ic.Bounds(), xdraw.Src, nil)
	draw.Draw(inv.px, cell.Intersect(r), m, cell.Intersect(r).Min, draw.Over)
}

// text draws s in the game's font at (x, y).
func (inv *inventory) text(s string, x, y int, c color.RGBA) {
	DrawText(inv.px, inv.glyph, s, x, y, c)
}

func (inv *inventory) fill(r image.Rectangle, c color.RGBA) {
	draw.Draw(inv.px, r, image.NewUniform(c), image.Point{}, draw.Src)
}

// wrapName splits a name too long for n characters at a space into two
// lines of at most n ("": it fits on one).
func wrapName(s string, n int) (string, string) {
	if len(s) <= n {
		return s, ""
	}
	i := strings.LastIndex(upTo(s, n+1), " ")
	if i <= 0 {
		return upTo(s, n), upTo(s[min(len(s), n):], n)
	}
	return s[:i], upTo(s[i+1:], n)
}

// upTo is s in at most n characters.
func upTo(s string, n int) string { return s[:min(len(s), max(0, n))] }

// gold is a price in a question's title, in gold (ui.InkGold).
func gold(n int) string {
	return string(rune(InkGold)) + price(Money{Gold: n}) + string(rune(InkGold))
}

// readyMark is Data.ReadyMark before a readied item's name ("": none).
func readyMark(readied bool) string {
	if readied {
		return readiedMark
	}
	return ""
}
