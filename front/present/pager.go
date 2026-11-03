// pager.go is the presenter's pager: the keyboard help (HelpLines) and
// the journal text the game has the player read (View.Read), a page of
// the whole window at a time.

package present

import (
	"fmt"
	"gobox/front"
	"strconv"

	"gobox/input"
	"gobox/ui"
)

// pager is the pager up: the help, or the engine's Read.
type pager struct {
	ui.Pager                // the page P, and how far its tall picture is scrolled (Off)
	read     *front.Read    // the text being read (nil: the help); closing answers it with the page reached
	left     func(page int) // the presenter's own read (the clue book): closing calls it, the engine waits for nothing
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

// openHelp puts the help up over the screen, unless a line is being
// typed or a page over the whole window is up already.
func (p *Presenter) openHelp(v front.View) {
	if typing(v) != nil || p.modal(v) {
		return
	}
	p.pager = &pager{}
}

// syncRead opens the pager on a Read the engine waits for, and drops one
// the engine no longer waits for.
func (p *Presenter) syncRead(v front.View) {
	if pg := p.pager; pg != nil && pg.read != nil && pg.left == nil && pg.read != v.Read {
		p.pager = nil
	}
	if r := v.Read; r != nil && (p.pager == nil || p.pager.read != r) {
		p.pager = &pager{Pager: ui.Pager{P: r.Page}, read: r}
		p.speak()
	}
}

// speak voices the journal page up (the engine's SpeakPage).
func (p *Presenter) speak() {
	if r, n := p.pager.read, p.pager.P; r != nil {
		p.call(func(e front.Game) { e.SpeakPage(r.Lines, n) })
	}
}

// pages is the pager's page as drawn, and how far its tall picture
// scrolls (0: none); nil with no pager up.
func (p *Presenter) pages(v front.View) (*front.Pages, int) {
	if pg := p.pager; pg != nil {
		return p.pagesAt(v, pg.P, pg.Off)
	}
	return nil, 0
}

// pagesAt is page n of the pager's text, its tall picture scrolled off.
func (p *Presenter) pagesAt(v front.View, n, off int) (*front.Pages, int) {
	pg := p.pager
	switch {
	case pg == nil:
		return nil, 0
	case pg.read != nil && p.eng != nil:
		return p.eng.JournalPage(pg.read.Title, pg.read.Lines, n, off)
	case pg.read != nil:
		return nil, 0
	}
	return helpPage(v, n), 0
}

// pageSource is the pager's text for ui.Pager: front.Pages.Page is the
// page reached, so the clamp holds.
type pageSource struct {
	p *Presenter
	v front.View
}

func (s pageSource) Page(n, off int) (page, pages, maxOff int, ok bool) {
	pg, maxOff := s.p.pagesAt(s.v, n, off)
	if pg == nil {
		return 0, 0, 0, false
	}
	return pg.Page, pg.Pages, maxOff, true
}

// pagerTurn turns the page d (+1 on, -1 back); a tall picture scrolls
// first. Past the last page the pager closes.
func (p *Presenter) pagerTurn(v front.View, d int) {
	if pg, _ := p.pages(v); pg == nil {
		return
	}
	if p.pager.Turn(pageSource{p, v}, d) {
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
		p.send(front.ActOK, pg.P)
	}
}

// pagerKey takes key k on the pager: Esc closes it, the cursor keys back
// page (or scroll) back, every other key on.
func (p *Presenter) pagerKey(k uint16, v front.View) {
	a := actOf(k)
	if kp := keypad(k); kp != input.None {
		a = kp
	}
	closed, turned := p.pager.Key(a, pageSource{p, v})
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

// helpPage is page n of the help (HelpLines) on v's page: headings in
// white, keys in yellow, descriptions in light green.
func helpPage(v front.View, n int) *front.Pages {
	pages := ui.HelpPages(HelpLines(), helpRows)
	n = max(min(n, len(pages)-1), 0)
	cols := max(v.Cells.X, 40)
	run := func(x, c int, s string) front.PaneRun {
		s = s[:max(0, min(len(s), cols-x))]
		return front.PaneRun{X: x, Text: s, Fg: v.Ink[c], Bg: v.Ink[0], Upcase: true}
	}
	const title = "Help"
	pg := &front.Pages{Title: run((cols-1-len(title))/2, 15, title), Top: 2, Page: n, Pages: len(pages)}
	for _, l := range pages[n] {
		if l[1] == "" {
			pg.Lines = append(pg.Lines, []front.PaneRun{run(2, 15, l[0])})
			continue
		}
		pg.Lines = append(pg.Lines, []front.PaneRun{run(3, 14, l[0]), run(helpCol, 10, l[1])})
	}
	if len(pages) > 1 {
		more := "page " + strconv.Itoa(n+1) + "/" + strconv.Itoa(len(pages))
		pg.Hint = run(cols-3-len(more), 10, more)
	}
	return pg
}
