package present

import (
	"fmt"
	"gobox/front"
	"image"
	"image/color"
	"math"
	"slices"
	"strings"
	"time"

	"gobox/gpu"
	"gobox/ui"

	"gobox/input"
)

// The page screens taking decisions whose questions are the presenter's:
// the creation sheet (View.Creation: its menu, the drop-downs, the typed
// lines), the grid pickers (View.Picker: the cursor, the frame, the
// pages), the icon editor's bar (View.IconDrop) and the save slots
// (View.Slots: the list, the name, Overwrite and Delete).

// Sub-screens of a page screen.
const (
	scrMain   = iota
	scrField  // a drop-down (scr.field)
	scrType   // a line being typed
	scrAsk    // a yes/no question
	scrPowers // the creation's power picks (Creation.Powers)
	scrHow    // a new character: Roll or Point Buy (Creation.PointBuy)
	scrBuy    // the point buy (pointbuy.go)
)

const scrRows = 14 // the lines a page screen's menu shows at a time

type screenUI struct {
	ptr     image.Point // the pointer's cell, as last seen: only a moved one takes the picker's cursor
	up      string      // the screen up (screenName): another one starts on its main menu
	sub     int
	sel     int
	top     int
	field   int
	prompt  string        // scrType: the question ...
	text    string        // ... the line typed ...
	caret   int           // ... its cursor ...
	max     int           // ... how long it may get ...
	act     front.ActKind // ... and what it sends
	then    func(string)  // ... or what takes the line instead (Minimal's next ability)
	n       int
	ask     string // scrAsk: the question; yes on Yes, back to sub on No
	yes     func()
	back    int
	backSel int             // the row the question was asked from
	mainSel int             // the main menu's row that opened the submenu up: back restores it
	cur     int             // the picker's cursor
	at      Rect            // where the page went on the window ...
	pw, ph  int             // ... and its size in pixels
	anchor  *front.CellRect // the drop-down's field: the menu opens on it
	newChar bool            // ActCreate sent: the creation sheet coming up asks Roll or Point Buy
	buy     []int           // scrBuy: the scores bought ...
	buyStat int             // ... and the one Raise and Lower change
}

func screenName(v front.View) string {
	switch {
	case v.Picker != nil:
		return "picker"
	case v.Creation != nil:
		return "creation"
	case v.Slots != nil:
		return "slots"
	case v.IconDrop && askMenu(v) == nil:
		return "icon"
	}
	return ""
}

// pageScreen: one of them is up.
func pageScreen(v front.View) bool { return screenName(v) != "" }

func (p *Presenter) sendA(a front.Action) {
	if !p.slotAct(a) {
		p.inv.acts = append(p.inv.acts, a)
	}
}

// screenItems are the lines of the page screen up and its title.
func (p *Presenter) screenItems(v front.View) (string, []ui.MenuRow) {
	s := &p.scr
	act := func(k front.ActKind, n int) func() { return func() { p.send(k, n) } }
	switch s.sub {
	case scrAsk:
		return s.ask, []ui.MenuRow{{Text: "Yes", Key: 'Y', Do: func() { s.open(scrMain); s.yes() }}, {Text: "No", Key: 'N', Do: func() { s.sub, s.sel = s.back, s.backSel }}}
	case scrType:
		return s.prompt, []ui.MenuRow{{Text: caretLine(s.text, s.caret, s.max, p.blink()), Key: 0, Do: nil}}
	}
	switch screenName(v) {
	case "creation":
		cr := v.Creation
		switch s.sub {
		case scrField:
			fl := crField(cr, s.field)
			var it []ui.MenuRow
			for _, i := range fieldOrder(fl) {
				it = append(it, ui.MenuRow{Text: fl.Options[i].Text, Key: 0, Do: func() {
					f := s.field
					s.open(scrMain)
					if s.anchor = nil; f >= len(cr.Fields) {
						s.open(scrPowers) // back to the picks
					}
					p.sendX(front.ActSetField, f, i)
				}})
			}
			return "", it
		case scrPowers:
			var it []ui.MenuRow
			for i, fl := range cr.Powers {
				t := "(choose)"
				if fl.Cur >= 0 {
					t = fl.Options[fl.Cur].Text
				}
				it = append(it, ui.MenuRow{Text: upFirst(fl.Name) + ": " + t, Key: 0, Do: func() { p.openField(cr, len(cr.Fields)+i); s.anchor = nil }})
			}
			return "Powers", append(it, ui.MenuRow{Text: "Exit", Key: 'E', Do: func() { s.open(scrMain) }})
		case scrHow:
			return "Abilities", []ui.MenuRow{{Text: "Roll", Key: 'R', Do: func() { s.open(scrMain) }}, {Text: "Point Buy", Key: 'P', Do: func() { p.openBuy(cr) }}}
		case scrBuy:
			return p.buyItems(cr)
		}
	case "slots":
		sl := v.Slots
		named := func(k front.ActKind, name string) func() {
			return func() { s.open(scrMain); p.sendA(front.Action{Kind: k, Text: name}) }
		}
		var it []ui.MenuRow
		if sl.CanSave {
			it = append(it, ui.MenuRow{Text: "New Slot", Key: 'N', Do: func() { p.typeLine("Save as", "", 20, front.ActSaveSlot, 0) }})
		}
		w, lc := slotCols(max(v.Cells.X, 40))
		for i, name := range sl.Names {
			lv := ""
			if sl.Levels[i] > 0 {
				lv = fmt.Sprintf("%3d", sl.Levels[i])
			}
			row := fmt.Sprintf("%-*s%-*s%*s", lc, upTo(name, lc-1), w-lc-len(sl.Times[i]), lv, len(sl.Times[i]), sl.Times[i])
			it = append(it, ui.MenuRow{Text: row, Key: 0, Do: func() {
				if !sl.CanSave {
					named(front.ActLoad, name)()
					return
				}
				p.askThen("Overwrite "+name+"?", named(front.ActSaveSlot, name))
			}})
		}
		if len(it) == 0 {
			it = append(it, ui.MenuRow{Text: "No saved games yet", Key: 0, Do: nil})
		}
		return "Saved Games", append(it, ui.MenuRow{Text: "Exit", Key: 'E', Do: act(front.ActClose, 0)})
	}
	return "", nil
}

func (p *Presenter) blink() bool { return time.Now().UnixMilli()/400%2 == 0 }

// typeLine starts a typed line: prompt, starting as text, sending act N
// with the line on Enter.
func (p *Presenter) typeLine(prompt, text string, n int, act front.ActKind, an int) {
	s := &p.scr
	s.open(scrType)
	s.back, s.sub, s.prompt, s.text, s.caret, s.max, s.act, s.n = scrMain, scrType, prompt, text, len(text), n, act, an
	s.anchor, s.then = nil, nil
}

// minimal asks the minimums one ability after another, STR first, each
// showing the last one (a blank line keeps it), then rolls to them.
func (p *Presenter) minimal(cr *front.Creation) {
	cur := strings.Fields(cr.Minimums)
	got := make([]string, 0, 6)
	var ask func()
	ask = func() {
		i := len(got)
		old := ""
		if i < len(cur) {
			old = cur[i]
		}
		p.typeLine("Minimum "+statNames[i]+" ("+old+")", "", 6, front.ActRoll, 1)
		p.scr.then = func(t string) {
			if t = strings.TrimSpace(t); t == "" {
				t = "-" // not a value: the last one stays
			}
			if got = append(got, t); len(got) < 6 {
				ask()
				return
			}
			p.sendA(front.Action{Kind: front.ActRoll, N: 1, Text: strings.Join(got, " ")})
		}
	}
	ask()
}

// fieldOrder is the drop-down's rows: the value the field holds first,
// over the field's text, the rest below in their order.
func fieldOrder(fl front.Field) []int {
	var order []int
	if fl.Cur >= 0 {
		order = append(order, fl.Cur)
	}
	for i := range fl.Options {
		if i != fl.Cur {
			order = append(order, i)
		}
	}
	return order
}

// typeAt starts a typed line in place of the sheet's text at r.
func (p *Presenter) typeAt(r front.CellRect, prompt string, n int, act front.ActKind, an int) {
	p.typeLine(prompt, "", n, act, an)
	p.scr.anchor = &r
}

// askThen asks q; yes runs on Yes.
func (p *Presenter) askThen(q string, yes func()) {
	s := &p.scr
	if s.sub == scrMain {
		s.mainSel = s.sel
	}
	s.back, s.backSel, s.sub, s.ask, s.yes, s.sel = s.sub, s.sel, scrAsk, q, yes, 1
}

// openField opens field f's drop-down on its text, the value it holds highlighted.
func (p *Presenter) openField(cr *front.Creation, f int) {
	s := &p.scr
	s.open(scrField)
	s.field, s.sel = f, 0 // the value it holds: the first row
	at := crField(cr, f).At
	s.anchor = &at
}

// drawScreen draws the page screen up over its page: its menu, the
// picker's frame.
func (p *Presenter) drawScreen(dst *gpu.Image, v front.View) {
	s := &p.scr
	if n := screenName(v); n != s.up {
		s.up, s.sub, s.sel, s.top, s.anchor = n, scrMain, 0, 0, nil
		if n == "creation" && s.newChar && v.Creation.PointBuy {
			s.sub, s.mainSel = scrHow, 0
		}
		s.newChar = false
		if v.Picker != nil {
			s.cur = v.Picker.Cur
		}
		if sl := v.Slots; sl != nil { // on the game last loaded or saved
			p.pointer = p.camp.pointer // the resting pointer does not take it; only a moved one does
			if i := slices.Index(sl.Names, sl.Current); i >= 0 && sl.Current != "" {
				s.sel = i
				if sl.CanSave {
					s.sel++ // under New Slot
				}
			}
		}
	}
	p.winRects = p.winRects[:0]
	if pk := v.Picker; pk != nil {
		s.cur = min(max(s.cur, 0), pk.N-1)
		if i := s.cur - pk.Page*pk.PerPage; i >= 0 && i < len(pk.Cells) {
			r := p.cellRect(pk.Cells[i])
			selectOutline(r, max(1, int(s.at.W/float64(s.pw))), func(e image.Rectangle, c color.RGBA) { dst.SubImage(e).(*gpu.Image).Fill(c) })
		}
		if s.sub != scrAsk {
			return
		}
	}
	if sp := p.spots(v); sp != nil && s.sub == scrMain { // the sheet's parts and the bar: the focus framed
		s.sel = min(max(s.sel, 0), len(sp)-1)
		selectOutline(p.cellRect(sp[s.sel].r), max(1, int(s.at.W/float64(s.pw))), func(e image.Rectangle, c color.RGBA) { dst.SubImage(e).(*gpu.Image).Fill(c) })
		return
	}
	title, items := p.screenItems(v)
	lines, keys, off := make([]string, len(items)), p.screenKeys(v, items), ui.RowsOff(items)
	for i, it := range items {
		lines[i] = it.Text
	}
	if s.sub == scrType { // the line being typed is no row to pick
		off[0] = false
	}
	if s.sub == scrField && v.Creation != nil { // grey: picking it changes what depends on it
		fl := crField(v.Creation, s.field)
		for r, i := range fieldOrder(fl) {
			off[r] = !fl.Options[i].Fits
		}
	}
	s.sel = min(max(s.sel, 0), len(lines)-1)
	if screenName(v) == "slots" && s.sub == scrMain {
		p.slotList(dst, lines, keys, off)
		return
	}
	s.top = max(0, min(max(s.top, s.sel-scrRows+1), s.sel, len(lines)-scrRows)) // 0 on an empty list (sel -1: lines[-1:] panicked)
	end := min(len(lines), s.top+scrRows)
	sel := s.sel - s.top
	if s.sub == scrType {
		sel = -1
	}
	if a := s.anchor; a != nil && (s.sub == scrField || s.sub == scrType) { // the drop-down or the typed line on its field
		k := s.at.W / float64(s.pw)
		w := 0
		for _, l := range lines {
			w = max(w, len(l))
		}
		r := p.cellRect(*a)
		ky := s.at.H / float64(s.ph)
		cx := int(math.Round(float64(r.Min.X) - k*8 + k*float64((w+2)*8)/2)) // the first row's text on the field's, to the pixel: the box's border left of it, no blank column
		rects := slices.Clone(p.menuRects)
		p.menuTight = true
		defer func() { p.menuTight = false }()
		p.drawMenu(dst, p.glyph, "", lines[s.top:end], keys[s.top:end], off[s.top:end], sel, k, cx, int(math.Round(float64(r.Min.Y)-ky*8)))
		p.winRects, p.menuRects = slices.Clone(p.menuRects), rects
		p.menuBar(dst, p.winRects, len(lines), end-s.top, &s.top, &s.sel, k)
		return
	}
	k := p.keyWindow(dst, title, lines[s.top:end], keys[s.top:end], off[s.top:end], sel, false)
	p.menuBar(dst, p.winRects, len(lines), end-s.top, &s.top, &s.sel, k)
}

// slotKey does the slot page's footer key h on the selected save: Enter
// picks the row, L loads, D deletes after a Yes, Esc leaves.
func (p *Presenter) slotKey(v front.View, h byte) {
	s, sl := &p.scr, v.Slots
	switch h {
	case footPick:
		p.screenPick(v, s.sel-s.top)
		return
	case footBack:
		p.send(front.ActClose, 0)
		return
	}
	i := s.sel
	if sl.CanSave {
		i-- // New Slot
	}
	if i < 0 || i >= len(sl.Names) {
		return
	}
	name := sl.Names[i]
	switch h {
	case 'L':
		p.sendA(front.Action{Kind: front.ActLoad, Text: name})
	case 'D':
		p.askThen("Delete "+name+"?", func() { s.open(scrMain); p.sendA(front.Action{Kind: front.ActDeleteSlot, Text: name}) })
	}
}

// slotFootAt is the footer word of the slot page under (x, y), or 0.
func (p *Presenter) slotFootAt(v front.View, x, y int) byte {
	s := &p.scr
	if v.Slots == nil || s.sub != scrMain || s.pw == 0 {
		return 0
	}
	for _, w := range slotFooter(v.Slots, s.pw/8) {
		if image.Pt(x, y).In(p.cellRect(front.CellRect{X1: w.x, Y1: s.ph/8 - 3, X2: w.x + len(w.label) - 1, Y2: s.ph/8 - 3})) {
			return w.key
		}
	}
	return 0
}

// screenKeys are the hotkeys of the page screen's rows: the saves on the
// slot list and the training hall's new spells have none (no digits
// either), the rest as every menu.
func (p *Presenter) screenKeys(v front.View, items []ui.MenuRow) []byte {
	learn := v.Training != nil && len(v.Training.Learn) > 0
	if (screenName(v) != "slots" || p.scr.sub != scrMain) && p.scr.sub != scrField && !learn { // the drop-downs: no hotkeys, no digits
		return ui.RowKeys(items)
	}
	k := make([]byte, len(items))
	for i, r := range items {
		k[i] = r.Key
	}
	return k
}

// slotCols is the slot rows' text width on a page cols wide and the
// level column's start (5 wide, "Level"); the date ends the row.
func slotCols(cols int) (w, lc int) {
	w = cols - 6
	return w, w/2 - 2
}

// slotList draws the save slots' list on the page itself, inside the
// Saved Games box (screenPage): a row a line from row 3 down to above
// "No saving now", the name at column 3 and the time at 25, as the
// original-style F10 page; the rows scroll with the selection.
func (p *Presenter) slotList(dst *gpu.Image, lines []string, keys []byte, off []bool) {
	s := &p.scr
	cols, n := s.pw/8, max(1, s.ph/8-8)
	s.top = min(max(s.top, s.sel-n+1), s.sel, max(0, len(lines)-n))
	end := min(len(lines), s.top+n)
	w := cols - 4 // columns 2 to cols-3, inside the border
	if gpu.Sized(&p.mc.Img, w*8, (end-s.top)*8) {
		p.mc.Buf = make([]byte, w*8*(end-s.top)*8*4)
	}
	clear(p.mc.Buf)
	set := func(x, y int, c color.RGBA) {
		o := (y*w*8 + x) * 4
		p.mc.Buf[o], p.mc.Buf[o+1], p.mc.Buf[o+2], p.mc.Buf[o+3] = c.R, c.G, c.B, 255
	}
	for r := s.top; r < end; r++ {
		y0 := (r - s.top) * 8
		if r == s.sel && inkBlue >= 0 {
			for y := y0; y < y0+8; y++ {
				for x := range w * 8 {
					set(x, y, col(inkBlue))
				}
			}
		}
		t, hk := ui.HotText(lines[r], keys[r])
		for i := 0; i < len(t) && i+1 < w; i++ {
			c := col(inkGreen)
			switch {
			case off[r]:
				c = col(inkDark)
			case i == hk, r == s.sel && inkBlue < 0:
				c = col(inkWhite)
			}
			g := p.glyph(t[i])
			for y := range 8 {
				for b := range 8 {
					if g[y]&(0x80>>b) != 0 {
						set((i+1)*8+b, y0+y, c)
					}
				}
			}
		}
	}
	p.mc.Img.WritePixels(p.mc.Buf)
	at := p.cellRect(front.CellRect{X1: 2, Y1: 3, X2: cols - 3, Y2: 3 + end - s.top - 1})
	op := &gpu.DrawImageOptions{}
	op.GeoM.Scale(float64(at.Dx())/float64(w*8), float64(at.Dy())/float64((end-s.top)*8))
	op.GeoM.Translate(float64(at.Min.X), float64(at.Min.Y))
	dst.DrawImage(p.mc.Img, op)
	for r := s.top; r < end; r++ {
		p.winRects = append(p.winRects, p.cellRect(front.CellRect{X1: 2, Y1: 3 + r - s.top, X2: cols - 3, Y2: 3 + r - s.top}))
	}
	p.menuBar(dst, p.winRects, len(lines), end-s.top, &s.top, &s.sel, float64(at.Dy())/float64((end-s.top)*8))
}

// cellRect is a cell rectangle of the page on the window.
func (p *Presenter) cellRect(c front.CellRect) image.Rectangle {
	s := &p.scr
	kx, ky := s.at.W/float64(s.pw), s.at.H/float64(s.ph)
	return image.Rect(int(s.at.X+float64(c.X1*8)*kx), int(s.at.Y+float64(c.Y1*8)*ky), int(s.at.X+float64((c.X2+1)*8)*kx), int(s.at.Y+float64((c.Y2+1)*8)*ky))
}

// screenPick picks shown line i of the page screen's menu.
func (p *Presenter) screenPick(v front.View, i int) {
	_, items := p.screenItems(v)
	if i += p.scr.top; i >= 0 && i < len(items) && items[i].Do != nil {
		p.scr.sel = i
		items[i].Do()
	}
}

// screenKey takes key k on a page screen.
func (p *Presenter) screenKey(k uint16, v front.View) bool {
	s := &p.scr
	if s.sub == scrType {
		switch done, cancel, _ := editKey(k, &s.text, &s.caret, s.max, false); {
		case done && s.then != nil:
			s.open(s.back)
			s.then(s.text)
		case done:
			s.open(s.back)
			p.sendA(front.Action{Kind: s.act, N: s.n, Text: s.text})
		case cancel:
			s.open(s.back)
		}
		return true
	}
	if pk := v.Picker; pk != nil && s.sub != scrAsk {
		return p.pickerKey(k, pk)
	}
	if sp := p.spots(v); sp != nil && s.sub == scrMain {
		return p.spotKey(k, v, sp)
	}
	if sl := v.Slots; sl != nil && s.sub == scrMain {
		h := byte(k)
		if h >= 'a' && h <= 'z' {
			h -= 'a' - 'A'
		}
		switch actOf(k) { // the pad's X and Y: L and D
		case input.Overview:
			h, k = 'L', 'L'
		case input.Minimap:
			h, k = 'D', 'D'
		}
		if (h == 'L' && !sl.LoadOnly || h == 'D') && k < 128 {
			p.slotKey(v, h)
			return true
		}
	}
	if s.sub == scrBuy && p.buyKey(k) {
		return true
	}
	_, items := p.screenItems(v)
	n := len(items)
	keys := p.screenKeys(v, items)
	i, back, ok := ui.MenuNav(actOf(k), k, &s.sel, keys)
	switch {
	case back && s.sub == scrAsk:
		s.sub, s.sel = s.back, s.backSel
	case back && s.sub != scrMain:
		s.open(scrMain)
		s.anchor = nil
	case back:
		p.screenPick(v, n-1-s.top) // Exit
	case i >= 0:
		p.screenPick(v, i-s.top)
	}
	return ok
}

// pickerKey moves the picker's cursor (turning the page under it), Enter
// takes the picture, Esc leaves.
func (p *Presenter) pickerKey(k uint16, pk *front.Picker) bool {
	s := &p.scr
	c, n, cols, per := s.cur, pk.N, pk.Cols, pk.PerPage
	switch actOf(k) {
	case input.Up:
		c = max(c-cols, c%cols)
	case input.Down:
		if c+cols < n {
			c += cols
		}
	case input.Left:
		c = max(c-1, 0)
	case input.Right:
		c = min(c+1, n-1)
	case input.Home:
		c = 0
	case input.End:
		c = n - 1
	case input.PgUp:
		c = max(c-per, 0)
	case input.PgDn:
		c = min(c+per, n-1)
	case input.Enter, input.Space, input.PadA:
		p.pickerTake(pk, c)
		return true
	case input.Esc:
		p.send(front.ActClose, 0)
		return true
	default:
		return false
	}
	p.pickerTo(pk, c)
	return true
}

// pickerTo moves the cursor to c, on its page.
func (p *Presenter) pickerTo(pk *front.Picker, c int) {
	p.scr.cur = c
	if pg := c / pk.PerPage; pg != pk.Page {
		p.send(front.ActPage, pg)
	}
}

// pickerTake takes picture i, asking first where the picker asks.
func (p *Presenter) pickerTake(pk *front.Picker, i int) {
	if i < 0 || i >= pk.N {
		return
	}
	take := func() { p.send(front.ActPicked, i) }
	if pk.Ask != "" && i < len(pk.Labels) {
		p.askThen(fmt.Sprintf(pk.Ask, pk.Labels[i]), take)
		return
	}
	take()
}

// screenMouse takes the pointer on a page screen's page at cell (cx, cy):
// the picker's pictures and pager, the creation sheet's parts, the icon
// editor's.
func (p *Presenter) screenMouse(cx, cy int, left bool, v front.View) bool {
	s := &p.scr
	in := func(r front.CellRect) bool { return image.Pt(cx, cy).In(image.Rect(r.X1, r.Y1, r.X2+1, r.Y2+1)) }
	moved := image.Pt(cx, cy) != s.ptr
	s.ptr = image.Pt(cx, cy)
	if pk := v.Picker; pk != nil && s.sub == scrMain {
		for i, r := range pk.Cells {
			if in(r) {
				if moved || left {
					s.cur = pk.Page*pk.PerPage + i
				}
				if left {
					p.pickerTake(pk, s.cur)
				}
				return true
			}
		}
		switch {
		case !left:
		case pk.Prev != (front.CellRect{}) && in(pk.Prev) && pk.Page > 0:
			p.pickerTo(pk, (pk.Page-1)*pk.PerPage)
		case pk.Next != (front.CellRect{}) && in(pk.Next) && pk.Page < pk.Pages-1:
			p.pickerTo(pk, (pk.Page+1)*pk.PerPage)
		default:
			return false
		}
		return true
	}
	sp := p.spots(v)
	if sp == nil || s.sub != scrMain {
		return false
	}
	for i, t := range sp {
		if in(t.r) {
			if moved || left {
				s.sel = i
			}
			if left {
				t.do()
			}
			return true
		}
	}
	if ed := v.IconEditor; ed != nil && left { // the poses: the head above, the weapon below
		if a, ok := iconClick(ed, cx, cy); ok {
			p.sendA(a)
			return true
		}
	}
	return false
}

// open goes to submenu sub: back on the main menu, the row that opened the
// submenu comes back; leaving the main menu keeps that row; a submenu
// starts on its first row.
func (s *screenUI) open(sub int) {
	switch {
	case sub == scrMain:
		s.sel = s.mainSel
	case s.sub == scrMain:
		s.mainSel, s.sel = s.sel, 0
	default:
		s.sel = 0
	}
	s.sub, s.top = sub, 0
}

// spot is a part of a special screen's page (the creation sheet, the icon
// editor) the focus goes to: its cells, its hotkey and what it does. The
// bar's words on the prompt row are spots too.
type spot struct {
	r   front.CellRect
	key byte
	do  func()
}

var statNames = [6]string{"STR", "INT", "WIS", "DEX", "CON", "CHA"}

// spots are the special screen's parts, in reading order, the bar's words
// last; nil on other screens.
func (p *Presenter) spots(v front.View) []spot {
	var sp []spot
	at := func(r front.CellRect, key byte, do func()) { sp = append(sp, spot{r, key, do}) }
	send := func(k front.ActKind, n int) func() { return func() { p.send(k, n) } }
	switch ed, cr := v.IconEditor, v.Creation; {
	case v.Picker != nil:
		return nil
	case ed != nil:
		click := func(x, y int) func() {
			return func() {
				if a, ok := iconClick(ed, x, y); ok {
					p.sendA(a)
				}
			}
		}
		if ed.Custom {
			at(ed.Remove, 'R', send(front.ActIconRemove, 0))
		} else {
			at(ed.Head, 'H', click(ed.Head.X1, ed.Head.Y1))
			at(ed.Weapon, 'W', click(ed.Weapon.X1, ed.Weapon.Y1))
			at(ed.Size, 'S', click(ed.Size.X1, ed.Size.Y1))
			for i := range iconParts {
				for sh := range 2 {
					x, y := ed.Parts.X1+3*sh, ed.Parts.Y1+i
					at(front.CellRect{X1: x, Y1: y, X2: x + 1, Y2: y}, 0, click(x, y))
				}
			}
			for j, c := range ed.Colours {
				if c >= 0 && ed.Palette != (front.CellRect{}) {
					x, y := ed.Palette.X1+2*(j%8), ed.Palette.Y1+j/8
					at(front.CellRect{X1: x, Y1: y, X2: x + 1, Y2: y}, 0, click(x, y))
				}
			}
		}
	case cr != nil:
		at(cr.Name, 0, func() { p.typeAt(cr.Name, "Name", 15, front.ActName, 0) })
		for f, fl := range cr.Fields {
			at(fl.At, 0, func() { p.openField(cr, f) })
		}
		for i, r := range cr.Stats {
			at(r, 0, func() { p.typeAt(r, cr.StatNames[i], 5, front.ActStat, i) })
		}
		at(cr.Head, 0, send(front.ActPicker, front.PickHead))
		at(cr.Body, 0, send(front.ActPicker, front.PickBody))
		at(cr.Icon, 'I', send(front.ActIconEdit, -1))
	default:
		return nil
	}
	row := p.scr.ph/8 - 1 // the prompt row
	for _, w := range barWords(v) {
		at(front.CellRect{X1: w.x, Y1: row, X2: w.x + len(w.label) - 1, Y2: row}, w.key, p.barDo(v, w.key))
	}
	return sp
}

// barWords are the special screen's bar on the prompt row, as the
// original's: the creation's Roll, Minimal, Keep, Exit; the icon
// editor's Keep, Exit.
func barWords(v front.View) []footWord {
	ws := []string{"Keep", "Exit"}
	if v.IconEditor == nil {
		ws = []string{"Roll", "Minimal", "Keep", "Exit"}
		if v.Creation != nil && len(v.Creation.Powers) > 0 {
			ws = []string{"Roll", "Minimal", "Powers", "Keep", "Exit"}
		}
	}
	var out []footWord
	x := 0
	for _, w := range ws {
		out = append(out, footWord{w, w[0], x})
		x += len(w) + 1
	}
	return out
}

// barDo is what the bar's word with hotkey h does.
func (p *Presenter) barDo(v front.View, h byte) func() {
	return func() {
		cr := v.Creation
		switch {
		case h == 'K' && cr != nil && slices.ContainsFunc(cr.Powers, func(f front.Field) bool { return f.Cur < 0 }), h == 'P' && cr != nil:
			p.scr.open(scrPowers) // a power still to pick: the picks first
		case h == 'K':
			p.send(front.ActKeep, 0)
		case h == 'E' && (v.IconEditor != nil || cr == nil):
			p.send(front.ActClose, 0)
		case h == 'E':
			p.askThen(cr.Abandon, func() { p.send(front.ActClose, 0) })
		case h == 'R':
			p.send(front.ActRoll, 0)
		case h == 'M':
			p.minimal(cr)
		}
	}
}

// spotKey takes key k on a special screen: the arrows move the focus to
// the nearest part that way, Tab to the next, Enter does the focused part,
// a hotkey its part, Esc is the bar's Exit.
func (p *Presenter) spotKey(k uint16, v front.View, sp []spot) bool {
	s := &p.scr
	dir := map[input.Act]image.Point{input.Up: {0, -1}, input.Down: {0, 1}, input.Left: {-1, 0}, input.Right: {1, 0}}
	switch a := actOf(k); {
	case dir[a] != image.Point{}:
		s.sel = nearest(sp, s.sel, dir[a])
	case a == input.Minimap: // Tab, pad Y
		s.sel = (s.sel + 1) % len(sp)
	case a == input.Enter || a == input.Space || a == input.PadA:
		sp[s.sel].do()
	case a == input.Esc:
		p.barDo(v, 'E')()
	case k < 128:
		h := byte(k) &^ 0x20
		for i, t := range sp {
			if t.key == h && h != 0 {
				s.sel = i
				t.do()
				return true
			}
		}
		return false
	default:
		return false
	}
	return true
}

// nearest is the spot from cur the direction d reaches first: ahead of
// it, the least distance with the sideways part counting double; cur when
// none is that way.
func nearest(sp []spot, cur int, d image.Point) int {
	mid := func(r front.CellRect) image.Point { return image.Pt(r.X1+r.X2, r.Y1+r.Y2) } // doubled
	c, best, bd := mid(sp[cur].r), cur, 1<<30
	for i, t := range sp {
		o := mid(t.r).Sub(c)
		ahead, side := o.X*d.X+o.Y*d.Y, o.X*d.Y-o.Y*d.X
		if ahead <= 0 {
			continue
		}
		if side < 0 {
			side = -side
		}
		if n := ahead + 2*side; n < bd {
			best, bd = i, n
		}
	}
	return best
}

// crField is creation field f: the four on the page, then the power picks.
func crField(cr *front.Creation, f int) front.Field {
	if f < len(cr.Fields) {
		return cr.Fields[f]
	}
	return cr.Powers[f-len(cr.Fields)]
}

// upFirst is s with its first letter upper case.
func upFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
