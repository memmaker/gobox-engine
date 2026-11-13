package ui

import (
	"hash/fnv"
	"image"
	"image/color"
	"strconv"
	"strings"

	"gobox/asset"
	"gobox/ega"
	"gobox/game"
)

// Picker is page page of the grid picker in s (page < 0: the one holding
// the picture it opened on; nil: none up) on a page of cols by rows cells:
// as many columns and rows as the box holds, centred, the rest on further
// pages with a pageCut line at the bottom.
func Picker(k *game.Picker, cells image.Point, page, cols, rows int) *FrontPicker {
	n := 0
	if k != nil {
		n = len(k.Pics)
	}
	if n == 0 || k.Cell.X < 1 || k.Cell.Y < 1 {
		return nil
	}
	cols, rows = shownPage(cells)
	textCols, textRows := max(cols, 40), max(rows, 25)
	winRight, bot := textCols-2, textRows-2 // bot: the box's inner bottom, the page's last row holds its border
	cw, ch := k.Cell.X, k.Cell.Y
	if len(k.Labels) > 0 { // a roster: the portrait (11 by 5 cells) and the name beside it fit the cell (present's picker)
		ch = max(ch, 5)
		for _, l := range k.Labels {
			cw = max(cw, 12+len(l))
		}
	}
	gap, nc := 1, winRight/(cw+1)
	if winRight/cw > nc { // a blank between the columns only when it costs none
		gap, nc = 0, winRight/cw
	}
	nc = max(1, min(nc, n))
	nr, pageCut := max(1, bot/ch), 0
	if (n+nc-1)/nc > nr { // paged: the pageCut line takes the box's last row
		nr, pageCut = max(1, (bot-1)/ch), bot
	}
	nr = min(nr, (n+nc-1)/nc)
	x0, y0 := 1+(winRight-(nc*(cw+gap)-gap))/2, 1+(bot-pageCut/bot-nr*ch)/2
	per := nc * nr
	pages := max(1, (n+per-1)/per)
	cur := min(max(k.Cur, 0), n-1)
	if page < 0 {
		page = cur / per
	}
	page = min(max(page, 0), pages-1)
	v := &FrontPicker{N: n, Page: page, Pages: pages, PerPage: per, Cols: nc, Cur: cur, Bottom: bot, Labels: k.Labels, Ask: k.Ask,
		Classes: k.Classes, Portraits: k.Portraits}
	if len(k.Labels) > 0 {
		v.Title = "Saved Characters"
	}
	if len(k.Labels) > 0 && len(k.Portraits) == 0 { // the game shows no faces: gobox's default by name, as a member without one
		v.Portraits = make([]image.Image, len(k.Labels))
		for i, l := range k.Labels {
			v.Portraits[i] = poolPicture("portraits", poolPickName(game.Member{Name: l}, "portraits", "head_"))
		}
	}
	first := page * per
	for i := first; i < n && i < first+per; i++ {
		x, y := x0+i%nc*(cw+gap), y0+(i/nc-page*nr)*ch
		v.Cells = append(v.Cells, FrontCellRect{X1: x, Y1: y, X2: x + cw - 1, Y2: y + ch - 1})
		if p := k.Pics[i]; p.Pixels != nil && p.W > 0 {
			f := pix(p)
			f.File, f.Block, f.Frame, f.Name = p.File, p.Block, p.Frame, p.Name
			f.At = image.Rect(x*8, y*8, x*8+p.W, y*8+p.H)
			v.Pics = append(v.Pics, f)
		}
	}
	if pageCut > 0 {
		t := "< page " + strconv.Itoa(page+1) + "/" + strconv.Itoa(pages) + " >"
		x := (textCols - len(t)) / 2
		h := len(t) / 2 // each half of the pageCut line pages, not only its < and > cells
		v.Prev, v.Next = FrontCellRect{X1: x, Y1: pageCut, X2: x + h - 1, Y2: pageCut}, FrontCellRect{X1: x + h, Y1: pageCut, X2: x + len(t) - 1, Y2: pageCut}
	}
	return v
}

// shownPage is the page the game shows, whatever page was asked for: its Cells,
// else the original's 40 by 25 when it keeps its own (it never took the
// resize, so a wider layout runs off the page it draws).
func shownPage(c image.Point) (int, int) {
	if c.X > 0 && c.Y > 0 {
		return c.X, c.Y
	}
	return 40, 25
}

// IconEditor is the combat icon designer's page from s (nil: none up) on a
// page of cols by rows cells, its poses and swatches in ink.
func IconEditor(k *game.IconEditor, palette []color.RGBA, own game.Member, ink *[16]color.RGBA, cols, rows int) *FrontIconEditor {
	if k == nil {
		return nil
	}
	winRight, winBottom := max(cols, 40)-2, max(rows, 25)-3
	sc := max(1, min((winRight-16)/18, (winBottom-13)/9)) // the poses' whole scale
	const previewRow, previewCols, partCol, swatchCol, paletteCol = 2, 2, 2, 12, 2
	wc, pr := previewCols+2*(9*sc+2), previewRow+9*sc+2
	cell := func(x1, y1, x2, y2 int) FrontCellRect { return FrontCellRect{X1: x1, Y1: y1, X2: x2, Y2: y2} }
	ed := &FrontIconEditor{Large: k.Large, Custom: k.Custom,
		Head: cell(wc, previewRow, wc+5, previewRow), Weapon: cell(wc, previewRow+1, wc+5, previewRow+1),
		Size:  cell(wc, previewRow+2, wc+9, previewRow+2),
		Names: cell(partCol, pr, swatchCol-1, pr+5), Parts: cell(swatchCol, pr, swatchCol+4, pr+5),
		Palette: cell(paletteCol, pr+7, paletteCol+15, pr+8)}
	if *ink == ([16]color.RGBA{}) { // unset: the EGA's own (Ink)
		var egaInk [16]color.RGBA
		for i := range egaInk {
			egaInk[i] = ega.RGBA(i)
		}
		ink = &egaInk
	}
	look := func(n int) color.RGBA {
		if len(palette) == 16 {
			return palette[n&15]
		}
		return IndexedRGBA([]uint8{k.Looks[n&15]}, 1, -1, ink).RGBAAt(0, 0)
	}
	m := subject(own)
	var cs []byte
	for _, c := range k.Colour {
		cs = append(cs, byte(c[0]&15|c[1]<<4))
	}
	for p := range 2 {
		x := previewCols + (9*sc+2)*p
		ed.PoseAt[p] = cell(x, previewRow, x+9*sc-1, previewRow+9*sc-1)
		head, body := memberIcon(m)
		if pic, ok := poolIcon(head, body, cs, p, false); ok && !k.Custom && head != "" {
			if pic.Palette != nil {
				ed.Poses[p] = PalettedRGBA(pic.Pix, pic.PixW, pic.Palette)
			} else {
				ed.Poses[p] = IndexedRGBA(pic.Pix, pic.PixW, -1, ink)
			}
		} else if k.Custom {
			ed.Poses[p] = k.Icons[p]
		} else if po := k.Poses[p]; po.Pixels != nil && po.W > 0 {
			if po.Palette != nil {
				ed.Poses[p] = PalettedRGBA(po.Pixels, po.W, po.Palette)
			} else {
				ed.Poses[p] = IndexedRGBA(po.Pixels, po.W, po.Clear, ink)
			}
		}
	}
	for i, n := range k.Cells {
		ed.Colours[i] = n
		if n >= 0 {
			ed.PaletteInk[i] = look(n)
		}
	}
	for i, c := range k.Colour {
		ed.Colour[i] = c
		for s := range 2 {
			ed.Swatch[i][s] = look(c[s])
		}
	}
	if ed.Custom { // a dropped PNG takes no parts or colours: it can only go
		ed.Remove = cell(wc, previewRow, wc+5, previewRow)
		ed.Head, ed.Weapon, ed.Size, ed.Names, ed.Parts, ed.Palette = FrontCellRect{}, FrontCellRect{}, FrontCellRect{}, FrontCellRect{}, FrontCellRect{}, FrontCellRect{}
	}
	return ed
}

// Title is the title picture under the party menu: art's "title/logo"
// (the picture's pixels), else logo (the game's).
func Title(logo *Picture, art map[string]game.Picture) *Picture {
	if p := pix(art["title/logo"]); p.Pix != nil {
		logo = &p
	}
	return logo
}

// subject is the member whose portrait and icon are on show: the creation's
// draft, else the selected one, with gobox's defaults where it names none.
func subject(own game.Member) game.Member { return defaultIcon(defaultPortrait(own)) }

// OwnSubject is subject as the rules state it, without the defaults.
func OwnSubject(mode game.Mode, c *game.Creation, party []game.Member, sel int) game.Member {
	switch {
	case Drafting(mode, c):
		return c.Draft
	case sel >= 0 && sel < len(party):
		return party[sel]
	}
	return game.Member{}
}

// pictured: a picture of the game's (not a fill) already covers part of r.
func pictured(pics []Picture, r image.Rectangle) bool {
	for _, pc := range pics {
		if !pc.Fill && pc.At.Overlaps(r) {
			return true
		}
	}
	return false
}

// defaultPortrait is m with a pool head and body filled in where it names none,
// picked by its race and class so the same kind of character gets the same
// face; the game never chooses one.
func defaultPortrait(m game.Member) game.Member {
	if m.PortraitHeadID != "" && m.PortraitBodyID != "" || m.Portrait.File != "" || m.Portrait.Pixels != nil {
		return m
	}
	if m.PortraitHeadID == "" {
		m.PortraitHeadID = poolPickName(m, "portraits", "head_")
	}
	if m.PortraitBodyID == "" {
		m.PortraitBodyID = poolPickName(m, "portraits", "body_")
	}
	return m
}

// defaultIcon is m with a pool combat head and body filled in when it has
// neither pool ids nor original icon numbers, picked like defaultPortrait.
func defaultIcon(m game.Member) game.Member {
	if m.IconHeadID != "" || m.Icon != nil {
		return m
	}
	m.IconHeadID, m.IconBodyID = poolPickName(m, "combat", "combat_head_"), poolPickName(m, "combat", "combat_body_")
	return m
}

// poolPick is one kind/prefix pool name chosen by m's race and class ("": none).
func poolPickName(m game.Member, kind, prefix string) string {
	h := fnv.New32a()
	h.Write([]byte(m.Race + "/" + m.Class))
	if m.Race == "" && m.Class == "" { // nothing known but the name (a saved character's): by name
		h.Write([]byte(m.Name))
	}
	var ns []string
	for _, n := range asset.SpriteNames(Series, kind) {
		if strings.HasPrefix(n, prefix) {
			ns = append(ns, n)
		}
	}
	if len(ns) == 0 {
		return ""
	}
	return ns[int(h.Sum32()%uint32(len(ns)))]
}
