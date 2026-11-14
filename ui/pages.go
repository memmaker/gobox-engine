package ui

import (
	"fmt"
	"gobox/adventure"
	"gobox/game"
	"image"
	"image/color"
	"image/draw"
	"slices"
	"strconv"
)

// The screens the engine publishes as data (the character sheet, the
// creation's, the pickers, the save slots, the icon editor) are laid out
// here on the page's text cells as the original's code put them there.

// The text colours by role (fontPalette, applyTheme).
var inkBlack, inkBlue, inkGrey, inkDark, inkGreen, inkCyan, inkRed, inkMagenta, inkYellow, inkWhite int

// overlay is a picture drawn over the page at the window's resolution:
// the part Src of Img (empty: all of it) scaled into the page rectangle At
// (a portrait, a journal scan).
type overlay struct {
	Img     image.Image
	Src, At image.Rectangle
	Crisp   bool // pixel art: nearest neighbour, its aspect kept, centred in At
}

// Part is the rectangle of Img the overlay shows.
func (o overlay) Part() image.Rectangle {
	if o.Src.Empty() {
		return o.Img.Bounds()
	}
	return o.Src
}

type cell struct {
	set    bool
	ch     byte
	fg, bg color.RGBA
}

// cellPage is the page as the presenter draws it: the View's pictures and
// text, and the screens' text, borders and portraits over them.
type cellPage struct {
	cols, rows int
	ink        [16]color.RGBA
	cells      [][]cell
	pics       []Picture  // the View's pictures, then the cells cleared (fills)
	boxes      []CellRect // DrawBorder's insides: the border's cells round them (the window border)
	overs      []overlay
}

// newCellPage is a blank page cells in size (40x25 at least) in ink, under pics.
func newCellPage(cells image.Point, ink [16]color.RGBA, pics []Picture) *cellPage {
	c := &cellPage{cols: max(cells.X, 40), rows: max(cells.Y, 25), ink: ink}
	c.cells = make([][]cell, c.rows)
	for y := range c.cells {
		c.cells[y] = make([]cell, c.cols)
	}
	c.pics = append(c.pics, pics...)
	return c
}

// run puts r's text on row y.
func (c *cellPage) run(r PaneRun, y int) {
	if y < 0 || y >= c.rows {
		return
	}
	for i := 0; i < len(r.Text) && r.X+i < c.cols; i++ {
		if r.X+i >= 0 {
			c.cells[y][r.X+i] = cell{true, r.Text[i], fontInk(r.Fg), fontInk(r.Bg)}
		}
	}
}

// str = DisplayString: s from cell (x, y) in colour on black, upcased.
func (c *cellPage) str(x, y, colour int, s string) {
	if x > c.cols-1 || y > c.cols-1 {
		return
	}
	for i := 0; i < len(s) && x+i < c.cols && y < c.rows; i++ {
		c.cells[y][x+i] = cell{true, s[i], egaInk[colour], c.ink[inkBlack]} // text in gobox's colours, not the page palette
	}
}

// clear = ClearCells: the cells black.
func (c *cellPage) clear(x1, y1, x2, y2 int) {
	r := image.Rect(x1*8, y1*8, x2*8+8, y2*8+8)
	c.pics = append(c.pics, Picture{At: r, Fill: true, Colour: c.ink[inkBlack]})
	c.unset(r)
}

func (c *cellPage) unset(r image.Rectangle) {
	r = r.Intersect(image.Rect(0, 0, c.cols*8, c.rows*8))
	for y := r.Min.Y / 8; y < (r.Max.Y+7)/8; y++ {
		for x := r.Min.X / 8; x < (r.Max.X+7)/8; x++ {
			c.cells[y][x].set = false
		}
	}
}

// border = DrawBorder: the border's cells round the cells x1..x2, y1..y2.
func (c *cellPage) border(x1, y1, x2, y2 int) {
	r := CellRect{X1: x1, Y1: y1, X2: x2, Y2: y2}
	c.boxes = append(c.boxes, r)
	f := r.Framed().Pixels()
	for _, e := range []image.Rectangle{{f.Min, image.Pt(f.Max.X, f.Min.Y+8)}, {image.Pt(f.Min.X, f.Max.Y-8), f.Max}, {f.Min, image.Pt(f.Min.X+8, f.Max.Y)}, {image.Pt(f.Max.X-8, f.Min.Y), f.Max}} {
		c.unset(e)
	}
}

// box = DrawBoxTitled: cleared, bordered, the title centred on row y1.
func (c *cellPage) box(x1, y1, x2, y2, colour int, title string) {
	if len(title) > 0x28 {
		title = title[:0x28]
	}
	c.clear(x1, y1, x2, y2)
	c.border(x1, y1, x2, y2)
	if title != "" {
		c.str((x1+x2-len(title))/2, y1, colour, title)
	}
}

// screenPage is v's page as the presenter draws it: the page's pictures
// and boxes and the screens' text over them.
// contentPictures are the original art a screen of its own shows from the
// engine's page: the picker's pictures in their cells, the portrait and
// combat icon on the sheets; never a frame or a fill (the credits sit on
// black, not over the title picture).
func (p *Presenter) contentPictures() []Picture {
	var in []image.Rectangle
	pk, ed := p.picker(p.ink()), p.iconEditor(p.ink())
	switch {
	case p.slotsView() != nil, ed != nil && pk == nil:
	case pk != nil:
		return pk.Pics
	default:
		var boxes []CellRect
		if cr := p.creation(); cr != nil {
			boxes = append(boxes, cr.Head, cr.Body, cr.Icon)
		}
		for _, b := range boxes {
			if b != (CellRect{}) { // a box the screen has: not one at the origin
				in = append(in, b.Pixels())
			}
		}
	}
	var out []Picture
	for _, pc := range p.pictures() {
		if pc.Img == nil && pc.Pix == nil || pc.File == "" && pc.Name == "" { // the pool's (withPoolPortrait, withPoolIcon) have a Name, no File
			continue
		}
		if slices.ContainsFunc(in, func(r image.Rectangle) bool { return !r.Empty() && pc.At.Overlaps(r) }) {
			out = append(out, pc)
		}
	}
	return out
}

func (p *Presenter) screenPage() *cellPage {
	sl := p.slotsView()
	cr, pk, ed := p.creation(), p.picker(p.ink()), p.iconEditor(p.ink())
	var sp *game.SheetPage
	var portrait image.Image
	if sl == nil {
		sp = p.sheetPage()
		portrait = p.sheetPortrait(sp)
	} else {
		pk, ed = nil, nil
	}
	pics, ink := p.contentPictures(), p.ink() // nothing of the engine's page: only the art in the screen's own places
	if ed != nil || pk != nil {               // also a picker over the icon editor, in its palette: no grey cells
		ink[inkBlack] = color.RGBA{A: 255} // the combat palette shows black dark grey (for the icon's black): the page black as every screen
	}
	if cr != nil && pk == nil && sl == nil { // the combat icon on grey, as in the icon editor: its black parts show
		pics = append([]Picture{{At: cr.Icon.Pixels(), Fill: true, Colour: col(inkDark)}}, pics...)
	}
	c := newCellPage(p.cells(), ink, pics)
	if cr != nil && cr.Scale > 1 && pk == nil && sl == nil {
		c.pics = c.scaled(cr)
	}
	c.texts(p.prompt(), p.flags(), p.barShown(sl))
	if sp != nil {
		d, colour := p.sheetDraft(), nameColour(p.sheetOwn())
		if d != nil {
			colour = 11 // the member being made: as a well one's
		}
		c.sheet(sp, d, colour, portrait, cr)
	}
	if pk != nil {
		c.picker(pk)
	}
	if ed != nil {
		c.iconEditor(ed)
	}
	if p.barShown(sl) { // the bar on the prompt row
		for _, w := range p.barWords() {
			c.str(w.x, c.rows-1, inkWhite, w.label[:1])
			c.str(w.x+1, c.rows-1, inkGreen, w.label[1:])
		}
	}
	if sl != nil {
		c.box(1, 1, c.cols-2, c.rows-2, inkWhite, "Saved Games")
		if len(sl.Names) > 0 { // the column heads above the rows (slotList: text from column 3)
			w, lc := slotCols(c.cols)
			h := inkWhite
			c.str(3, 2, h, "Name")
			c.str(3+lc, 2, h, "Level")
			c.str(3+w-4, 2, h, "Date")
		}
		if !sl.CanSave && !sl.LoadOnly {
			c.str(2, c.rows-4, inkYellow, "No saving now") // promptRow-3
		}
		for _, w := range slotFooter(sl, c.cols) {
			c.str(w.x, c.rows-3, inkYellow, w.label) // promptRow-2
		}
	}
	return c
}

// barShown is whether the special screen's bar is on the prompt row.
func (p *Presenter) barShown(sl *Slots) bool {
	return (p.creation() != nil || p.iconEditor(p.ink()) != nil) && p.picker(p.ink()) == nil && sl == nil
}

// footWord is a word of the slot page's footer: its text, the key it
// stands for (footPick, L, D, footBack) and its column.
type footWord struct {
	label string
	key   byte
	x     int
}

// The footer words that are no letter: Enter (pick the row) and Esc (back).
const footPick, footBack byte = 1, 2

// slotFooter is the slot page's footer on a page cols wide, as the
// original-style F10 page's: spread over the row inside the frame.
func slotFooter(sl *Slots, cols int) []footWord {
	ws := []footWord{{"Enter=save", footPick, 0}, {"L=load", 'L', 0}, {"D=delete", 'D', 0}, {"Esc=back", footBack, 0}}
	switch {
	case sl.LoadOnly:
		ws = append([]footWord{{"Enter=load", footPick, 0}}, ws[2:]...)
	case !sl.CanSave:
		ws = ws[1:]
	case sl.NoLoad: // a hint key only for an act allowed
		ws = append(ws[:1:1], ws[2:]...)
	}
	gap := cols - 4
	for _, w := range ws {
		gap -= len(w.label)
	}
	gap /= len(ws) - 1
	x := 2
	for i := range ws {
		ws[i].x, x = x, x+len(ws[i].label)+gap
	}
	return ws
}

// texts lays out the game's text: the text pane, the prompt row, the view
// window's text, the party panel (roster, status line), after-combat and
// the credits.
func (c *cellPage) texts(prompt []PaneRun, f *game.Flags, bar bool) {
	for _, r := range prompt {
		if bar { // the bar owns the row: a prompt under it shows through its gaps
			break
		}
		c.run(r, c.rows-1)
	}
	if f.Credits && creditBoxes() == nil { // one frame on the page's edges, the lines laid out over the page inside it
		c.border(1, 1, c.cols-2, c.rows-2)
		var page []adventure.Credit
		x1, x2, y2 := c.cols, 0, 1
		for _, l := range credits() {
			if l.Page == f.CreditPage {
				page = append(page, l)
				x1, x2, y2 = min(x1, l.X), max(x2, l.X+len(l.Text)), max(y2, l.Y)
			}
		}
		// The rows spread from the original's 1..y2 over 1..rows-6 (a blank
		// row, the remaster's two, the frame below): only its blank rows
		// grow, so lines under one another stay together (a spread of every
		// row tore groups apart); the block centred across.
		used := map[int]bool{}
		for _, l := range page {
			used[l.Y] = true
		}
		var blanks []int
		for y := 1; y < y2; y++ {
			if !used[y] {
				blanks = append(blanks, y)
			}
		}
		extra := max(0, c.rows-7-(y2-1))
		grow := map[int]int{} // the rows each blank row adds, the extra shared out evenly
		for i, y := range blanks {
			grow[y] = extra*(i+1)/len(blanks) - extra*i/len(blanks)
		}
		dx := max(0, (c.cols-(x2-x1))/2) - x1
		for _, l := range page {
			y := l.Y
			for _, b := range blanks {
				if b < l.Y {
					y += grow[b]
				}
			}
			c.str(l.X+dx, y, inkEGA(l.Colour), l.Text) // a row above 1 (a title on the frame's top edge) stays where it is
		}
		c.remasterCredit()
		return
	}
	if f.Credits { // the original's boxes, an edge on its page's following this page's
		for _, b := range creditBoxes() {
			x2, y2 := b.X2, b.Y2
			if x2 >= 38 {
				x2 = c.cols - 2
			}
			if y2 >= 23 {
				y2 = c.rows - 2
			}
			c.border(b.X1, b.Y1, x2, y2)
		}
		// Centred on wider pages; on pages tall enough, a blank row after
		// each of creditGaps' rows (Pool: before "game created by", after it
		// and after the version).
		dx, gaps := max(0, (c.cols-40)/2), 0
		var page []adventure.Credit
		for _, l := range credits() {
			if l.Page == f.CreditPage {
				page = append(page, l)
			}
		}
		if len(page) == 0 {
			return
		}
		last := page[len(page)-1]
		gaps = min(len(creditGaps), max(0, c.rows-2-last.Y)) // as many as fit: the first ones, so no row stays on the lower box's border
		// On a page taller still, the blank rows between the sections under
		// the gaps grow, shared out evenly, so the lines fill the box down to
		// a blank row over the remaster's (the 30-row credits, por-go
		// b77c46ea): sections stay together, the box is never half empty.
		used := map[int]bool{}
		for _, l := range page {
			used[l.Y] = true
		}
		var blanks []int
		for y := last.Y; gaps == len(creditGaps) && gaps > 0 && y > slices.Max(creditGaps); y-- { // only once every gap is in
			if !used[y] {
				blanks = append(blanks, y)
			}
		}
		extra := 0
		if len(blanks) > 0 {
			extra = max(0, c.rows-6-(last.Y+gaps))
		}
		slices.Reverse(blanks)
		grow := map[int]int{}
		for i, y := range blanks {
			grow[y] = extra*(i+1)/len(blanks) - extra*i/len(blanks)
		}
		for _, l := range page {
			y := l.Y
			for _, g := range creditGaps[:gaps] {
				if l.Y > g {
					y++
				}
			}
			for _, b := range blanks {
				if b < l.Y {
					y += grow[b]
				}
			}
			c.str(l.X+dx, y, inkEGA(l.Colour), l.Text)
		}
		c.remasterCredit()
	}
}

// remasterLines is the remaster's one credit line, gobox's for every game: on
// every credits page and over every credits picture. Split in two where the
// window is too narrow for it in one row.
var remasterLines = [2]string{"Opinionated Remaster", "directed by Felix Ruzzoli using Claude"}

// remasterRows is remasterLines in the rows a page cols cells wide shows
// them: one, or two when the whole line does not fit with a cell each side.
func remasterRows(cols int) []string {
	if s := remasterLines[0] + " " + remasterLines[1]; len(s) <= cols-2 {
		return []string{s}
	}
	return remasterLines[:]
}

// remasterCredit puts the remaster credit centred in yellow, its last row
// the one above the frame's bottom edge (rows-3).
func (c *cellPage) remasterCredit() {
	rows := remasterRows(c.cols)
	for i, s := range rows {
		c.str((c.cols-len(s))/2, c.rows-2-len(rows)+i, inkYellow, s)
	}
}

// capOr is the end cap's place when b has that cap, else side's.
func capOr(b [Places]image.Image, cap, side int) int {
	if b[cap] != nil {
		return cap
	}
	return side
}

// cellBorder draws the window border's cells into m round the cells of r. The
// page boxes are text-grid cells drawn into the page's CPU image (uploaded
// once per View, cropped by the adventure frame), not window()'s GPU pass.
func cellBorder(m *image.RGBA, wb windowBorder, cells image.Point, r CellRect) {
	f := r.Framed()
	b := wb.pieces
	tile := func(i, x, y int) {
		if t := b[i]; t != nil {
			draw.Draw(m, image.Rect(x*8, y*8, x*8+8, y*8+8), t, t.Bounds().Min, draw.Over)
		}
	}
	tile(TopLeft, f.X1, f.Y1)
	tile(TopRight, f.X2, f.Y1)
	tile(BottomLeft, f.X1, f.Y2)
	tile(BottomRight, f.X2, f.Y2)
	for x := f.X1 + 1; x < f.X2; x++ {
		tile(Top, x, f.Y1)
		tile(Bottom, x, f.Y2)
	}
	for y := f.Y1 + 1; y < f.Y2; y++ {
		l, r := Left, Right
		if y == f.Y1+1 {
			l, r = capOr(b, LeftTopCap, l), capOr(b, RightTopCap, r)
		}
		if y == f.Y2-1 {
			l, r = capOr(b, LeftBottomCap, l), capOr(b, RightBottomCap, r)
		}
		tile(l, f.X1, y)
		tile(r, f.X2, y)
	}
	if wb.own {
		return // the game's own pieces
	}
	all := wb.edges
	es := func(key string, n int) []image.Image { // the edge's own table for length n, else its general one
		if t := all[key+"@"+strconv.Itoa(n)]; len(t) > 0 {
			return t
		}
		return all[key]
	}
	put := func(t image.Image, x, y int) {
		if t != nil {
			draw.Draw(m, image.Rect(x*8, y*8, x*8+8, y*8+8), t, t.Bounds().Min, draw.Src)
		}
	}
	run := func(seq []image.Image, n int, at func(i int) (int, int)) {
		for i := range n {
			x, y := at(i)
			put(seq[EdgeCell(i, n, len(seq))], x, y)
		}
	}
	h, w := f.Y2-f.Y1+1, f.X2-f.X1+1
	cols := cells.X
	if cols == 0 {
		cols = 40
	}
	for _, side := range []struct {
		key string
		x   int
	}{{"left", f.X1}, {"right", f.X2}} {
		key := side.key
		if side.x > 0 && side.x < cols-1 && len(es("divider", h)) > 0 { // between two boxes: the divider
			key = "divider"
		}
		if seq := es(key, h); len(seq) > 0 {
			run(seq, h, func(i int) (int, int) { return side.x, f.Y1 + i })
		}
	}
	for _, edge := range []struct {
		key, outer string
		y          int
		outerRow   bool
	}{{"rule", "top", f.Y1, f.Y1 == 0}, {"rule", "bottom", f.Y2, f.Y2 >= cells.Y-1}} {
		key := edge.key
		if edge.outerRow {
			key = edge.outer
		}
		if seq := es(key, w); len(seq) > 0 {
			run(seq, w, func(i int) (int, int) { return f.X1 + i, edge.y })
		}
	}
}

// draw draws the page into m (its pictures, the borders' cells, the text
// in the game's font).
func (c *cellPage) draw(m *image.RGBA, wb windowBorder, cells image.Point, note Note) {
	drawPictures(m, c.pics, image.Point{})
	for _, r := range c.boxes {
		cellBorder(m, wb, cells, r)
	}
	for y, row := range c.cells {
		for x, k := range row {
			if k.set {
				glyphAt(m, note.Font(k.ch, true), x*8, y*8, k.fg, k.bg)
			}
		}
	}
}

// fit is the largest rectangle of b's shape inside r, centred across and
// on r's bottom edge (a portrait stands on its frame's foot).
func fit(b, r image.Rectangle) image.Rectangle {
	w, h := r.Dx(), b.Dy()*r.Dx()/max(b.Dx(), 1)
	if h > r.Dy() {
		w, h = b.Dx()*r.Dy()/max(b.Dy(), 1), r.Dy()
	}
	return image.Rect(0, 0, w, h).Add(r.Min.Add(image.Pt((r.Dx()-w)/2, r.Dy()-h)))
}

// sheet lays the character sheet out (ovr20:004A CharacterSheet).
// d is the draft being made (nil: a member's sheet), nameColour the name's
// ink and portrait the PNG or default head in the portrait's box.
func (c *cellPage) sheet(s *game.SheetPage, d *game.Creation, nameColour int, portrait image.Image, cr *FrontCreation) {
	right, bottom := c.cols-2, c.rows-3
	c.unset(image.Rect(0, 0, c.cols*8, (bottom+2)*8)) // the sheet's page: ClearScreen came first (the 3D view's status line once wrote over it)
	c.border(1, 1, right, bottom)
	c.str(1, 1, nameColour, s.Name)
	if s.NPC {
		c.str(len(s.Name)+3, 1, inkGreen, "(NPC)")
	}
	if d != nil && rowFields(d) {
		for i, f := range d.Fields {
			c.str(1, 3+i, inkWhite, picked(f))
		}
	} else {
		x := 1
		c.str(x, 3, inkWhite, s.Sex)
		x += len(s.Sex) + 1
		c.str(x, 3, inkWhite, s.Race)
		x += len(s.Race) + 1
		c.str(x, 3, inkWhite, "Age "+strconv.Itoa(s.Age))
		c.str(1, 4, inkWhite, s.Alignment)
		c.str(1, 5, inkWhite, s.Class)
	}
	if d == nil || len(d.Stats) == 0 {
		for i, l := range []string{"STR ", "INT ", "WIS ", "DEX ", "CON ", "CHA "} { // and showStat
			y := i + 7
			c.str(1, y, inkGreen, l)
			c.clear(5, y, 11, y)
			x := 5
			if s.Stats[i] < 10 {
				x++
			}
			c.str(x, y, inkGreen, strconv.Itoa(s.Stats[i]))
			if i == 0 {
				c.str(7, 7, inkGreen, s.Str00)
			}
		}
	}
	if d != nil {
		for i, r := range d.Stats { // the game's own abilities, as given
			y, x := i+7, StatCol(r.Label)
			c.str(1, y, inkGreen, r.Label+" ")
			c.clear(x, y, x+6, y)
			c.str(x, y, inkGreen, fmt.Sprintf("%2d", r.Value)+r.Extra)
		}
	}
	if s.Saves[0] > 0 { // extra: the saving throws beside the abilities
		for i, l := range []string{"Paralyze", "Petrify", "Rod", "Breath", "Spell"} {
			c.str(13, 7+i, inkWhite, l)
			c.str(22, 7+i, inkGreen, strconv.Itoa(s.Saves[i]))
		}
	}
	if !s.Creating {
		c.str(1, 15, inkWhite, "Level")
		c.str(7, 15, inkWhite, s.Levels)
		c.str(17, 15, inkWhite, "Exp "+strconv.Itoa(s.Exp))
		if s.Next > 0 {
			col := inkWhite
			if s.Exp >= s.Next {
				col = inkGreen
			}
			c.str(17, 16, col, "Next "+strconv.Itoa(s.Next))
		}
	}
	y := 0x11 // showCombatStats
	if d != nil && d.Derived != nil {
		// the game's figures in order, two to a row
		for i, f := range d.Derived {
			x, fy := 1+16*(i%2), y+i/2
			c.str(x, fy, inkWhite, f.Label)
			c.str(x+len(f.Label)+1, fy, inkGreen, f.Value)
		}
	} else {
		c.str(1, y, inkWhite, "AC    ")
		c.str(4, y, inkGreen, strconv.Itoa(s.AC))
		c.str(1, y+1, inkWhite, "HP    ")
		hp := inkGreen
		if s.HP < s.MaxHP {
			hp = inkYellow
		}
		c.str(4, y+1, hp, strconv.Itoa(s.HP))
		c.str(9, y, inkWhite, "THAC0   ")
		c.str(15, y, inkGreen, strconv.Itoa(s.THAC0))
		c.str(8, y+1, inkWhite, "Damage        ")
		c.str(15, y+1, inkGreen, s.Damage)
		c.str(0x16+3, y+1, inkWhite, "Movement   ")
		c.str(0x16+0xC, y+1, inkGreen, strconv.Itoa(s.Move))
	}
	if !s.Creating {
		y = 0x14
		c.str(1, y, inkWhite, "Status")
		c.str(8, y, inkGreen, s.Status)
		if s.Effects != "" {
			c.str(1, 0x13, inkWhite, "Effects")
			c.str(9, 0x13, inkGreen, s.Effects)
		}
	}
	for j, f := range extraFields(s.Weapon, s.Armour, s.Encumbrance, s.Coins) { // the cells of each: Weapon, Armour, Load, Coins
		at := [...][2]int{{1, 13}, {1, 14}, {1, 21}, {17, 21}}[j]
		if f[1] != "" {
			c.str(at[0], at[1], inkWhite, f[0])
			c.str(at[0]+len(f[0])+1, at[1], inkGreen, f[1])
		}
	}
	if b := sheetPortraitBox(s); b != (CellRect{}) {
		c.border(b.X1, b.Y1, b.X2, b.Y2)
		if portrait != nil {
			c.overs = append(c.overs, overlay{Img: portrait, At: fit(portrait.Bounds(), b.Pixels())}) // the whole box: the rope runs outside its cells
		}
	}
	k, box := 1, CellRect{}
	if cr != nil { // the creation's combat icon, from a PNG
		k, box = max(1, cr.Scale), cr.Icon
	}
	for p := range 2 {
		if im := ownPose(s.IconPNG, p, false); im != nil {
			x, y := box.X1+5*k*p, box.Y1
			c.overs = append(c.overs, overlay{Img: im, At: image.Rect(x*8, y*8, x*8+24*k, y*8+24*k)})
		}
	}
	if s.Spells != "" { // SpellList's title over the name row
		c.box(1, 1, right, 1, inkBlack, "")
		c.str(1, 1, nameColour, s.Name+"'s")
		c.str(len(s.Name)+4, 1, inkGreen, s.Spells)
	}
}

// scaled takes the creation sheet's portrait and icon pictures, which the
// engine draws at 1x from the portrait box's and the icon box's corners,
// off the page as crisp overlays cr.Scale times their size; the rest stay.
func (c *cellPage) scaled(cr *FrontCreation) []Picture {
	port := cr.Head.Pixels().Union(cr.Body.Pixels())
	var keep []Picture
	for _, pc := range c.pics {
		var from image.Point
		switch {
		case pc.Fill || pc.Img == nil:
			keep = append(keep, pc)
			continue
		case pc.At.Overlaps(port):
			from = port.Min
		case pc.At.Overlaps(cr.Icon.Pixels()):
			from = cr.Icon.Pixels().Min
		default:
			keep = append(keep, pc)
			continue
		}
		min := from.Add(pc.At.Min.Sub(from).Mul(cr.Scale))
		c.overs = append(c.overs, overlay{Img: pc.Img, At: image.Rectangle{min, min.Add(pc.At.Size().Mul(cr.Scale))}, Crisp: true})
	}
	return keep
}

// picker lays a grid picker's page out (showPicker): the box, the pager,
// the saved characters' names, classes and PNG portraits.
func (c *cellPage) picker(pk *FrontPicker) {
	c.border(1, 1, c.cols-2, pk.Bottom)
	if pk.Title != "" { // centred on the box's top edge
		c.str((c.cols-len(pk.Title))/2, 1, inkWhite, pk.Title)
	}
	if pk.Pages > 1 {
		c.str(pk.Prev.X1, pk.Prev.Y1, inkWhite, "< page "+strconv.Itoa(pk.Page+1)+"/"+strconv.Itoa(pk.Pages)+" >")
	}
	for k, r := range pk.Cells {
		i := pk.Page*pk.PerPage + k
		if i < len(pk.Portraits) && pk.Portraits[i] != nil {
			box := r // a pool picture fills its whole cell (a body's is 6 rows, not the head's 5)
			if len(pk.Labels) > 0 { // a roster: the 11 by 5 portrait left of the name and class
				box = CellRect{X1: r.X1, Y1: r.Y1, X2: r.X1 + 10, Y2: r.Y1 + 4}
			}
			c.overs = append(c.overs, overlay{Img: pk.Portraits[i], At: fit(pk.Portraits[i].Bounds(), box.Pixels()), Crisp: true}) // nearest neighbour
		}
		if i < len(pk.Labels) {
			c.str(r.X1+12, r.Y1+1, inkWhite, pk.Labels[i])
		}
		if i < len(pk.Classes) && pk.Classes[i] != "" {
			c.str(r.X1+12, r.Y1+3, inkGreen, pk.Classes[i])
		}
	}
}

// iconEditor lays the combat icon editor's page out: the frame, the poses,
// the words, the swatches and the palette (their colours the engine's).
func (c *cellPage) iconEditor(ed *FrontIconEditor) {
	c.border(1, 1, c.cols-2, c.rows-3) // the bar under it
	for p, r := range ed.PoseAt {
		c.pics = append(c.pics, Picture{At: r.Pixels(), Fill: true, Colour: col(inkDark)}) // the pose on grey (the page's, not the combat palette's dark): its black parts show
		img := ed.Poses[p]
		if ed.Indexed {
			img = picImage(ed.PosePix[p], c.ink)
		}
		if img != nil {
			c.overs = append(c.overs, overlay{Img: img, At: r.Pixels(), Crisp: true})
		}
	}
	c.str(ed.PoseAt[0].X1+2, ed.PoseAt[0].Y1-1, inkWhite, "ready")
	c.str(ed.PoseAt[1].X1+2, ed.PoseAt[1].Y1-1, inkWhite, "action")
	if ed.Custom { // a dropped PNG: it can only go
		c.str(ed.Remove.X1, ed.Remove.Y1, inkGreen, "Remove")
		return
	}
	c.str(ed.Head.X1, ed.Head.Y1, inkGreen, "Head")
	c.str(ed.Weapon.X1, ed.Weapon.Y1, inkGreen, "Weapon")
	size := "small"
	if ed.Large {
		size = "large"
	}
	c.str(ed.Size.X1, ed.Size.Y1, inkGreen, "Size "+size)
	c.str(ed.Names.X1, ed.Names.Y1-1, inkWhite, "Colours")
	c.str(ed.Parts.X1, ed.Parts.Y1-1, inkWhite, "1  2")
	for i, name := range iconParts {
		c.str(ed.Names.X1, ed.Names.Y1+i, inkGreen, name)
	}
	swatch := func(x, y int, colour, frame color.RGBA) { // 2 cells wide, framed
		r := image.Rect(x*8, y*8, x*8+16, y*8+8)
		c.pics = append(c.pics, Picture{At: r, Fill: true, Colour: frame}, Picture{At: r.Inset(1), Fill: true, Colour: colour})
	}
	for i := range iconParts {
		for sh := range 2 {
			frame := col(inkDark) // the page's greys, not the combat palette's
			if i == ed.Part && sh == ed.Shade {
				frame = col(inkWhite)
			}
			sw := ed.Swatch[i][sh]
			if ed.Indexed {
				sw = c.ink[ed.SwatchIx[i][sh]]
			}
			swatch(ed.Parts.X1+3*sh, ed.Parts.Y1+i, sw, frame)
		}
	}
	for i, n := range ed.Colours {
		if n >= 0 {
			pi := ed.PaletteInk[i]
			if ed.Indexed {
				pi = c.ink[ed.PaletteIx[i]]
			}
			swatch(ed.Palette.X1+2*(i%8), ed.Palette.Y1+i/8, pi, col(inkDark))
		}
	}
}

// iconClick is what a click on cell (cx, cy) of the icon editor does.
func iconClick(ed *FrontIconEditor, cx, cy int) (Action, bool) {
	in := func(r CellRect) bool { return cx >= r.X1 && cx <= r.X2 && cy >= r.Y1 && cy <= r.Y2 }
	if ed.Custom {
		if in(ed.Remove) {
			return Action{Kind: ActIconRemove}, true
		}
		return Action{}, false
	}
	head, weapon := Action{Kind: ActPicker, N: PickIconHead}, Action{Kind: ActPicker, N: PickWeapon}
	switch {
	case in(ed.Head):
		return head, true
	case in(ed.Weapon):
		return weapon, true
	case in(ed.Size):
		return Action{Kind: ActIconSize}, true
	case in(ed.Parts):
		return Action{Kind: ActIconColour, N: cy - ed.Parts.Y1, X: (cx - ed.Parts.X1) / 3, Y: -1}, true
	case in(ed.Palette):
		if v := ed.Colours[(cy-ed.Palette.Y1)*8+(cx-ed.Palette.X1)/2]; v >= 0 {
			return Action{Kind: ActIconColour, N: ed.Part, X: ed.Shade, Y: v}, true
		}
		return Action{}, false
	}
	for _, r := range ed.PoseAt { // the head on the poses' upper part, the weapon below
		mid := r.Y1 + (r.Y2-r.Y1+1)*4/9
		if in(CellRect{X1: r.X1, Y1: r.Y1, X2: r.X2, Y2: mid - 1}) {
			return head, true
		}
		if in(CellRect{X1: r.X1, Y1: mid, X2: r.X2, Y2: r.Y2}) {
			return weapon, true
		}
	}
	return Action{}, false
}

// inkEGA is ink k's EGA colour number (InkNone: the body text's, inkGreen).
func inkEGA(k game.Ink) int {
	if n, ok := k.EGA(); ok {
		return n
	}
	return inkGreen
}
