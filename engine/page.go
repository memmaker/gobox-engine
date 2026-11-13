package engine

import (
	"image"
	"image/color"

	"gobox/asset"
)

// The page (extra): what the original drew on its 320x200 screen, kept as
// typed layers (pictures, symbols and fills in drawing order) under the text
// cells, so a presenter draws the screen from the View.

// layer is one drawing call: w by h pixels d at at (mask: the value left
// out, -1 none).
type layer struct {
	at     image.Rectangle
	d      []byte
	w      int
	mask   int
	opaque bool            // it hides everything under it
	clip   image.Rectangle // drawn only inside (empty: all of it)
	id     uint64
	src    *asset.Image // what d is a frame of (nil: the engine's own pixels)
	frame  int          // its frame
}

type pageState struct {
	layers []layer
	ids    uint64 // layers drawn
	tx     texts  // the text the game wrote (text.go)
}

// draw records l over the page: layers it hides go, and the text it
// covers.
func (e *Engine) draw(l layer) {
	pg := &e.pg
	r := l.at.Intersect(image.Rect(0, 0, textCols*8, textRows*8))
	if r.Empty() {
		return
	}
	if l.opaque {
		keep := pg.layers[:0]
		for _, o := range pg.layers {
			if !o.at.Intersect(r).Eq(o.at) {
				keep = append(keep, o)
			}
		}
		clear(pg.layers[len(keep):])
		pg.layers = keep
	}
	pg.ids++
	l.id = pg.ids
	pg.layers = append(pg.layers, l)
	e.unwrite(r)

}

// erase takes what is drawn inside r off the page, and its text. A
// picture reaching into the cell round r goes too: the original cleared
// boxes inside their frames, and the frame's cells covered the rest.
func (e *Engine) erase(r image.Rectangle) {
	pg := &e.pg
	keep := pg.layers[:0]
	for _, o := range pg.layers {
		if !o.at.In(r.Inset(-8)) {
			keep = append(keep, o)
		}
	}
	clear(pg.layers[len(keep):])
	pg.layers = keep
	e.unwrite(r)
}

// picture records a blit of w by h pixels d, frame frame of src, at x0,
// y0 (mask as blit's).
func (e *Engine) picture(x0, y0, w, h int, d []byte, mask int, src *asset.Image, frame int) {
	d = d[:w*h]
	opaque := true
	for _, p := range d {
		if int(p) == mask || p > 16 && p < asset.AmigaBase {
			opaque = false
			break
		}
	}
	e.draw(layer{at: image.Rect(x0, y0, x0+w, y0+h), d: d, w: w, mask: mask, opaque: opaque, src: src, frame: frame})
}

// clearPage starts an empty page (ClearScreen).
func (e *Engine) clearPage() {
	clear(e.pg.layers)
	e.pg.layers = e.pg.layers[:0]
	e.newPage()
	e.pg.tx = texts{}
}

// newPage ends what belongs to the page being replaced: the sheet's
// page (its portrait box), the icon's drop.
func (e *Engine) newPage() {
	e.snap.portrait, e.snap.iconDrop, e.snap.iconBox, e.snap.sheet, e.snap.results, e.snap.full = CellRect{}, nil, CellRect{}, nil, nil, false
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

// pagePictures are the page's layers as pixels.
func (e *Engine) pagePictures() []Picture {
	out := make([]Picture, 0, len(e.pg.layers))
	for i := range e.pg.layers {
		out = append(out, e.layerPicture(&e.pg.layers[i]))
	}
	return out
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

// layerPicture is l as the page shows it: its pixels (or its DOS twin's),
// the presenter's to colour (Picture.Pix).
func (e *Engine) layerPicture(l *layer) Picture {
	d, w := l.d, l.w
	at := l.at
	var pal []color.RGBA // the shown art's own colours (Image.Palette)
	if l.src != nil {
		pal = l.src.Palette
	}
	if a := e.art(l.src, l.frame); a != l.src && len(l.d) == len(l.src.Frame(l.frame)) {
		pal = a.Palette // a whole frame: its DOS twin, maybe bigger, at its own size, placed by its own cell offset (the 3D sprites')
		d, w = a.Frame(l.frame), a.Width*8
		if b := image.Rect(0, 0, w, len(d)/w); b.Size() != at.Size() {
			at = b.Add(at.Min.Add(image.Pt(a.X-l.src.X, a.Y-l.src.Y).Mul(8)))
		}
	}
	s := Picture{Pix: d, PixW: w, At: at, Palette: pal}
	if l.mask >= 0 && l.mask < 256 {
		m := uint8(l.mask)
		s.Mask = &m
	}
	if l.src != nil {
		s.File, s.Block, s.Frame = l.src.File, l.src.Block, l.src.Base+l.frame
	}
	if !l.clip.Empty() {
		s.At = at.Intersect(l.clip)
		s.Pix, s.PixW = crop(d, w, s.At.Sub(at.Min)), s.At.Dx()
	}
	return s
}

// crop is the pixels of r in d, w wide.
func crop(d []byte, w int, r image.Rectangle) []byte {
	out := make([]byte, 0, r.Dx()*r.Dy())
	for y := r.Min.Y; y < r.Max.Y; y++ {
		out = append(out, d[y*w+r.Min.X:y*w+r.Max.X]...)
	}
	return out
}

// adventure: the party is out in the world (not the party menu, the
// finale's pictures, or help, the journal, the maps over it).
func (e *Engine) adventure() bool {
	sheet := e.view.on || e.iconEd != nil // the sheet (and its pickers) and the icon editor (also opened from camp's menus) are pages of their own over the world
	return e.beginFlag() != 0 && e.beginFlag() != 7 && !sheet
}
