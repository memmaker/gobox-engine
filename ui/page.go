package ui

import (
	"bytes"
	"fmt"
	"gobox/asset"
	"gobox/game"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strconv"
	"strings"

	xdraw "golang.org/x/image/draw"

	"gobox/gpu"
)

// Rect is where the page lands on the window, in device pixels.
type Rect struct{ X, Y, W, H float64 }

// Note is a toast shown on the bottom line of the page (row 24, in place
// of the prompt), in the game's current font.
type Note struct {
	Text  string
	Show  bool
	Glyph func(byte) [8]byte
	Font  func(ch byte, upcase bool) [8]byte // a pane run's glyphs (front.PaneRun)
}

// page is the engine's screen (pageSize: 320x200 at least) drawn from the
// View's Page and Text.
type page struct {
	img    *gpu.Image
	px     *image.RGBA
	seq    uint64
	noted  bool
	own    string // the presenter's screens it shows (Screens)
	edit   string // the icon editor's content it shows (iconKey)
	drawn  bool
	page   []Picture            // what img shows: the View's pictures...
	text   string               // ...and text (pageText)
	overs  []overlay            // the screens' portraits over the page
	scans  map[image.Image]scan // the overlays on the GPU
	fit    *gpu.Image           // fit's whole-scale copy
	pane   *gpu.Image           // the text pane (paneImage)
	pic    *gpu.Image           // the picture window (pictureBox)
	toast  *gpu.Image           // the note's line over the 3D view
	credit *gpu.Image           // the remaster credit over a credits picture
	notes  *gpu.Image           // the text box's "> Notes" (paneBox)
}

// pageSize is v's page in pixels: its text cells, 8 by 8 (the original's
// 320x200 for a View without any).
func (p *Presenter) pageSize() (w, h int) {
	if p.cells().X == 0 || p.cells().Y == 0 {
		return asset.ScreenW, asset.ScreenH
	}
	return p.cells().X * 8, p.cells().Y * 8
}

// grid is the text cells the window holds: 40x25 at least, more at the
// largest whole scale the original's screen fits (scale), so a larger
// window shows more rather than bigger cells; the engine lays its screens
// out to it (ActResize).
func grid(dst *gpu.Image) (cols, rows int) {
	g := gridOf(dst.Bounds())
	return g.X, g.Y
}

// gridOf is grid for a window of bounds b.
func gridOf(b image.Rectangle) image.Point {
	k := int(Scale(b))
	return image.Pt(max(40, b.Dx()/(8*k)), max(25, b.Dy()/(8*k)))
}

// cells is the page's size in text cells: the credits fill the window at K
// (grid), whatever page the game keeps, never the 40x25 page scaled up;
// every other page is the original's (none given: 40x25).
func (p *Presenter) cells() image.Point {
	if !p.flags().Credits {
		return image.Point{}
	}
	return gridOf(p.screen)
}

// fit draws img, w by h unscaled pixels, over the whole of dst: nearest
// neighbour at the largest whole scale that fits (sharp pixels), then
// stretched the rest of the way to the window's edges, under a cell on
// each axis when the page is the window's grid. No black bars, ever:
// every screen fills the window. It returns the scale on each axis.
func (p *Presenter) fit(dst, img *gpu.Image, w, h int) (kx, ky float64) {
	b := dst.Bounds()
	k := max(1, min(b.Dx()/w, b.Dy()/h))
	if drawnAt != nil {
		drawnAt(drawnPage, dst, float64(k))
	}
	kx, ky = float64(b.Dx())/float64(w), float64(b.Dy())/float64(h)
	op := &gpu.DrawImageOptions{}
	if kx == float64(k) && ky == float64(k) {
		op.GeoM.Scale(kx, ky)
		dst.DrawImage(img, op)
		return
	}
	pg := &p.pg
	gpu.Sized(&pg.fit, w*k, h*k)
	op.GeoM.Scale(float64(k), float64(k))
	pg.fit.DrawImage(img, op)
	op = &gpu.DrawImageOptions{} // nearest: the pixels stay sharp
	op.GeoM.Scale(kx/float64(k), ky/float64(k))
	dst.DrawImage(pg.fit, op)
	return
}

// scan is an overlay on the GPU: the picture itself (w, h 0), or a copy
// scaled up to w by h window pixels (Catmull-Rom: sharper than the GPU's
// bilinear).
type scan struct {
	im   *gpu.Image
	w, h int
	part image.Rectangle // the part of the scan a scaled copy (w, h) is of
}

// changes is the change count the page was drawn at (0: no game set).
func (p *Presenter) changes() uint64 {
	if u := p.ui(); u != nil {
		return u.seq
	}
	return 0
}

// pageImage draws v's page (only when a new View or the note changed) and
// returns it.
func (p *Presenter) pageImage(note Note) *gpu.Image {
	pg := &p.pg
	w, h := p.pageSize()
	if gpu.Sized(&pg.img, w, h) {
		pg.px, pg.drawn = image.NewRGBA(image.Rect(0, 0, w, h)), false
	}
	own := p.pageKey()
	var sh sfShown
	sf := p.sfOf()
	if sf != nil && !p.modal() {
		sh = p.sfState(sf)
		own += fmt.Sprintf("%v%+v%+v", *sf, sh, p.gameScreen().ShipSheet)
	}
	if pg.drawn && !note.Show && !pg.noted && pg.own == own && pg.edit == iconKey(p.iconEditor(p.ink())) && (p.changes() == pg.seq || !p.modal() && !p.ownPage() && p.samePage(pg)) { // the View changed, the page not: no redraw, no upload
		pg.seq = p.changes()
		return pg.img
	}
	pg.drawn, pg.seq, pg.noted, pg.own, pg.edit = true, p.changes(), note.Show, own, iconKey(p.iconEditor(p.ink()))
	pg.page, pg.text = append(pg.page[:0], p.pictures()...), p.pageText()
	m := pg.px
	if p.modal() { // help, the journal, the maps, the overview: from their data alone, nothing of the page under them
		draw.Draw(m, m.Rect, image.Black, image.Point{}, draw.Src)
		pg.overs = nil
	} else {
		c := p.screenPage()
		if sf != nil {
			c.spaceFight(sf, sh, p.askUp(), p.gameScreen())
		}
		c.draw(m, p.windowBorder(), p.cells(), note)
		pg.overs = c.overs
	}
	p.drawModal(m, note)
	if note.Show {
		noteLine(m, note, h-8)
	}
	pg.img.WritePixels(m.Pix)
	return pg.img
}

// iconKey names the icon editor's content (the game need not bump the change count
// for an edit): a change redraws the page.
func iconKey(e *FrontIconEditor) string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%p %p %v %v %v %v %v %v", e.Poses[0], e.Poses[1], e.Large, e.Part, e.Shade, e.Colour, e.Custom, e.Colours)
}

// ownPage: v shows a screen the presenter lays out from its data
// (screenPage), drawn anew with each View.
func (p *Presenter) ownPage() bool {
	return p.sheetPage() != nil || p.picker(p.ink()) != nil || p.slotsView() != nil || p.iconEditor(p.ink()) != nil
}

// modal: v shows a screen over the whole page (help, the journal, the
// full map, the overview), drawn from its data over the page under it.
func (p *Presenter) modal() bool {
	return p.pager != nil || p.fullMap != nil || p.overview
}

// pagerMark is the column of pager pg's mark for the page before (d -1)
// or after (d +1) on a page cols cells wide, on its hint row; ok: there is
// such a page.
func pagerMark(pg *Pages, cols, d int) (x int, ok bool) {
	if d < 0 {
		return 1, pg.Page > 0
	}
	return cols - 2, pg.Page < pg.Pages-1
}

// drawModal draws v's screen over the whole page (modal) into m as the
// game draws such a page: a box round it, text in the game's font.
func (p *Presenter) drawModal(m *image.RGBA, note Note) {
	if !p.modal() {
		return
	}
	cols, rows := m.Rect.Dx()/8, m.Rect.Dy()/8
	run := func(r PaneRun, y int) {
		for i := 0; i < len(r.Text); i++ {
			glyphAt(m, note.Font(r.Text[i], r.Upcase), (r.X+i)*8, y*8, r.Fg, r.Bg)
		}
	}
	str := func(x, y, c int, s string) { // DisplayString
		run(PaneRun{X: x, Text: s[:max(0, min(len(s), cols-x))], Fg: col(c), Bg: col(inkBlack), Upcase: true}, y)
	}
	fill := func(r image.Rectangle, c color.RGBA) { draw.Draw(m, r, image.NewUniform(c), image.Point{}, draw.Src) }
	box := func(x2, y2 int) { // DrawBoxTitled(1, 1, x2, y2): the inside cleared, the border's cells round it
		fill(image.Rect(8, 8, x2*8+8, y2*8+8), col(inkBlack))
		cellBorder(m, p.windowBorder(), p.cells(), CellRect{X1: 1, Y1: 1, X2: x2, Y2: y2})
	}
	prompt := func(s string, c int) { // the prompt row cleared, s on it
		str(0, rows-1, inkBlack, strings.Repeat(" ", cols))
		str(0, rows-1, c, s)
	}
	switch {
	case p.pager != nil:
		pg, _ := p.pages()
		box(cols-2, rows-2)
		run(pg.Title, 1)
		for _, s := range pg.Scans { // the paper; the scan goes over the window (drawOverlays)
			fill(s.At, s.Paper)
		}
		for i, l := range pg.Lines {
			for _, r := range l {
				run(r, pg.Top+i)
			}
		}
		run(pg.Hint, rows-2)
		for _, d := range []int{-1, 1} { // the clickable marks (pagerMark): ◀ while there is a page before, ▶ while one after
			if x, ok := pagerMark(pg, cols, d); ok {
				for i := range 4 { // a triangle 4 pixels deep, 7 tall, pointing d
					c := x*8 + 2 + i
					if d < 0 {
						c = x*8 + 5 - i
					}
					fill(image.Rect(c, (rows-2)*8+1+i, c+1, (rows-2)*8+8-i), col(inkGreen))
				}
			}
		}
	case p.fm() != nil:
		fm := p.fm()
		lx := fm.LegendAt.X
		fill(m.Rect, col(inkBlack))
		str(0, 0, inkWhite, fm.Name[:min(len(fm.Name), lx-1)])
		if len(fm.Legend) > 0 { // no labels: no header over nothing
			str(lx, fm.LegendAt.Y-1, inkWhite, "Labels") // a row below the window's top edge
		}
		for i, l := range fm.Legend {
			str(lx, fm.LegendAt.Y+i, labelInks[l.Kind], string(l.Letter))
			str(lx+2, fm.LegendAt.Y+i, inkGreen, l.Text[:min(len(l.Text), cols-lx-2)])
		}
		str(1, rows-2, inkCyan, fm.Info[:min(len(fm.Info), cols-2)])
		mi := mapImage(fm.Map, fm.Cell, p.ink(), note.Font)
		draw.Draw(m, mi.Bounds().Add(fm.At), mi, image.Point{}, draw.Over)
		n, c := fm.Cell, fm.At.Add(fm.Cursor.Mul(fm.Cell))
		for _, r := range []image.Rectangle{image.Rect(0, 0, n, 1), image.Rect(0, n-1, n, n), image.Rect(0, 0, 1, n), image.Rect(n-1, 0, n, n)} {
			fill(r.Add(c), col(inkYellow))
		}
		if fm.Label != nil {
			prompt("Label: ", 13)
			str(7, rows-1, inkWhite, *fm.Label+"_")
		}
	case p.overview:
		box(cols-2, rows-2)
		const title = "Party"
		str((cols-1-len(title))/2, 1, inkWhite, title)
		y := 3
		// Spells and XP to the next level only where the window has room for them
		str(2, y, inkWhite, "Name")
		str(19, y, inkWhite, "HP")
		str(27, y, inkWhite, "AC")
		str(31, y, inkWhite, "Status")
		if cols >= 50 {
			str(43, y, inkWhite, "Spells")
		}
		if cols >= 61 {
			str(50, y, inkWhite, "XP to next")
		}
		y += 2
		for i, c := range p.party() {
			if y > rows-9 { // above the purse
				break
			}
			colour := inkCyan
			if i == p.selectedIndex() {
				colour = inkWhite
			}
			str(2, y, colour, c.Name[:min(len(c.Name), 15)])
			hpc := inkGreen
			switch {
			case c.HP <= 0:
				hpc = inkRed
			case c.HP*2 < c.MaxHP:
				hpc = inkYellow
			}
			str(19, y, hpc, strconv.Itoa(c.HP)+"/"+strconv.Itoa(c.MaxHP))
			str(27, y, inkGreen, strconv.Itoa(c.AC))
			str(31, y, inkGreen, c.Status)
			if cols >= 50 {
				str(43, y, inkGreen, c.SpellsReady)
			}
			if cols >= 61 {
				xp := "-"
				if c.NextExp > 0 {
					xp = strconv.Itoa(max(c.NextExp-expOf(c), 0))
				}
				str(50, y, inkGreen, xp)
			}
			w := 0 // the hit point bar under the name, 32 steps
			if c.MaxHP > 0 && c.HP > 0 {
				w = min((c.HP*0x20+c.MaxHP-1)/c.MaxHP, 0x20)
			}
			px, py := 2*8, (y+1)*8+2
			fill(image.Rect(px, py, px+0x20*4, py+3), col(inkDark))
			fill(image.Rect(px, py, px+w*4, py+3), col(hpc))
			y += 2
		}
		if c := game.CommonOf(p.state()); c.Purse != nil { // a blank row between the pairs and before the items
			str(2, rows-7, inkYellow, fmt.Sprintf("%d gold  %d copper", c.Purse[Coin], c.Purse[game.Copper]))
			str(2, rows-5, inkYellow, fmt.Sprintf("%d gems  %d jewelry", c.Purse[game.Gems], c.Purse[game.Jewelry]))
			str(2, rows-3, inkCyan, fmt.Sprintf("Party items %d/%d", c.PoolItems, c.PartyCap))
		}
		// no prompt: the frame runs round the whole page, any key still closes it
	}
}

// drawPictures draws pics on black into m, whose top left is page pixel
// at: a fill as its colour, a picture over what is under it (the borders
// are the page's: cellPage).
func drawPictures(m *image.RGBA, pics []Picture, at image.Point) {
	draw.Draw(m, m.Bounds(), image.Black, image.Point{}, draw.Src)
	for _, s := range pics {
		r := s.At.Sub(at).Add(m.Rect.Min)
		if s.At.Empty() { // no place given (a game's Common.Picture): the whole area
			r = m.Rect
		}
		switch {
		case s.Fill:
			draw.Draw(m, r, image.NewUniform(s.Colour), image.Point{}, draw.Src)
		case s.Img != nil:
			if b := s.Img.Bounds(); b.Dx() == r.Dx() && b.Dy() == r.Dy() || r.Empty() || b.Empty() {
				draw.Draw(m, r, s.Img, b.Min, draw.Over)
			} else { // not the area's size (a picker's cell): scaled to fit, aspect kept, centred
				w, h := r.Dx(), b.Dy()*r.Dx()/b.Dx()
				if h > r.Dy() {
					w, h = b.Dx()*r.Dy()/b.Dy(), r.Dy()
				}
				xdraw.NearestNeighbor.Scale(m, image.Rect(0, 0, w, h).Add(r.Min.Add(image.Pt((r.Dx()-w)/2, (r.Dy()-h)/2))), s.Img, b, draw.Over, nil)
			}
		}
	}
}

// pictureBox draws the pictures in page rectangle r (the picture window)
// into *img.
func pictureBox(img **gpu.Image, pics []Picture, r image.Rectangle) *gpu.Image {
	m := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	drawPictures(m, pics, r.Min)
	return upload(img, m)
}

// noteLine is the one toast: the note in red (EGA 12), centred on a black
// line at y of m, the bottom row of every view (the page's prompt line, the
// 3D view's band, the inventory's row under its menu row, combat's status line).
func noteLine(m *image.RGBA, note Note, y int) {
	w := m.Rect.Dx()
	draw.Draw(m, image.Rect(0, y, w, y+8), image.Black, image.Point{}, draw.Src)
	x0 := (w - 8*len(note.Text)) / 2
	for i := 0; i < len(note.Text); i++ {
		glyphAt(m, note.Glyph(note.Text[i]), x0+i*8, y, col(inkRed), color.RGBA{A: 255})
	}
}

// paneImage draws the text pane's runs on black in *img (made or remade
// at the pane's size): what drawBox shows inside the pane's border.
func paneImage(img **gpu.Image, pn *Pane, note Note) *gpu.Image {
	r := paneCells(pn).Pixels()
	m := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	draw.Draw(m, m.Rect, image.Black, image.Point{}, draw.Src)
	for y, l := range pn.Lines {
		for _, run := range l {
			for i := 0; i < len(run.Text); i++ {
				glyphAt(m, note.Font(run.Text[i], run.Upcase), (run.X+i)*8, y*8, run.Fg, run.Bg)
			}
		}
	}
	for i := 3; i < len(m.Pix); i += 4 { // opaque: a run's transparent background must not show the 3D view through
		m.Pix[i] = 255
	}
	return upload(img, m)
}

// viewTextImage draws the text over the view window (ViewText, rows
// rows from 1, each line from the inside's left edge) as the window's
// inside r shows it.
func viewTextImage(img **gpu.Image, rows [][]PaneRun, r image.Rectangle, note Note) *gpu.Image {
	m := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	draw.Draw(m, m.Rect, image.Black, image.Point{}, draw.Src)
	for y, l := range rows {
		for _, run := range l {
			for i := 0; i < len(run.Text); i++ {
				glyphAt(m, note.Font(run.Text[i], run.Upcase), (run.X+i)*8, (y+1)*8-r.Min.Y, run.Fg, run.Bg)
			}
		}
	}
	return upload(img, m)
}

// uploaded is the pixels each upload target holds: the same pixels again are
// not written (a WritePixels submits the frame so far and waits for the GPU).
var uploaded = map[*gpu.Image][]byte{}

// upload writes m to *img, made or remade at m's size, unless it holds m already.
func upload(img **gpu.Image, m *image.RGBA) *gpu.Image {
	old := *img
	if gpu.Sized(img, m.Rect.Dx(), m.Rect.Dy()) {
		delete(uploaded, old)
	}
	if px, ok := uploaded[*img]; ok && bytes.Equal(px, m.Pix) {
		return *img
	}
	(*img).WritePixels(m.Pix)
	uploaded[*img] = m.Pix // m is the caller's new image every time: kept, not copied
	return *img
}

// glyphAt draws glyph g at x, y: set bits fg, the rest bg.
func glyphAt(m *image.RGBA, g [8]byte, x, y int, fg, bg color.RGBA) {
	fg, bg = fontInk(fg), fontInk(bg)
	asset.DrawGlyph(m, g, x, y, &fg, &bg)
}

// fontInk is c as gobox's font colour: the EGA colour nearest c in gobox's
// own RGB, whatever palette the game drew c from (no game sets font colours).
func fontInk(c color.RGBA) color.RGBA {
	best, d := 0, 1<<30
	for i, k := range egaInk {
		dr, dg, db := int(k.R)-int(c.R), int(k.G)-int(c.G), int(k.B)-int(c.B)
		if e := dr*dr + dg*dg + db*db; e < d {
			best, d = i, e
		}
	}
	k := egaInk[best]
	k.A = c.A
	return k
}

// drawPage shows the whole page over the whole window (fit), with the
// overlays at the window's own resolution; it returns where. A page not yet
// of the window's grid (the engine has the resize still to do) is shown
// at a whole scale in the middle meanwhile, so nothing stretches.
func (p *Presenter) drawPage(dst *gpu.Image, note Note) Rect {
	img := p.pageImage(note)
	dst.Fill(color.Black)
	b := dst.Bounds()
	w, h := p.pageSize()
	var r Rect
	if cols, rows := grid(dst); cols*8 == w && rows*8 == h {
		kx, ky := p.fit(dst, img, w, h)
		r = Rect{0, 0, kx * float64(w), ky * float64(h)}
	} else {
		k := math.Max(1, math.Floor(math.Min(float64(b.Dx())/float64(w), float64(b.Dy())/float64(h))))
		if drawnAt != nil {
			drawnAt(drawnPage, dst, k)
		}
		r = Rect{(float64(b.Dx()) - k*float64(w)) / 2, (float64(b.Dy()) - k*float64(h)) / 2, k * float64(w), k * float64(h)}
		op := &gpu.DrawImageOptions{}
		op.GeoM.Scale(k, k)
		op.GeoM.Translate(r.X, r.Y)
		dst.DrawImage(img, op)
	}
	os := append([]overlay(nil), p.pg.overs...)
	if pg, _ := p.pages(); pg != nil {
		for _, s := range pg.Scans {
			os = append(os, overlay{Img: s.Img, Src: s.Src, At: s.At})
		}
	}
	p.drawOverlays(dst, os, r, w, h)
	return r
}

// drawOverlays draws the overlays (journal scans) over the page of w by h
// pixels at r.
func (p *Presenter) drawOverlays(dst *gpu.Image, os []overlay, r Rect, w, h int) {
	pg := &p.pg
	for img, im := range pg.scans { // drop the copies of scans no longer shown
		keep := false
		for _, o := range os {
			keep = keep || o.Img == img
		}
		if !keep {
			im.im.Deallocate()
			delete(pg.scans, img)
		}
	}
	kx, ky := r.W/float64(w), r.H/float64(h)
	for _, o := range os {
		b := o.Part()
		tw, th := int(float64(o.At.Dx())*kx), int(float64(o.At.Dy())*ky) // its place in window pixels
		x, y := r.X+float64(o.At.Min.X)*kx, r.Y+float64(o.At.Min.Y)*ky
		// a picture smaller than its place (a portrait) is scaled up by whole
		// pixels when that nearly fills the place (crisp pixel art, on its
		// foot), else by Catmull-Rom to the exact size; a larger one (a journal
		// scan) shrinks bilinear on the GPU
		s := wholeScale(b.Dx(), b.Dy(), tw, th)
		sw, sh := 0, 0
		if s == 0 && tw >= b.Dx() && th >= b.Dy() && !o.Crisp {
			sw, sh = tw, th
		}
		sc, ok := pg.scans[o.Img]
		if !ok || sc.w != sw || sc.h != sh || sw > 0 && sc.part != b { // a scaled copy is of one part: a scrolled scan copies its new one
			if ok {
				sc.im.Deallocate()
			}
			if pg.scans == nil {
				pg.scans = map[image.Image]scan{}
			}
			src := o.Img
			if sw > 0 {
				m := image.NewRGBA(image.Rect(0, 0, sw, sh))
				xdraw.CatmullRom.Scale(m, m.Bounds(), o.Img, b, draw.Src, nil)
				src = m
			}
			sc = scan{gpu.NewImageFromImage(src), sw, sh, b}
			pg.scans[o.Img] = sc
		}
		im := sc.im
		op := &gpu.DrawImageOptions{Filter: gpu.FilterLinear}
		switch {
		case o.Crisp: // sharp pixels at any scale, the aspect kept
			f := min(float64(tw)/float64(b.Dx()), float64(th)/float64(b.Dy()))
			im = im.SubImage(b).(*gpu.Image)
			op.Filter = gpu.FilterNearest
			op.GeoM.Scale(f, f)
			x += (float64(tw) - f*float64(b.Dx())) / 2
			y += (float64(th) - f*float64(b.Dy())) / 2
		case s > 0:
			im = im.SubImage(b).(*gpu.Image)
			op.Filter = gpu.FilterNearest
			op.GeoM.Scale(float64(s), float64(s))
			x += float64(tw-s*b.Dx()) / 2
			y += float64(th - s*b.Dy())
		case sw > 0:
			op.Filter = gpu.FilterNearest
		default:
			im = im.SubImage(b).(*gpu.Image) // drawn from its own top left
			op.GeoM.Scale(float64(tw)/float64(b.Dx()), float64(th)/float64(b.Dy()))
		}
		op.GeoM.Translate(x, y)
		dst.DrawImage(im, op)
	}
}

// samePage reports whether v's page and text are what pg.img shows. The
// engine publishes a View every tick, most of them with the page unchanged;
// a layer's picture is the same when it is the same pixels.
func (p *Presenter) samePage(pg *page) bool {
	if len(p.pictures()) != len(pg.page) || p.pageText() != pg.text {
		return false
	}
	for i, s := range p.pictures() {
		if o := pg.page[i]; s.At != o.At || s.Fill != o.Fill || s.Colour != o.Colour || !sameImage(s.Img, o.Img) {
			return false
		}
	}
	return true
}

// pageText is the text v's page shows (screenPage's texts), to tell pages
// apart.
func (p *Presenter) pageText() string {
	var b strings.Builder
	if p.pane() != nil {
		fmt.Fprint(&b, p.pane().Lines)
	}
	fmt.Fprint(&b, p.prompt(), p.viewText(), p.flags().Credits, p.ink())
	if p.results() != nil {
		fmt.Fprintf(&b, "%+v", *p.results())
	}
	return b.String()
}

func sameImage(a, b image.Image) bool {
	switch a := a.(type) {
	case *image.RGBA:
		b, ok := b.(*image.RGBA)
		return ok && a.Rect == b.Rect && a.Stride == b.Stride && len(a.Pix) == len(b.Pix) && (len(a.Pix) == 0 || &a.Pix[0] == &b.Pix[0])
	case nil:
		return b == nil
	}
	return false
}

// paneBox draws the text pane in its box at scale k round (cx, cy), and on
// its bottom border, right, the "> Notes" that puts the text into the
// journal (ActNote; Ctrl-N, the pad's Back).
func (p *Presenter) paneBox(dst *gpu.Image, note Note, k float64, cx, cy int) {
	txt := paneCells(p.pane()).Framed()
	drawBox(dst, paneImage(&p.pg.pane, p.pane(), note), p.border.val.tiles, txt, k, cx, cy)
	const label = " > Notes "
	m := image.NewRGBA(image.Rect(0, 0, 8*len(label), 8))
	draw.Draw(m, m.Rect, image.Black, image.Point{}, draw.Src)
	DrawText(m, note.Glyph, label, 0, 0, col(inkYellow))
	px := txt.Pixels()
	hw, hh := int(k*float64(px.Dx())/2), int(k*float64(px.Dy())/2)
	p.paneRect = image.Rect(cx-hw, cy-hh, cx+hw, cy+hh) // card menus keep clear of it
	x, y := float64(cx)+k*float64(px.Dx()/2-16-m.Rect.Dx()), float64(cy)+k*float64(px.Dy()/2-8)
	op := &gpu.DrawImageOptions{}
	op.GeoM.Scale(k, k)
	op.GeoM.Translate(x, y)
	dst.DrawImage(upload(&p.pg.notes, m), op)
	p.noteBtn = image.Rect(int(x), int(y), int(x+k*float64(m.Rect.Dx())), int(y+k*8))
}
