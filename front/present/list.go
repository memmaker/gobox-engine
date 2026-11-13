package present

import (
	"gobox/front"
	"image"
	"math"
	"slices"
	"strings"

	"gobox/gpu"
	"gobox/ui"

	"gobox/input"
)

// Lists (View.Ask (AskList)): the presenter scrolls and highlights the lines; the
// question's options follow them. A line picked is ActChoose -1, an option
// ActChoose N, each with the line highlighted as X.

const listRows = 12 // the lines shown at a time

type listUI struct {
	sel  int         // the row highlighted: a line, then the options
	top  int         // the first line shown
	opts []int       // the options shown (Menu.Items indices), as last laid out
	at   image.Point // the pointer, as last seen: only a moved one takes the highlight
	of   string      // the list's prompt: the same list asked again (after an act on a line) keeps its highlight
}

// listOpts are the options offered with line i: those it does not leave out.
func listOpts(v front.View, i int) []int {
	var out []int
outer:
	for n, it := range askMenu(v).Commands {
		if i >= 0 && i < len(askList(v)) {
			for _, w := range askList(v)[i].Without {
				if w == it.Text {
					continue outer
				}
			}
		}
		out = append(out, n)
	}
	return out
}

// listLine is the line the highlight is on or was last on.
func (p *Presenter) listLine(v front.View) int {
	return min(p.list.sel, len(askList(v))-1)
}

func (p *Presenter) drawList(dst *gpu.Image, glyph func(byte) [8]byte, v front.View, k float64, cx, top int) {
	l := &p.list
	n := len(askList(v))
	l.opts = listOpts(v, p.listLine(v))
	l.sel = min(max(l.sel, 0), n+len(l.opts)-1)
	if l.sel < n {
		l.top = min(max(l.top, l.sel-listRows+1), l.sel)
	}
	l.top = max(0, min(l.top, n-listRows))
	var lines []string
	var off []bool
	keys := slices.Clone(p.listKeys(v)[l.top:min(n, l.top+listRows)])
	for _, ln := range askList(v)[l.top:min(n, l.top+listRows)] {
		lines, off = append(lines, ln.Text), append(off, ln.Off)
	}
	w, all := 0, p.listKeys(v) // the blank row as wide as the widest line of the whole list: the box keeps one width while it scrolls
	for i, ln := range askList(v) {
		t, _ := ui.HotText(ln.Text, all[i])
		w = max(w, ui.VisLen(t))
	}
	lines, keys = append(lines, strings.Repeat(" ", w)), append(keys, 0)
	for _, o := range l.opts {
		lines, keys = append(lines, askMenu(v).Commands[o].Text), append(keys, askMenu(v).Commands[o].Key)
	}
	p.drawMenu(dst, glyph, askMenu(v).Prompt, lines, keys, off, p.listRow(n), k, cx, top)
	var sel *int
	if l.sel < n {
		sel = &l.sel
	}
	p.menuBar(dst, p.menuRects, n, min(n, listRows), &l.top, sel, k)
	if i := p.listLine(v); i >= 0 && askList(v)[i].Spell != 0 && len(p.menuRects) > 0 {
		p.drawSpellTip(dst, askList(v)[i].Spell, k, cx)
	}
}

// listKeys are the hotkeys of the list's lines then its options: the
// options' letters, digits on the lines that can be picked; none at all
// when more than nine lines can be (the first nine only would be uneven).
func (p *Presenter) listKeys(v front.View) []byte {
	n := len(askList(v))
	keys := make([]byte, n, n+len(p.list.opts))
	for _, o := range p.list.opts {
		keys = append(keys, askMenu(v).Commands[o].Key)
	}
	on := 0
	for _, l := range askList(v) {
		if !l.Off {
			on++
		}
	}
	if on > 9 {
		return keys
	}
	return ui.DigitKeys(keys, func(i int) bool { return i < n && !askList(v)[i].Off })
}

// listRow is the drawn row of the highlight.
func (p *Presenter) listRow(n int) int {
	if p.list.sel < n {
		return p.list.sel - p.list.top
	}
	return min(n, listRows) + 1 + p.list.sel - n
}

// listPick takes the row sel: a line or an option.
func (p *Presenter) listPick(v front.View, sel int) {
	n := len(askList(v))
	if modifying(v) { // a line, or Raise and Lower: that line up or down; the game's options as a list's
		switch o := sel - n; {
		case sel < n && !askList(v)[sel].Off:
			p.raise(sel, 1)
			return
		case o >= 0 && o < len(p.list.opts) && p.list.opts[o] >= len(v.Ask.Commands):
			p.raise(p.listLine(v), 1-2*(p.list.opts[o]-len(v.Ask.Commands)))
			return
		}
	}
	switch {
	case sel < n && !askList(v)[sel].Off:
		p.inv.acts = append(p.inv.acts, front.Action{Kind: front.ActChoose, N: -1, X: sel})
	case sel >= n && sel-n < len(p.list.opts):
		p.inv.acts = append(p.inv.acts, front.Action{Kind: front.ActChoose, N: p.list.opts[sel-n], X: p.listLine(v)})
	default:
		return
	}
	p.asked = ""
}

// raise sends line i's value up (d +1) or down (-1) on AskModify.
func (p *Presenter) raise(i, d int) {
	if i >= 0 {
		p.inv.acts = append(p.inv.acts, front.Action{Kind: front.ActRaise, N: i, X: d})
	}
}

// listKey takes key k on a list: the standard menu's keys (menuNav) over
// the lines and the options; a move passes over the lines left out (Off).
func (p *Presenter) listKey(k uint16, v front.View) bool {
	l := &p.list
	n := len(askList(v))
	if _, ok := memberKey(k, 0, len(p.party()), false); ok { // the member the list is for (keyAct)
		return false
	}
	if a := actOf(k); modifying(v) && (a == input.Left || a == input.Right) && l.sel < n { // the line highlighted down or up
		p.raise(l.sel, map[input.Act]int{input.Left: -1, input.Right: 1}[a])
		return true
	}
	keys := p.listKeys(v)
	i, back, ok := ui.MenuNav(actOf(k), k, &l.sel, keys)
	d := 1
	if a := actOf(k); a == input.Up || a == input.PgUp || a == input.End {
		d = -1
	}
	for range len(keys) {
		if l.sel >= n || !askList(v)[l.sel].Off {
			break
		}
		l.sel = (l.sel + d + len(keys)) % len(keys)
	}
	switch {
	case back:
		p.send(front.ActClose, 0)
	case i >= 0:
		p.listPick(v, i)
	}
	return ok
}

// listAt is the row (a line, then the options) drawn as menu row r, -1 the
// blank between them.
func (p *Presenter) listAt(v front.View, r int) int {
	n := len(askList(v))
	shown := min(n, listRows)
	switch {
	case r < shown:
		return p.list.top + r
	case r > shown:
		return n + r - shown - 1
	}
	return -1
}

// listClick takes a click on a drawn row; outside the popup it is back.
func (p *Presenter) listClick(x, y int, v front.View) {
	r := slices.IndexFunc(p.menuRects, image.Pt(x, y).In)
	switch {
	case r < 0:
		p.send(front.ActClose, 0)
	case p.listAt(v, r) >= 0:
		p.listPick(v, p.listAt(v, r))
	}
}

// listHover highlights the row under (x, y) when the pointer moved.
func (p *Presenter) listHover(x, y int, v front.View) {
	if image.Pt(x, y) == p.list.at {
		return
	}
	p.list.at = image.Pt(x, y)
	if r := slices.IndexFunc(p.menuRects, image.Pt(x, y).In); r >= 0 {
		if i := p.listAt(v, r); i >= 0 && (i >= len(askList(v)) || !askList(v)[i].Off) {
			p.list.sel = i
		}
	}
}

// drawPageMenu pops a page's question (its options or its list) up over
// the bottom of the page.
func (p *Presenter) drawPageMenu(dst *gpu.Image, note Note, v front.View) {
	if sfOf(v) != nil { // the ship fight's bar is on its page (spacefight.go)
		return
	}
	if p.pageScreen(v) {
		p.pageMenu = true
		p.drawScreen(dst, v)
		if t := typing(v); t != nil && t.Prompt != "" { // the game's typed line (the name after creation's Keep): over the screen, never under it
			p.drawTyping(dst, note, v, math.Max(1, math.Floor(scale(dst))), dst.Bounds().Dx()/2, 0)
		}
		return
	}
	if p.fullMap != nil { // no bar: the cursor cell's context menu, once opened
		p.pageMenu, p.winRects = true, p.winRects[:0]
		if p.fullMap.menu && p.fullMap.label == nil {
			p.drawMapMenu(dst, p.mapItems())
		}
		return
	}
	if t := typing(v); t != nil && t.Prompt == "" && !p.pageMenu { // a line whose question is the text box's own: the line alone, never invisible (a prompted one is drawn by the screen's own path)
		p.drawTyping(dst, note, v, math.Max(1, math.Floor(scale(dst))), dst.Bounds().Dx()/2, 0)
	}
	m := askMenu(v)
	if m == nil || typing(v) != nil || choosing(v) {
		return
	}
	if askList(v) != nil {
		b := dst.Bounds()
		k := math.Max(1, math.Floor(scale(dst)))
		p.asking(v, func() {
			p.drawList(dst, note.Glyph, v, k, b.Dx()/2, b.Dy()/2-int(k*float64(min(len(askList(v)), listRows)+len(m.Commands)+6)*4))
		})
		return
	}
	if t := strings.TrimSpace(m.Prompt); len(m.Commands) == 0 && t != "" { // a message with no choice ("You found some loot!"): the prompt alone until a key, never invisible
		b := dst.Bounds()
		k := math.Max(1, math.Floor(scale(dst)))
		p.drawMenu(dst, note.Glyph, "", []string{t}, nil, nil, -1, k, b.Dx()/2, b.Dy()/2-int(k*12))
		return
	}
	if len(m.Commands) < 2 { // "press Enter": nothing to pick
		return
	}
	if id := menuID(m); id != p.pageAsked { // its own: the page clears p.asked every frame, which put the highlight back on the default after every arrow
		p.pageAsked, p.sel = id, askDefault(v)
	}
	p.pageMenu = true
	rows := itemRows(m)
	lines, keys := make([]string, len(rows)), ui.RowKeys(rows)
	for i, r := range rows {
		lines[i] = r.Text
	}
	p.sel = min(max(p.sel, 0), len(lines)-1)
	p.asking(v, func() { p.keyWindow(dst, m.Prompt, lines, keys, nil, p.sel, true) })
}

// drawMapMenu draws m as the full map's cursor cell's context menu,
// beside the cell (winRects its lines).
func (p *Presenter) drawMapMenu(dst *gpu.Image, m *front.Menu) {
	// p.sel starts at Walk where the menu opens (pageMenuKey, Mouse); not here: the page clears p.asked every frame, which undid every hover
	rows := itemRows(m)
	lines, keys := make([]string, len(rows)), ui.RowKeys(rows)
	w := 0
	for i, r := range rows {
		lines[i] = r.Text
		w = max(w, len(r.Text)+4)
	}
	p.sel = min(max(p.sel, 0), len(lines)-1)
	fm, s, b := p.fm(), &p.scr, dst.Bounds()
	kx, ky := s.at.W/float64(s.pw), s.at.H/float64(s.ph)
	c := fm.At.Add(fm.Cursor.Add(image.Pt(1, 0)).Mul(fm.Cell)) // the cell's top right, on the page
	k := math.Max(1, math.Floor(scale(dst)))
	bw, bh := int(k*float64(w*8)), int(k*float64((len(lines)+2)*8))
	x, top := int(s.at.X+float64(c.X)*kx), int(s.at.Y+float64(c.Y)*ky)
	x, top = min(x, b.Dx()-bw), min(top, b.Dy()-bh)
	rects := slices.Clone(p.menuRects) // not the game's menu: its clicks stay
	p.drawMenu(dst, p.glyph, "", lines, keys, nil, p.sel, k, x+bw/2, top)
	p.winRects, p.menuRects = slices.Clone(p.menuRects), rects
}

// pageMenuKey takes key k on a page's question: the standard menu's keys
// (menuNav).
func (p *Presenter) pageMenuKey(k uint16, v front.View) bool {
	if p.pageScreen(v) {
		return p.screenKey(k, v)
	}
	if p.sheetUp(v) {
		return p.sheetKey(k)
	}
	if askHeld(v, k) {
		return true
	}
	switch i, back, ok := ui.MenuNav(actOf(k), k, &p.sel, ui.RowKeys(itemRows(askMenu(v)))); {
	case back:
		return false // ActClose (keyAct)
	case i >= 0:
		p.send(front.ActChoose, i)
	default:
		return ok
	}
	return true
}

// itemRows are an engine menu's items as menu rows (picking is ActChoose).
func itemRows(m *front.Menu) []ui.MenuRow {
	if m == nil { // the question went since the last frame (a key between them)
		return nil
	}
	rows := make([]ui.MenuRow, len(m.Commands))
	for i, it := range m.Commands {
		rows[i] = ui.MenuRow{Text: it.Text, Key: it.Key, Do: func() {}}
	}
	return rows
}

// openList: a list opens at its first line; the same list asked again (after
// an act on a line: Ready, Drop) stays on the line acted on.
func (p *Presenter) openList(v front.View) {
	if askList(v) != nil && askList(p.shownView) == nil && v.Ask.Prompt != p.list.of {
		p.list = listUI{sel: max(0, slices.IndexFunc(askList(v), func(l front.Entry) bool { return !l.Off })), of: v.Ask.Prompt}
	}
}
