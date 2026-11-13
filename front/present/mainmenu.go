package present

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"slices"
	"strings"
	"time"

	"gobox/game"
	"gobox/game/play"
	"gobox/gpu"
	"gobox/gpu/vector"
	"gobox/ui"

	"gobox/blobber"

	"gobox/front"
)

// The main menu (the party menu, View.MainMenu): the title's lettering on
// black, the functions under it in a box wrapped round them, the party's
// cards where the 3D view has them.

// The title's picture (Data.TitleBackdrop) shows alone for titleHold when
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
	backdrop srcImage          // title on the GPU
	title    image.Image       // the title's picture kept under the menu (Data.TitleBackdrop; nil: none)
	painted  bool              // title is a painted picture of its own (a SourcePNG layer), not the intro's last: the menu fades in over it without the zoom
	cards    []image.Rectangle // the party's cards drawn
	rows     string            // the rows last drawn, and whether in the submenu: rows the game changed start the highlight on the first
	rowsSub  bool

	titleDrawn bool // this frame drew the main menu's screen: the title's picture (if any) under it (TitleDrawn)
}

// drawMain draws the main menu.
func (m *mainUI) drawMain(h MainHost, dst *gpu.Image, v front.MainMenu) {
	m.titleDrawn = true
	dst.Fill(color.Black)
	b := dst.Bounds()
	bd := m.title // the title's picture stays: the menu comes in over it
	if bd == nil {
		m.drawMainUI(h, dst, v)
		return
	}
	bb := bd.Bounds()
	k := min(float64(b.Dx())/float64(bb.Dx()), float64(b.Dy())/float64(bb.Dy())) // fit, keeping the aspect: the window has its shape (SetWindowAspect); the full screen gets black bars rather than a cropped or stretched picture
	op := &gpu.DrawImageOptions{}
	op.GeoM.Scale(k, k)
	op.GeoM.Translate((float64(b.Dx())-float64(bb.Dx())*k)/2, (float64(b.Dy())-float64(bb.Dy())*k)/2)
	dst.DrawImage(m.backdrop.of(bd), op)
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
	if z := 0.1 + 0.9*float64(max(a, 0)); !m.painted { // over the intro's last picture the menu zooms in from the centre out
		lo.GeoM.Translate(-float64(b.Dx())/2, -float64(b.Dy())/2)
		lo.GeoM.Scale(z, z)
		lo.GeoM.Translate(float64(b.Dx())/2, float64(b.Dy())/2)
	}
	lo.ColorScale.ScaleAlpha(max(a, 0))
	dst.DrawImage(m.layer, lo)
}

// mainFade is how far the menu has faded in over the title (0 not yet, 1
// fully); the first call starts the clock.
func (m *mainUI) fade() float32 {
	if m.at.IsZero() {
		m.at = now()
	}
	return float32(now().Sub(m.at)-titleHold) / float32(titleFade)
}

// mainSkip ends the title's hold and fade on a key or click while they
// run, so the menu shows at once; it reports whether it did. A click that
// did is used up (its target was not visible yet); a key but Enter
// and Space goes on to the menu (Key).
func (m *mainUI) skip(hall bool) bool {
	if hall || m.title == nil || m.fade() >= 1 { // the hall's menu, and a game without a title picture, have no title to wait for
		return false
	}
	m.at = now().Add(-titleHold - titleFade)
	return true
}

// mainNames are the party menu's functions as the menu names them.
var mainNames = map[front.ActKind]string{
	front.ActBegin: "BEGIN Adventuring", front.ActLoad: "Load Game", front.ActQuickStart: "Quick Start", front.ActQuit: blobber.QuitRow,
	front.ActAdd: "Add", front.ActCreate: "Create", front.ActForget: "Forget", front.ActDropMember: "Drop",
	front.ActModify: "Modify", front.ActTraining: "Train", front.ActNewClass: "New Class", front.ActMember: "View", front.ActRemove: "Remove",
}

// mainLong are the party menu's functions by their full names, the keys
// a series renames them by in Data.Texts.
var mainLong = map[front.ActKind]string{
	front.ActAdd: "Add Character", front.ActCreate: "Create Character", front.ActForget: "Forget Character", front.ActDropMember: "Drop Character",
	front.ActModify: "Modify Character", front.ActTraining: "Train Character", front.ActMember: "View Character", front.ActRemove: "Remove Character",
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
		name := mainNames[k]
		if t := Text(mainLong[k]); t != "" { // the series' name for it (Data.Texts: BR2's Remove is Purge, its Drop Remove)
			name, _, _ = strings.Cut(t, " ")
		}
		key := name[0]
		if k == front.ActQuit { // Q is Quick Start's
			key = 'D'
		}
		it = append(it, ui.MenuRow{Text: name, Key: key, Do: do})
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
			n := -1
			if e.Act == front.ActCommand {
				n = e.N
			}
			do := func() { h.Send(e.Act, n) }
			if e.Seq != "" {
				do = func() { playSeq = e.Seq }
			}
			it = append(it, ui.MenuRow{Text: e.Text, Key: e.Key, Do: do})
		}
	}
	return it
}

// newRows starts the highlight on the first row when the game put up new
// rows (not the submenu's own open or Exit, which set it): the old
// highlight means nothing on them.
func (m *mainUI) newRows(lines []string) {
	if r := strings.Join(lines, "\n"); r != m.rows {
		if m.rows != "" && m.sub == m.rowsSub {
			m.sel = 0
		}
		m.rows, m.rowsSub = r, m.sub
	}
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
	m.newRows(lines)
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

// ReadPNG decodes the PNG at path, nil when it is not one.
func ReadPNG(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	im, err := png.Decode(f)
	if err != nil {
		return nil
	}
	return im
}

// paintedLayers reports whether ls hold a painted picture of their own
// (a SourcePNG layer) rather than the intro's own picture.
func paintedLayers(ls []game.Layer) bool {
	for _, l := range ls {
		if l.Source == game.SourcePNG {
			return true
		}
	}
	return false
}

// layersImage is the title backdrop's layers (Data.TitleBackdrop) drawn
// into one picture of their box (nil: none, or a picture that does not
// decode).
func layersImage(ls []game.Layer) image.Image {
	pics := make([]front.Picture, len(ls))
	var box image.Rectangle
	for i, l := range ls {
		f := front.Picture{File: l.File, Block: l.Block, Frame: l.Frame, Name: l.Name}
		if l.Source == game.SourcePNG {
			f = front.Picture{} // no DAX reference: read below
		} else if l.Pixels != nil {
			f = front.PixelsOf(l.Picture)
		}
		f.At, box = l.At, box.Union(l.At)
		pics[i] = f
	}
	pics = indexedPictures(daxPictures(pics), egaInk)
	for i, l := range ls {
		if l.Source == game.SourcePNG { // a painted picture (POR's art/TITLE.PNG)
			pics[i].Img = ReadPNG(l.File)
		}
		if pics[i].Img == nil {
			return nil
		}
	}
	if box.Empty() && len(pics) == 1 { // one picture with no place: its own size
		box = pics[0].Img.Bounds().Sub(pics[0].Img.Bounds().Min)
		pics[0].At = box
	}
	if box.Empty() {
		return nil
	}
	m := image.NewRGBA(image.Rect(0, 0, box.Dx(), box.Dy()))
	drawPictures(m, pics, box.Min)
	return m
}

// mainMenu is the party menu the state puts up: its open functions, whether
// it is the training hall's, the title picture and the members (Open nil:
// not up; the hall's own screen, the creation, the slots or a picker cover it).
func (p *Presenter) mainMenu() front.MainMenu {
	if p.eng == nil || p.slots != nil || p.pickerOver() {
		return front.MainMenu{}
	}
	k := menuKey{p.uiKey(), p.ran, p.training, p.creating, p.slots != nil, p.pickerOver(), p.cardSelected()}
	return p.lay.menu.get(k, p.layoutMainMenu)
}

// layoutMainMenu is mainMenu from the state.
func (p *Presenter) layoutMainMenu() front.MainMenu {
	r, s := p.eng.r, p.state()
	mode := r.modeOf(s)
	if r.refusalAsk != nil {
		mode = game.ModeAsk
	}
	c := play.CommonOf(s)
	if p.training && trainingOf(s) != nil || p.creating && Drafting(mode, creationOf(s)) {
		return front.MainMenu{}
	}
	open, hall := PartyMenu(mode, trainingOf(s), r.asking(s), c.Where.Hall)
	if len(open) == 0 {
		return front.MainMenu{}
	}
	m := front.MainMenu{Open: open, Hall: hall, Party: len(c.Party), Selected: -1}
	if !hall && c.PowerPick == nil {
		m.Logo = Title(nil, p.eng.art())
	}
	if i := p.cardSelected(); i >= 0 && i < len(c.Party) {
		m.Selected, m.Name = i, c.Party[i].Name
	}
	return m
}

// mainMenuUp: the party menu is up.
func (p *Presenter) mainMenuUp() bool { return p.mainMenu().Open != nil }
