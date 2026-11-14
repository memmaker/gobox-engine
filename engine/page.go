package engine

import "gobox/asset"

// The page (extra): the text the original wrote on its screen, in cells
// (text.go); what the view window shows is a fact of its own (viewShown).

type pageState struct {
	tx texts // the text the game wrote (text.go)
}

// clearPage starts an empty page (ClearScreen).
func (e *Engine) clearPage() {
	e.newPage()
	e.pg.tx = texts{}
}

// newPage ends what belongs to the page being replaced: the sheet's
// page (its portrait box), the icon's drop.
func (e *Engine) newPage() {
	e.snap.portrait, e.snap.iconDrop, e.snap.iconBox, e.snap.sheet, e.snap.results, e.snap.full = CellRect{}, nil, CellRect{}, nil, nil, false
	e.snap.shown, e.es.bigPic = viewShown{}, nil
}

// viewShown is what the original last put in the view window: a picture
// (an animation's frame, the approaching sprite's, a still) or the
// encounter portrait (Engine.port); zero: nothing (the 3D view, the area
// map, text). The page replaced and the window showing anything else end it.
type viewShown struct {
	pic      *asset.Image
	portrait bool
}

// setWin makes w what the view window shows; a picture shown goes with
// any other.
func (e *Engine) setWin(w Window) {
	e.snap.win = w
	if w != WinPicture {
		e.snap.shown = viewShown{}
	}
}

// show3D makes the view window the first-person view (ovr30:05FD
// Draw3DView); the text over the picture frame goes.
func (e *Engine) show3D() {
	e.setWin(WinThreeD)
	e.unwrite(InnerWindow.Pixels())
}

// showPic puts im in the view window.
func (e *Engine) showPic(im *asset.Image) {
	e.setWin(WinPicture)
	e.snap.shown = viewShown{pic: im}
}

// redrawKind is a screen that draws itself up again at the page's size.
type redrawKind byte

const (
	rdSheet    redrawKind = iota // the character sheet
	rdCreation                   // the creation sheet
)

// pushRedraw makes k the screen resizePage draws, until popRedraw puts
// the one before back.
func (e *Engine) pushRedraw(k redrawKind) { e.ext.redraws = append(e.ext.redraws, k) }

func (e *Engine) popRedraw() {
	if n := len(e.ext.redraws); n > 0 {
		e.ext.redraws = e.ext.redraws[:n-1]
	}
}

// redrawNow draws the screen on top of the redraws again.
func (e *Engine) redrawNow() {
	if n := len(e.ext.redraws); n > 0 {
		switch e.ext.redraws[n-1] {
		case rdSheet:
			e.CharacterSheet()
		case rdCreation:
			if e.creation != nil {
				e.creationSheet(e.creation.c)
			}
		}
	}
}

// pagePane is the text pane out adventuring or when it holds text (nil otherwise).
func (e *Engine) pagePane() *Pane {
	if p := e.pane(); e.adventure() || p.Shown() {
		return p
	}
	return nil
}

// pageText is the text over the view window, rows from 1.
func (e *Engine) pageText() [][]PaneRun {
	var out [][]PaneRun
	for y, row := range e.pg.tx.view {
		if l := e.runs(row); len(l) > 0 {
			for len(out) <= y {
				out = append(out, nil)
			}
			out[y] = l
		}
	}
	return out
}

// pane is the text pane's text as runs.
func (e *Engine) pane() *Pane {
	p := &Pane{Lines: make([][]PaneRun, paneRows)}
	for i, row := range e.pg.tx.pane {
		if len(row) > TextPane.X1 {
			p.Lines[i] = e.runs(row[TextPane.X1:])
		}
	}
	return p
}

// runs is a row of text as runs of one colour, X from the row's start.
func (e *Engine) runs(row []textCell) []PaneRun {
	var l []PaneRun
	for x, c := range row {
		if !c.set {
			continue
		}
		fg, bg := c.colour, c.bg
		if n := len(l) - 1; n >= 0 && l[n].X+len(l[n].Text) == x && l[n].Fg == fg && l[n].Bg == bg && l[n].Upcase == c.upcase {
			l[n].Text += string(c.c)
			continue
		}
		l = append(l, PaneRun{X: x, Text: string(c.c), Fg: fg, Bg: bg, Upcase: c.upcase})
	}
	return l
}

// adventure: the party is out in the world (not the party menu, the
// finale's pictures, or help, the journal, the maps over it).
func (e *Engine) adventure() bool {
	sheet := e.view.on || e.iconEd != nil // the sheet (and its pickers) and the icon editor (also opened from camp's menus) are pages of their own over the world
	return e.beginFlag() != 0 && e.beginFlag() != 7 && !sheet
}
