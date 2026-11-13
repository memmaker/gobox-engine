package ui

// PageSource is what a pager pages: page p of its text, with how far its
// tall picture is scrolled (off). It gives the page reached (p kept in
// range), how many there are, and the most off can scroll (0: none); ok
// false with no page to show.
type PageSource interface {
	Page(p, off int) (page, pages, maxOff int, ok bool)
}

// Pager is the state of a pager over the whole window: the page, and how
// far its tall picture is scrolled. What it shows, what closing answers
// and speaking the page stay with the game.
type Pager struct{ P, Off int }

// ScrollStep is how far Up/Down scroll a tall picture.
const ScrollStep = 16

// Turn turns the page d (+1 on, -1 back); a tall picture scrolls first.
// closed is true past the last page.
func (pg *Pager) Turn(src PageSource, d int) (closed bool) {
	page, pages, maxOff, ok := src.Page(pg.P, pg.Off)
	if !ok {
		return false
	}
	if maxOff < ScrollStep { // a picture a few pixels too tall pages like a short one
		maxOff = 0
	}
	switch {
	case maxOff > 0 && d < 0 && pg.Off > 0: // at the top, Up pages back
		pg.Off = max(pg.Off-ScrollStep, 0)
	case d < 0:
		pg.P, pg.Off = max(page-1, 0), 0
	case maxOff > 0 && pg.Off < maxOff: // at the bottom, Down pages on
		pg.Off = min(pg.Off+ScrollStep, maxOff)
	default:
		pg.P, pg.Off = page+1, 0
	}
	return pg.P >= pages
}

// Key takes act a (the key's Act, or its keypad digit) on the pager: Esc
// closes it, the cursor keys back page (or scroll) back, every other key
// on. closed is true when the pager is to close; turned when the page
// changed (to speak it).
func (pg *Pager) Key(a Act, src PageSource) (closed, turned bool) {
	switch a {
	case Esc:
		return true, false
	case Up, InputLeft, PgUp, Home:
		closed = pg.Turn(src, -1)
	default:
		closed = pg.Turn(src, 1)
	}
	return closed, !closed
}

// HelpPages cuts lines ({keys, text}; a heading has no text) into pages of
// at most rows lines, between sections where it can; a page does not start
// with a blank line. There is always a page.
func HelpPages(lines [][2]string, rows int) [][][2]string {
	pages := [][][2]string{}
	for len(lines) > 0 {
		for len(lines) > 0 && lines[0][0] == "" {
			lines = lines[1:]
		}
		n := min(rows, len(lines))
		if n < len(lines) { // break before the last section that starts on the page
			for k := n; k > 1; k-- {
				if lines[k][1] == "" {
					n = k
					break
				}
			}
		}
		if n > 0 {
			pages = append(pages, lines[:n])
		}
		lines = lines[n:]
	}
	if len(pages) == 0 {
		pages = append(pages, nil)
	}
	return pages
}
