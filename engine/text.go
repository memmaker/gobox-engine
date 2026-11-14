package engine

import (
	"gobox/game"
	"strings"
)

// The game's text (extra): the engine draws no text. What its routines
// wrote is kept as lines of typed text where it goes: the text pane's
// lines, the prompt and the view window's lines (opProtection's runes),
// published as Pane, Prompt and ViewText; the screens' own text is their
// View data (the status line, the roster, after-combat, the credits). Text
// leaves on events (clearPane, clearView, clearPrompt), never by rectangle.

// textCell is a character the game wrote: upcased (DisplayString's way)
// or not, in colour on bg.
type textCell struct {
	set, upcase bool
	c           byte
	colour, bg  game.Ink
}

// line is a line of text, its characters by column (the pane's from 1, as
// the original's cursor counts them).
type line []textCell

// put sets column x of l to c.
func (l *line) put(x int, c textCell) {
	if x < 0 || x >= textCols {
		return
	}
	for len(*l) <= x {
		*l = append(*l, textCell{})
	}
	(*l)[x] = c
}

// viewRows is the lines of text over the view window.
const viewRows = 0xF

// texts is the text the game wrote, by where it goes.
type texts struct {
	pane   [paneRows]line // the text pane's lines
	more   [][]PaneRun    // the pane's earlier lots when its text ran over (pageOn): gobox pages them
	prompt line           // the prompt
	view   [viewRows]line // the lines over the view window
}

// cell is ch in colour on black, upcased.
func cell(colour game.Ink, ch byte) textCell {
	return textCell{set: true, upcase: true, c: ch, colour: colour, bg: game.InkBlack}
}

// paneWrite = 0709:0626 displayString in the pane: s on pane line i from
// column x, each character upcased, no wrapping.
func (e *Engine) paneWrite(i, x int, colour game.Ink, s string) {
	if i < 0 || i >= paneRows {
		return
	}
	for k := 0; k < len(s) && k < 255; k++ {
		e.pg.tx.pane[i].put(x+k, cell(colour, s[k]))
	}
}

// cursorWrite = 0709:01DF display_char at the text cursor (DS:5E24 x,
// DS:5E25 y): the original's one cursor is in the pane (rows msgTop..) or
// over the view window (rows 1..15); the cursor moves on.
func (e *Engine) cursorWrite(colour game.Ink, ch byte) {
	switch r := e.msgRow; {
	case r >= msgTop && r <= paneBottom:
		e.pg.tx.pane[r-msgTop].put(e.msgCol, cell(colour, ch))
	case r >= 1 && r <= viewRows:
		e.pg.tx.view[r-1].put(e.msgCol, cell(colour, ch))
	}
	e.msgCol++
}

// clearPane blanks the text pane, its earlier lots with it.
func (e *Engine) clearPane() { e.pg.tx.pane, e.pg.tx.more = [paneRows]line{}, nil }

// clearPaneFrom blanks the pane's lines from i on (its earlier lots stay).
func (e *Engine) clearPaneFrom(i int) {
	for ; i < paneRows; i++ {
		e.pg.tx.pane[i] = nil
	}
}

// clearView takes the text off the view window (the picture replaced, the
// box cleared for text).
func (e *Engine) clearView() { e.pg.tx.view = [viewRows]line{} }

// clearWindow = ClearBox of the framed window: the view's text and the pane's.
func (e *Engine) clearWindow() { e.clearView(); e.clearPane() }

// clearPrompt = ClearRow24 (ovr26:06FB): blank the message/menu row.
func (e *Engine) clearPrompt() { e.pg.tx.prompt = nil }

// punct is the set of characters that end a word for paneFlow.
func punct(c byte) bool { return strings.IndexByte("!,-.:;?", c) >= 0 }

// paneFlow = 0709:0768 DisplayTextBlock in the pane from page row top
// (colour, clear, s): word wrapped, continuing at the text cursor when it
// lies in the pane from top. When the pane is full the lot is kept for
// gobox to page (the original's "Press any key to continue") and the pane
// starts over.
func (e *Engine) paneFlow(top int, colour game.Ink, clear bool, s string) {
	const left, right = 1, paneRight
	if e.msgCol < left || e.msgCol > right || e.msgRow < top || e.msgRow > paneBottom || clear {
		e.msgCol, e.msgRow = left, top
	}
	switch {
	case clear && top == msgTop:
		e.clearPane()
	case clear:
		e.clearPaneFrom(top - msgTop)
	}
	n := min(len(s), 255)
	if n == 0 {
		return
	}
	at := func(k int) byte { return s[k-1] } // 1-based
	i := 1
	printSeg := func(j int) {
		for ; i <= j; i++ {
			e.cursorWrite(colour, at(i))
		}
	}
	for {
		j := i
		for j < n && punct(at(j)) {
			j++
		}
		for j < n && !punct(at(j)) && at(j) != ' ' {
			j++
		}
		if at(j) != ' ' {
			for j+1 < n && punct(at(j+1)) {
				j++
			}
		}
		if j-i+e.msgCol > right {
			e.msgCol = left
			e.msgRow++
			for i < n && at(i) == ' ' {
				i++
			}
			if e.msgRow > paneBottom && i < n {
				e.msgCol, e.msgRow = left, top
				if top == msgTop {
					e.pageOn()
				} else {
					e.PromptRow24(game.InkLightMagenta, "Press any key to continue")
					e.clearPaneFrom(top - msgTop)
				}
				printSeg(j)
			}
		} else {
			printSeg(j)
		}
		if i > n {
			break
		}
	}
	if e.msgCol > right {
		e.msgCol = left
		e.msgRow++
	}
}

// pageOn ends a lot of the pane's text where the original asked for a key
// and cleared the pane: the lot is kept for gobox to page,
// the pane starts over.
func (e *Engine) pageOn() {
	more := append(e.pg.tx.more, e.pane().Lines...)
	e.clearPane()
	e.pg.tx.more = more
}

// paneText is DisplayTextBlock in the whole text pane.
func (e *Engine) paneText(colour game.Ink, clear bool, s string) {
	e.paneFlow(msgTop, colour, clear, s)
}

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
	for k := 0; k < len(s) && k < 255; k++ {
		e.pg.tx.prompt.put(k, cell(colour, s[k]))
	}
}
