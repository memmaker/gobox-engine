// cluebook.go is the clue book (B while exploring): its chapter list over
// the 3D view, a chapter read on to the book's end in the pager, then the
// list again. B opens the book again on the page it was left on, unless it
// was read to its end. The engine gives the text (ClueChapters, ClueRead).

package ui

import (
	"image"
	"slices"

	"gobox/gpu"
)

// actClueBook is the Clue Book command's: the presenter's own (cluebook.go).
const actClueBook ActKind = -5

// clueUI is the clue book's chapter list up.
type clueUI struct {
	sel, top int
	at       []image.Rectangle // the rows shown, on the window
	pointer  image.Point
}

// openClue opens the clue book: on the page last left, else the list.
func (p *Presenter) openClue() {
	if p.eng == nil || !p.exploring() || p.modal() {
		return
	}
	if r := p.clueRead; r != nil && !p.clueEnd {
		p.readClue(r)
		return
	}
	p.clue = &clueUI{}
}

// readClue has r read in the pager from its page; leaving it keeps the
// page and brings the list back.
func (p *Presenter) readClue(r *Read) {
	p.clue, p.clueRead = nil, r
	p.pager = &pager{Pager: Pager{P: r.Page}, read: r, left: func(page int) {
		r.Page = page
		pg, _ := p.eng.JournalPage(r.Title, r.Texts, 0, 0, p.pagerCells())
		p.clueEnd = page >= pg.Pages
		p.clue = &clueUI{sel: p.clueSel}
	}}
	p.speak()
}

// clueRows are the list's rows: the chapters, then Exit.
func (p *Presenter) clueRows() []string {
	return append(p.eng.ClueChapters(), "Exit")
}

// clueKeys are the rows' hotkeys: Exit's E only.
func clueKeys(n int) []byte {
	k := make([]byte, n)
	k[n-1] = 'E'
	return k
}

// cluePick takes row i: a chapter is read, Exit closes the book.
func (p *Presenter) cluePick(i int) {
	rows := p.clueRows()
	switch {
	case i < 0 || i >= len(rows):
	case i == len(rows)-1:
		p.clue = nil
	default:
		p.clueSel, p.clueEnd = i, false
		p.readClue(p.eng.ClueRead(i))
	}
}

// clueKey takes key k on the list (menuNav; Esc closes).
func (p *Presenter) clueKey(k uint16) {
	c := p.clue
	rows := p.clueRows()
	switch i, back, _ := MenuNav(actOf(k), k, &c.sel, clueKeys(len(rows))); {
	case back:
		p.clue = nil
	case i >= 0:
		p.cluePick(i)
	}
}

// clueMouse takes the pointer: it highlights the row under it, a click
// reads it, the wheel moves, the right button closes.
func (p *Presenter) clueMouse(pt image.Point, left, right bool, wheel int) {
	c := p.clue
	i := slices.IndexFunc(c.at, pt.In)
	switch {
	case right:
		p.clue = nil
	case wheel != 0:
		c.sel = min(max(c.sel-unit(wheel), 0), len(p.clueRows())-1)
	case left && i >= 0:
		p.cluePick(c.top + i)
	case left:
		p.clue = nil
	case i >= 0 && pt != c.pointer:
		c.sel = c.top + i
	}
	c.pointer = pt
}

// drawClue draws the list over the view: as many rows as fit, scrolled to
// the highlight, a scroll bar when they don't all fit.
func (p *Presenter) drawClue(dst *gpu.Image) {
	c := p.clue
	rows := p.clueRows()
	keys := clueKeys(len(rows))
	fit := max(3, dst.Bounds().Dy()/int(8*max(1, scale(dst)))-6)
	c.sel = min(max(c.sel, 0), len(rows)-1)
	c.top = min(max(c.top, c.sel-fit+1), c.sel, max(0, len(rows)-fit))
	end := min(len(rows), c.top+fit)
	k := p.keyWindow(dst, "Clue Book", rows[c.top:end], keys[c.top:end], nil, c.sel-c.top, false)
	c.at = slices.Clone(p.winRects)
	p.menuBar(dst, p.winRects, len(rows), end-c.top, &c.top, &c.sel, k) // more chapters than fit: the bar
}
