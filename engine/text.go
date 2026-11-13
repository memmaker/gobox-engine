package engine

import (
	"gobox/game"
	"image"
)

// The game's text (extra): the engine draws no text. What its routines
// wrote is kept as typed text where it goes: the text pane's lines, the
// prompt row and the view window's text (opProtection's runes), published
// as Pane, Prompt and ViewText; the screens' own text is
// their View data (the status line, the roster, after-combat, the credits).

// textCell is a character the game wrote: upcased (DisplayString's way)
// or not, in colour on bg.
type textCell struct {
	set, upcase bool
	c           byte
	colour, bg  game.Ink
}

// texts is the text the game wrote, by where it goes.
type texts struct {
	pane   [paneRows][]textCell // the text pane's rows, from column 0 of the page
	more   [][]PaneRun          // the pane's earlier lots when its text ran over (pageOn): gobox pages them
	prompt []textCell           // the prompt row
	view   [0xF][]textCell      // the view window's rows 1..15, from column 0
}

// row is the text row of the page's row y (nil: text written there is not kept).
func (t *texts) row(x, y int) []textCell {
	switch {
	case y == promptRow:
		if len(t.prompt) != textCols {
			t.prompt = make([]textCell, textCols)
		}
		return t.prompt
	case y >= msgTop && y <= paneBottom && x <= paneRight:
		r := &t.pane[y-msgTop]
		if *r == nil {
			*r = make([]textCell, paneRight+1)
		}
		return *r
	case y >= ViewWindow.Y1 && y <= ViewWindow.Y2 && x <= ViewWindow.X2:
		r := &t.view[y-1]
		if *r == nil {
			*r = make([]textCell, ViewWindow.X2+1)
		}
		return *r
	}
	return nil
}

// put writes ch at cell (x, y).
func (e *Engine) put(x, y int, colour, bg game.Ink, ch byte, upcase bool) {
	if x < 0 || x > lastCol || y < 0 || y > promptRow {
		return
	}
	if r := e.pg.tx.row(x, y); x < len(r) {
		r[x] = textCell{set: true, upcase: upcase, c: ch, colour: colour, bg: bg}
	}
}

// unwrite takes the text off the page rectangle r (pixels).
func (e *Engine) unwrite(r image.Rectangle) {
	r = r.Intersect(image.Rect(0, 0, textCols*8, textRows*8))
	if r.Empty() {
		return
	}
	x1, y1, x2, y2 := r.Min.X/8, r.Min.Y/8, (r.Max.X-1)/8, (r.Max.Y-1)/8
	if x1 <= 1 && y1 <= msgTop && x2 >= paneRight && y2 >= paneBottom {
		e.pg.tx.more = nil // the pane cleared: its earlier lots go with it
	}
	for y := y1; y <= y2; y++ {
		row := e.pg.tx.row(x1, y)
		for x := x1; x <= x2 && x < len(row); x++ {
			row[x] = textCell{}
		}
	}
}

// char = 0709:01DF display_char(upcase, ch, count, colour, y, x):
// ch at cell (x,y) in colour on black. Glyph index = upcase(ch) mod 0x40
// into 8X8D1 block 201 (the presenter's: gobox's font).
func (e *Engine) char(x, y int, colour game.Ink, ch byte, upcase bool) {
	e.put(x, y, colour, game.InkBlack, ch, upcase)
}

// text = 0709:0626 displayString(x, y, colour, str): characters left to
// right, each upcased, no wrapping (x stops at lastCol).
func (e *Engine) text(x, y int, colour game.Ink, s string) {
	for i := 0; i < len(s) && i < 255; i++ {
		e.char(x+i, y, colour, s[i], true)
	}
}

// blank = ClearCells (03F1:0137), ClearBox (0709:0DD3), 0709:00B7: the
// cells [x1..x2] x [y1..y2] black, their text gone.
func (e *Engine) blank(x1, y1, x2, y2 int) {
	e.erase(image.Rect(x1*8, y1*8, x2*8+8, y2*8+8))
}

// textBlock = 0709:0768 DisplayTextBlock (x1, y1, x2, y2, colour, clear,
// s): word wrapped text in the cell box, continuing at the text cursor
// (DS:5E24 x, DS:5E25 y) when it lies inside the box. When the box is full
// the player is asked to press a key and the box starts over.
func (e *Engine) textBlock(x1, y1, x2, y2 int, colour game.Ink, clear bool, s string) {
	if x1 > lastCol || y1 > promptRow || x2 > lastCol || y2 > lastCol {
		return
	}
	if e.msgCol < x1 || e.msgCol > x2 || e.msgRow < y1 || e.msgRow > y2 {
		e.msgCol, e.msgRow = x1, y1
	}
	if clear {
		e.blank(x1, y1, x2, y2)
		e.msgCol, e.msgRow = x1, y1
	}
	Flow(s, x1, y1, x2, y2, &e.msgCol, &e.msgRow, func(x, y int, c byte) { e.char(x, y, colour, c, true) }, func() {
		if x1 == 1 && y1 == msgTop && x2 == paneRight && y2 == paneBottom { // the pane: gobox pages it (the original's "Press any key to continue")
			e.pageOn()
			return
		}
		e.PromptRow24(game.InkLightMagenta, "Press any key to continue")
		e.blank(x1, y1, x2, y2)
	})
}

// pageOn ends a lot of the pane's text where the original asked for a key
// and cleared the pane: the lot is kept for gobox to page,
// the pane starts over.
func (e *Engine) pageOn() {
	more := append(e.pg.tx.more, e.pane().Lines...)
	e.clearPane()
	e.pg.tx.more = more
}

// paneText is textBlock in the text pane.
func (e *Engine) paneText(colour game.Ink, clear bool, s string) {
	e.textBlock(1, msgTop, paneRight, paneBottom, colour, clear, s)
}

// clearPane blanks the text pane.
func (e *Engine) clearPane() { e.blank(1, msgTop, paneRight, paneBottom) }

// clearPrompt = ClearRow24 (ovr26:06FB): blank the message/menu row.
func (e *Engine) clearPrompt() { e.pg.tx.prompt = nil }

// PromptRow24 = 0709:0BF5: clear row 24, show, wait for a key: in a fight
// the fight's key pause, outside one keyWait (BUG-CLASSES 11), unless a
// question is already up (its key is the wait).
func (e *Engine) PromptRow24(colour game.Ink, s string) {
	e.prompt(colour, s)
	switch {
	case e.inFight(): // the presenter paces a fight: no wait here
	case e.eclWait == nil:
		e.keyWait(cont{kind: kNop})
	}
}

// prompt shows s alone on the prompt row.
func (e *Engine) prompt(colour game.Ink, s string) {
	e.clearPrompt()
	e.text(0, promptRow, colour, s)
}
