package classic

import (
	"image"
	"image/color"
	"strconv"

	"gobox/front"
	"gobox/game"
)

// Picker is page page of the grid picker in s (page < 0: the one holding
// the picture it opened on; nil: none up) on a page of cols by rows cells:
// as many columns and rows as the box holds, centred, the rest on further
// pages with a pager line at the bottom.
func Picker(s game.State, page, cols, rows int) *front.Picker {
	k := s.Picker
	n := 0
	if k != nil {
		n = len(k.Pics)
	}
	if n == 0 || k.Cell.X < 1 || k.Cell.Y < 1 {
		return nil
	}
	textCols, textRows := max(cols, 40), max(rows, 25)
	winRight, bot := textCols-2, textRows-2 // bot: the box's inner bottom, the page's last row holds its border
	cw, ch := k.Cell.X, k.Cell.Y
	gap, nc := 1, winRight/(cw+1)
	if winRight/cw > nc { // a blank between the columns only when it costs none
		gap, nc = 0, winRight/cw
	}
	nc = max(1, min(nc, n))
	nr, pager := max(1, bot/ch), 0
	if (n+nc-1)/nc > nr { // paged: the pager line takes the box's last row
		nr, pager = max(1, (bot-1)/ch), bot
	}
	nr = min(nr, (n+nc-1)/nc)
	x0, y0 := 1+(winRight-(nc*(cw+gap)-gap))/2, 1+(bot-pager/bot-nr*ch)/2
	per := nc * nr
	pages := max(1, (n+per-1)/per)
	cur := min(max(k.Cur, 0), n-1)
	if page < 0 {
		page = cur / per
	}
	page = min(max(page, 0), pages-1)
	v := &front.Picker{N: n, Page: page, Pages: pages, PerPage: per, Cols: nc, Cur: cur, Bottom: bot, Labels: k.Labels, Ask: k.Ask,
		Classes: k.Classes, Portraits: k.Portraits}
	first := page * per
	for i := first; i < n && i < first+per; i++ {
		x, y := x0+i%nc*(cw+gap), y0+(i/nc-page*nr)*ch
		v.Cells = append(v.Cells, front.CellRect{X1: x, Y1: y, X2: x + cw - 1, Y2: y + ch - 1})
		if p := k.Pics[i]; p.Pixels != nil && p.W > 0 {
			f := pix(p)
			f.File, f.Block, f.Frame, f.Name = p.File, p.Block, p.Frame, p.Name
			f.At = image.Rect(x*8, y*8, x*8+p.W, y*8+p.H)
			v.Pics = append(v.Pics, f)
		}
	}
	if pager > 0 {
		t := "< page " + strconv.Itoa(page+1) + "/" + strconv.Itoa(pages) + " >"
		x := (textCols - len(t)) / 2
		v.Prev, v.Next = front.CellRect{X1: x, Y1: pager, X2: x, Y2: pager}, front.CellRect{X1: x + len(t) - 1, Y1: pager, X2: x + len(t) - 1, Y2: pager}
	}
	return v
}

// IconEditor is the combat icon designer's page from s (nil: none up) on a
// page of cols by rows cells, its poses and swatches in ink.
func IconEditor(s game.State, ink *[16]color.RGBA, cols, rows int) *front.IconEditor {
	k := s.IconEditor
	if k == nil {
		return nil
	}
	winRight, winBottom := max(cols, 40)-2, max(rows, 25)-3
	sc := max(1, min((winRight-16)/18, (winBottom-13)/9)) // the poses' whole scale
	const previewRow, previewCols, partCol, swatchCol, paletteCol = 2, 2, 2, 12, 2
	wc, pr := previewCols+2*(9*sc+2), previewRow+9*sc+2
	cell := func(x1, y1, x2, y2 int) front.CellRect { return front.CellRect{X1: x1, Y1: y1, X2: x2, Y2: y2} }
	ed := &front.IconEditor{Large: k.Large, Custom: k.Custom,
		Head: cell(wc, previewRow, wc+5, previewRow), Weapon: cell(wc, previewRow+1, wc+5, previewRow+1),
		Size:  cell(wc, previewRow+2, wc+9, previewRow+2),
		Names: cell(partCol, pr, swatchCol-1, pr+5), Parts: cell(swatchCol, pr, swatchCol+4, pr+5),
		Palette: cell(paletteCol, pr+7, paletteCol+15, pr+8)}
	look := func(n int) color.RGBA { return front.IndexedRGBA([]uint8{k.Looks[n&15]}, 1, -1, ink).RGBAAt(0, 0) }
	m := subject(s)
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
				ed.Poses[p] = front.PalettedRGBA(pic.Pix, pic.PixW, pic.Palette)
			} else {
				ed.Poses[p] = front.IndexedRGBA(pic.Pix, pic.PixW, -1, ink)
			}
		} else if k.Custom {
			ed.Poses[p] = k.Icons[p]
		} else if po := k.Poses[p]; po.Pixels != nil && po.W > 0 {
			ed.Poses[p] = front.IndexedRGBA(po.Pixels, po.W, po.Clear, ink)
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
		ed.Head, ed.Weapon, ed.Size, ed.Names, ed.Parts, ed.Palette = front.CellRect{}, front.CellRect{}, front.CellRect{}, front.CellRect{}, front.CellRect{}, front.CellRect{}
	}
	return ed
}

// Title is the title picture and the painted backdrop under the party
// menu: art's "title/logo" (the picture's pixels) and "title/backdrop"
// (File: the PNG), else logo and backdrop (the game's).
func Title(logo, backdrop *front.Picture, art map[string]game.Picture) (*front.Picture, *front.Picture) {
	if p := pix(art["title/logo"]); p.Pix != nil {
		logo = &p
	}
	if p := art["title/backdrop"]; p.File != "" {
		backdrop = &front.Picture{File: p.File}
	}
	return logo, backdrop
}

// subject is the member whose portrait and icon are on show: the creation's draft, else the selected one.
func subject(s game.State) game.Member {
	switch {
	case s.Creation != nil:
		return s.Creation.Draft
	case s.Selected >= 0 && s.Selected < len(s.Party):
		return s.Party[s.Selected]
	}
	return game.Member{}
}
