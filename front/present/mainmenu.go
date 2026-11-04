package present

import (
	"image"
	"image/color"
	"slices"
	"time"

	"gobox/gpu"
	"gobox/gpu/vector"
	"gobox/ui"

	"gobox/blobber"

	"gobox/front"
)

// The main menu (the party menu, View.MainMenu): the title's lettering on
// black, the functions under it in a box wrapped round them, the party's
// cards where the 3D view has them.

// The Amiga painted title (View.Backdrop) shows alone for titleHold when
// the main menu first comes up, then the menu fades in over it.
const titleHold, titleFade = 2 * time.Second, 1500 * time.Millisecond

// titleDim is how far the picture darkens under the menu (0 not, 1 black).
const titleDim = 0.6

// mainUI is the main menu's own state.
type mainUI struct {
	at       time.Time         // when the menu first showed the backdrop
	sel      int               // the highlight
	back     int               // the row that opened the Party submenu: Exit goes back to it
	sub      bool              // the Party submenu (the character functions) shows
	ask      int               // the question up (askQuit ..)
	layer    *gpu.Image        // the menu drawn for its fade in
	logo     srcImage          // MainMenu.Logo on the GPU
	backdrop srcImage          // MainMenu.Backdrop on the GPU
	cards    []image.Rectangle // the party's cards drawn
}

// drawMain draws the main menu.
func (m *mainUI) drawMain(h MainHost, dst *gpu.Image, v front.MainMenu) {
	dst.Fill(color.Black)
	b := dst.Bounds()
	if v.Backdrop == nil {
		m.drawMainUI(h, dst, v)
		return
	}
	bb := v.Backdrop.Bounds()
	k := min(float64(b.Dx())/float64(bb.Dx()), float64(b.Dy())/float64(bb.Dy())) // fit, keeping the aspect: the window has its shape (SetWindowAspect); the full screen gets black bars rather than a cropped or stretched picture
	op := &gpu.DrawImageOptions{}
	op.GeoM.Scale(k, k)
	op.GeoM.Translate((float64(b.Dx())-float64(bb.Dx())*k)/2, (float64(b.Dy())-float64(bb.Dy())*k)/2)
	dst.DrawImage(m.backdrop.of(v.Backdrop), op)
	a := m.fade()
	dim := uint8(titleDim * 255 * min(max(a, 0), 1)) // the picture darkens as the menu comes, so the lettering reads
	vector.DrawFilledRect(dst, 0, 0, float32(b.Dx()), float32(b.Dy()), color.NRGBA{A: dim})
	if a >= 1 {
		m.drawMainUI(h, dst, v)
		return
	}
	gpu.Sized(&m.layer, b.Dx(), b.Dy())
	m.layer.Clear()
	m.drawMainUI(h, m.layer, v)
	lo := &gpu.DrawImageOptions{}
	lo.ColorScale.ScaleAlpha(max(a, 0))
	dst.DrawImage(m.layer, lo)
}

// mainFade is how far the menu has faded in over the title (0 not yet, 1
// fully); the first call starts the clock.
func (m *mainUI) fade() float32 {
	if m.at.IsZero() {
		m.at = time.Now()
	}
	return float32(time.Since(m.at)-titleHold) / float32(titleFade)
}

// mainSkip ends the title's hold and fade on a key or click while they
// run, so the menu shows at once; it reports whether it did, and that
// input goes no further: the menu it would have acted on was not yet seen.
func (m *mainUI) skip(hall bool) bool {
	if hall || m.fade() >= 1 { // the hall's menu has no title to wait for
		return false
	}
	m.at = time.Now().Add(-titleHold - titleFade)
	return true
}

// mainNames are the party menu's functions as the menu names them.
var mainNames = map[front.ActKind]string{
	front.ActBegin: "BEGIN Adventuring", front.ActLoad: "Load Game", front.ActQuickStart: "Quick Start", front.ActQuit: blobber.QuitRow,
	front.ActAdd: "Add", front.ActCreate: "Create", front.ActForget: "Forget", front.ActDropMember: "Drop",
	front.ActModify: "Modify", front.ActTraining: "Train", front.ActNewClass: "New Class", front.ActMember: "View", front.ActRemove: "Remove",
}

// Main menu questions (mainUI.ask).
const (
	askNone = iota
	askQuit
	askDrop
	askSure   // the drop's second question
	askRemove // back to the saved characters?
)

// mainItems are the main menu's lines (the Party submenu with the
// character functions) or, with a question up, its answers and the
// question.
func (m *mainUI) items(h MainHost, v front.MainMenu) (string, []ui.MenuRow) {
	_, sel := h.PopUp()
	ask := func(q int) func() { return func() { m.ask, *sel = q, 1 } }
	no := ui.MenuRow{Text: "No", Key: 'N', Do: ask(askNone)}
	switch m.ask {
	case askQuit:
		return blobber.QuitQuestion, []ui.MenuRow{{Text: "Yes", Key: 'Y', Do: func() { m.ask = askNone; h.Send(front.ActQuit, 0) }}, no}
	case askDrop:
		return "Drop " + v.Name + " forever?", []ui.MenuRow{{Text: "Yes", Key: 'Y', Do: ask(askSure)}, no}
	case askRemove:
		return "Remove " + v.Name + " from the party?", []ui.MenuRow{{Text: "Yes", Key: 'Y', Do: func() { m.ask = askNone; h.Send(front.ActRemove, -1) }}, no}
	case askSure:
		return "Are you sure?", []ui.MenuRow{{Text: "Yes", Key: 'Y', Do: func() { m.ask = askNone; h.Send(front.ActDropMember, -1) }}, no}
	}
	open := func(k front.ActKind) bool { return slices.Contains(v.Open, k) }
	var it []ui.MenuRow
	add := func(k front.ActKind) {
		if !open(k) {
			return
		}
		do := func() { h.Send(k, -1) }
		switch k {
		case front.ActQuit:
			do = ask(askQuit)
		case front.ActDropMember:
			do = ask(askDrop)
		case front.ActRemove:
			do = ask(askRemove)
		}
		key := mainNames[k][0]
		if k == front.ActQuit { // Q is Quick Start's
			key = 'D'
		}
		it = append(it, ui.MenuRow{Text: mainNames[k], Key: key, Do: do})
	}
	if v.Hall { // the training hall's: the party's functions and the way out, no title, no submenu
		for _, k := range []front.ActKind{front.ActTraining, front.ActNewClass, front.ActMember, front.ActModify, front.ActAdd, front.ActCreate, front.ActRemove, front.ActDropMember} {
			add(k)
		}
		return "", append(it, ui.MenuRow{Text: "Leave the Hall", Key: 'L', Do: func() { h.Send(front.ActBegin, -1) }})
	}
	if m.sub {
		for _, k := range []front.ActKind{front.ActAdd, front.ActCreate, front.ActForget, front.ActDropMember, front.ActModify, front.ActTraining, front.ActMember, front.ActRemove} {
			add(k)
		}
		it = m.entries(h, it, true)
		return "", append(it, ui.MenuRow{Text: "Exit", Key: 'E', Do: func() { m.sub, m.sel = false, m.back }})
	}
	add(front.ActBegin)
	add(front.ActLoad) // Load Game leads when there are saved games, Quick Start without them
	if !open(front.ActLoad) {
		add(front.ActQuickStart)
	}
	it = append(it, ui.MenuRow{Text: "Party", Key: 'P', Do: func() { m.back, m.sub, m.sel = m.sel, true, 0 }})
	if open(front.ActLoad) {
		add(front.ActQuickStart)
	}
	add(front.ActQuit)
	return "", m.entries(h, it, false)
}

// entries is it with the game's own entries (Data.Entries) of the Party
// submenu (party) or the main list after it.
func (m *mainUI) entries(h MainHost, it []ui.MenuRow, party bool) []ui.MenuRow {
	for _, e := range menuEntries {
		if e.Party == party {
			it = append(it, ui.MenuRow{Text: e.Text, Key: e.Key, Do: func() { h.Send(e.Act, -1) }})
		}
	}
	return it
}

// drawMainUI draws the main menu's lettering, functions and cards.
func (m *mainUI) drawMainUI(h MainHost, dst *gpu.Image, v front.MainMenu) {
	up, sel := h.PopUp()
	b := dst.Bounds()
	mk := scale(dst)
	asking := m.ask
	m.ask = askNone
	_, items := m.items(h, v)
	m.ask = asking
	lines, keys := make([]string, len(items)), make([]byte, len(items))
	for i, it := range items {
		lines[i], keys[i] = it.Text, it.Key
	}
	menuH := float64(len(lines)+2) * 8 * mk
	var lh, lk int
	logo := logoImage(v.Logo)
	if logo != nil {
		lb := logo.Bounds()
		lk = max(1, min(b.Dx()*3/4/lb.Dx(), (b.Dy()-int(menuH))*2/3/lb.Dy()))
		lh = lb.Dy() * lk
	}
	gap := int(16 * mk)
	top := (b.Dy() - lh - gap - int(menuH)) / 2
	if logo != nil {
		op := &gpu.DrawImageOptions{}
		op.GeoM.Scale(float64(lk), float64(lk))
		op.GeoM.Translate(float64((b.Dx()-logo.Bounds().Dx()*lk)/2), float64(top))
		dst.DrawImage(m.logo.of(logo), op)
	}
	m.sel = min(max(m.sel, 0), len(lines)-1)
	h.DrawMenu(dst, lines, keys, m.sel, mk, b.Dx()/2, top+lh+gap)
	m.cards = h.DrawCards(dst, v.Hall) // the hall: the way to the next level under every portrait
	if m.ask != askNone {              // a question pops up over it
		q, items := m.items(h, v)
		var lines []string
		var keys []byte
		for _, it := range items {
			lines, keys = append(lines, it.Text), append(keys, it.Key)
		}
		h.KeyWindow(dst, q, lines, keys, *sel)
	}
	*up = m.ask != askNone
}

// pick picks line i of the main menu or its question.
func (m *mainUI) pick(h MainHost, v front.MainMenu, i int) {
	if _, items := m.items(h, v); i >= 0 && i < len(items) {
		if up, _ := h.PopUp(); !*up {
			m.sel = i
		}
		items[i].Do()
	}
}

// click takes a click on the main menu: a function or a portrait.
func (m *mainUI) click(h MainHost, v front.MainMenu, x, y int) bool {
	if up, _ := h.PopUp(); *up {
		i := slices.IndexFunc(h.WinRects(), image.Pt(x, y).In)
		if i < 0 { // a click outside: back (No)
			_, items := m.items(h, v)
			i = len(items) - 1
		}
		m.pick(h, v, i)
		return true
	}
	for i, r := range m.cards { // the party menu has no inventory: a left click selects (a right one: the sheet, Mouse)
		if image.Pt(x, y).In(r) {
			h.Send(front.ActLead, i)
			return true
		}
	}
	for i, r := range h.MenuRects() {
		if image.Pt(x, y).In(r) {
			m.pick(h, v, i)
			return true
		}
	}
	return false
}

// hover highlights the function under the pointer when it moves.
func (m *mainUI) hover(h MainHost, x, y int) {
	if !h.Point(x, y) {
		return
	}
	up, sel := h.PopUp()
	rects := h.MenuRects()
	if *up {
		rects = h.WinRects()
	}
	for i, r := range rects {
		if image.Pt(x, y).In(r) {
			if *up {
				*sel = i
			} else {
				m.sel = i
			}
		}
	}
}

// key moves the main menu's highlight (Up, Down), picks it (Enter, a
// hotkey); Esc answers No or goes Back; the member keys select a member.
func (m *mainUI) key(h MainHost, k uint16, v front.MainMenu) bool {
	_, items := m.items(h, v)
	up, psel := h.PopUp()
	sel, n := &m.sel, len(items)
	if *up {
		sel = psel
	}
	if n == 0 {
		return false
	}
	if mb, ok := memberKey(k, v.Selected, v.Party, true); ok && !*up { // not under a question about the member
		h.Send(front.ActLead, mb)
		return true
	}
	i, back, ok := ui.MenuNav(actOf(k), k, sel, ui.RowKeys(items))
	switch {
	case back && (*up || m.sub || v.Hall):
		m.pick(h, v, n-1) // No, Back, Leave the Hall
	case back:
		return false
	case i >= 0:
		m.pick(h, v, i)
	}
	return ok
}
