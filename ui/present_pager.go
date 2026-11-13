// pager.go is the presenter's pager: the keyboard help (HelpLines) and
// the journal text the game has the player read (Scene.Handout), a page of
// the whole window at a time.

package ui

import (
	"fmt"
	"gobox/game"
	"gobox/ui/input"
	"strconv"
)

// pager is the pager up: the help, or the engine's Read.
type pager struct {
	Pager                // the page P, and how far its tall picture is scrolled (Off)
	read  *game.Read     // the text being read (nil: the help); closing answers it with the page reached
	left  func(page int) // the presenter's own read (the clue book): closing calls it, the engine waits for nothing
}

// pageKey names what of the presenter's own screens the page shows: a
// change redraws it.
func (p *Presenter) pageKey() string {
	if pg := p.pager; pg != nil {
		return p.Screens() + fmt.Sprint(pg.P, pg.Off)
	}
	if f := p.fullMap; f != nil && f.label != nil {
		return p.Screens() + fmt.Sprint(f.cursor, *f.label, f.cur)
	}
	if f := p.fullMap; f != nil {
		return p.Screens() + fmt.Sprint(f.cursor)
	}
	return p.Screens()
}

// gridCols is the page's width in cells (40 at least; 40 with no game).
func (p *Presenter) gridCols() int {
	if u := p.ui(); u != nil {
		return max(u.grid.X, 40)
	}
	return 40
}

// openHelp puts the help up over the screen, unless a line is being
// typed or a page over the whole window is up already.
func (p *Presenter) openHelp() {
	if p.typingAsk() != nil || p.modal() {
		return
	}
	p.pager = &pager{}
}

// syncRead opens the pager on a Read the engine waits for, and drops one
// the engine no longer waits for.
func (p *Presenter) syncRead() {
	var read *game.Read
	if p.eng != nil {
		read = p.eng.r.readOf(p.state())
	}
	if pg := p.pager; pg != nil && pg.read != nil && pg.left == nil && pg.read != read {
		p.pager = nil
	}
	if r := read; r != nil && (p.pager == nil || p.pager.read != r) {
		p.pager = &pager{Pager: Pager{P: r.Page}, read: r}
		p.speak()
	}
}

// speak voices the journal page up (the engine's SpeakPage).
func (p *Presenter) speak() {
	if r, n := p.pager.read, p.pager.P; r != nil {
		p.call(func(e *classicGame) { e.SpeakPage(r.Texts, n) })
	}
}

// pages is the pager's page as drawn, and how far its tall picture
// scrolls (0: none); nil with no pager up.
func (p *Presenter) pages() (*Pages, int) {
	if pg := p.pager; pg != nil {
		return p.pagesAt(pg.P, pg.Off)
	}
	return nil, 0
}

// pagesAt is page n of the pager's text, its tall picture scrolled off.
func (p *Presenter) pagesAt(n, off int) (*Pages, int) {
	pg := p.pager
	switch {
	case pg == nil:
		return nil, 0
	case pg.read != nil && p.eng != nil:
		return p.eng.JournalPage(pg.read.Title, pg.read.Texts, n, off)
	case pg.read != nil:
		return nil, 0
	}
	return helpPage(p.gridCols(), n), 0
}

// pageSource is the pager's text for ui.Pager: front.Pages.Page is the
// page reached, so the clamp holds.
type pageSource struct{ p *Presenter }

func (s pageSource) Page(n, off int) (page, pages, maxOff int, ok bool) {
	pg, maxOff := s.p.pagesAt(n, off)
	if pg == nil {
		return 0, 0, 0, false
	}
	return pg.Page, pg.Pages, maxOff, true
}

// pagerTurn turns the page d (+1 on, -1 back); a tall picture scrolls
// first. Past the last page the pager closes.
func (p *Presenter) pagerTurn(d int) {
	if pg, _ := p.pages(); pg == nil {
		return
	}
	if p.pager.Turn(pageSource{p}, d) {
		p.closePager()
		return
	}
	p.speak()
}

// closePager closes the pager; a Read is answered with the page it was
// left on.
func (p *Presenter) closePager() {
	pg := p.pager
	p.pager = nil
	switch {
	case pg.left != nil:
		pg.left(pg.P)
	case pg.read != nil:
		p.send(ActOK, pg.P)
	}
}

// pagerKey takes key k on the pager: Esc closes it, the cursor keys back
// page (or scroll) back, every other key on.
func (p *Presenter) pagerKey(k uint16) {
	a := actOf(k)
	if kp := keypad(k); kp != input.None {
		a = kp
	}
	closed, turned := p.pager.Key(a, pageSource{p})
	if closed {
		p.closePager()
	}
	if turned {
		p.speak()
	}
}

// Help geometry: the descriptions start at helpCol, helpRows lines fit
// above the page number.
const (
	helpCol  = 0x10
	helpRows = 0x14
)

// helpPage is page n of the help (HelpLines) on a page cols cells wide: headings in
// white, keys in yellow, descriptions in light green.
func helpPage(cols, n int) *Pages {
	pages := HelpPages(HelpLines(), helpRows)
	n = max(min(n, len(pages)-1), 0)
	run := func(x, c int, s string) PaneRun {
		s = s[:max(0, min(len(s), cols-x))]
		return PaneRun{X: x, Text: s, Fg: egaInk[c], Bg: egaInk[inkBlack]}
	}
	const title = "Help"
	pg := &Pages{Title: run((cols-1-len(title))/2, 15, title), Top: 2, Page: n, Pages: len(pages)}
	for _, l := range pages[n] {
		if l[1] == "" {
			pg.Lines = append(pg.Lines, []PaneRun{run(2, 15, l[0])})
			continue
		}
		pg.Lines = append(pg.Lines, []PaneRun{run(3, 14, l[0]), run(helpCol, 10, l[1])})
	}
	if len(pages) > 1 {
		more := "page " + strconv.Itoa(n+1) + "/" + strconv.Itoa(len(pages))
		pg.Hint = run(cols-3-len(more), 10, more)
	}
	return pg
}
