// options.go is the Esc (pause) menu: its root, the Gameplay and Graphics
// pages, the host's rows and pages, and the yes/no questions it asks. It
// reads the keys itself (pad.go) and draws in the game's text window
// (keyWindow). It came from blobber, which now only draws the 3D view.

package ui

import (
	"fmt"
	"math"
	"os"
	"slices"
	"strings"

	"gobox/gpu"
	"gobox/gpu/vector"
	"gobox/ui/blobber"
	"image/color"
)

// MenuOptions are the settings on the Esc menu's pages: the 3D view's and the
// window's. The settings file holds their fields at its top level.
type MenuOptions struct {
	blobber.Options
	Fullscreen bool
}

// Settings are the options in use, the one copy: the 3D view reads its own
// part (NewView); set them before it is made (from a settings file).
// OnSettings is called with the new options whenever the player changes one.
var (
	Settings   = MenuOptions{Options: blobber.Options{Smooth: true, Fluid: true, Dither: true, Torch: 1, Speed: 1}, Fullscreen: true}
	OnSettings func(MenuOptions)
)

// optionsMenu is the Esc menu's state.
type optionsMenu struct {
	open, quit      bool
	page            string    // the settings page shown, "" the root
	confirm         *question // the yes/no the menu asks, or nil
	sel             int
	pageSel, askSel int // the root's row that opened the page, the question: back restores it
	lastPoint       int // the line the mouse was last on (-1 none)
}

// question is a yes/no the menu asks before do.
type question struct {
	q  string
	do func()
}

// menuLine is one line of the Esc menu: change runs it (d: 0 Enter, -1 Left, 1 Right); value, when
// set, is shown after the label.
type menuLine struct {
	label  string
	value  func(m *optionsMenu) string
	change func(m *optionsMenu, d int)
}

// Row is a host row on an Esc menu page: Get is the value shown, Change runs on Enter (0), Left (-1)
// and Right (1). At is the built-in row it goes before (rows of equal At keep their order).
type Row struct {
	Label  string
	At     int
	Get    func() string
	Change func(d int)
}

// ExtraRows are the host's rows by page ("Gameplay", "Graphics"); any other key is a page of its own,
// listed on the root menu after Graphics (sorted by title), with every row's At ignored.
var ExtraRows map[string][]Row

// FontNames, FontAt, SetFont and FontGlyph are the host's text fonts (the
// Graphics page's Font line opens their list); MenuGlyph, while the list is
// drawn, is line's glyph of ch for Window.
var (
	FontNames = func() []string { return []string{"Default"} }
	FontAt    = func() int { return 0 }
	SetFont   = func(int) {}
	FontGlyph func(font int, ch byte) [8]byte
	MenuGlyph func(line int, ch byte) [8]byte
)

// OnSaveLoad, when set, adds "Save / Load" to the Esc menu (the host's save slots); OnMainMenu
// adds "Back to Main Menu" (asked first).
var OnSaveLoad, OnMainMenu func()

// QuitRow and QuitQuestion are the one quit row and question, on the Esc
// menu and por's main menu.
const QuitRow, QuitQuestion = "Quit to DOS", "Quit to DOS? Unsaved progress is lost."

// openAt opens the menu at its root.
func (m *optionsMenu) openAt() {
	m.open, m.page, m.confirm, m.sel, m.lastPoint = true, "", nil, 0, -1
}

// update runs the menu for one frame and reports true when Quit was confirmed.
func (m *optionsMenu) update() bool {
	items := m.menuOptions()
	n := len(items)
	for i, k := range menuKeys(items) { // a hotkey (a letter, else a digit) picks its line
		if m.page != "Font" && (k >= '1' && k <= '9' && KeyJustPressed(gpu.Key1+gpu.Key(k-'1')) || k >= 'A' && k <= 'Z' && KeyJustPressed(gpu.KeyA+gpu.Key(k-'A'))) {
			m.sel = i
			items[i].change(m, 0)
			return m.quit
		}
	}
	if KeyRepeated(gpu.KeyUp) {
		m.sel = (m.sel + n - 1) % n
	}
	if KeyRepeated(gpu.KeyDown) {
		m.sel = (m.sel + 1) % n
	}
	switch { // as the host's menus (remaster menuNav): PgUp/PgDn ten lines, Home/End the ends
	case KeyRepeated(gpu.KeyPageUp):
		m.sel = max(0, m.sel-10)
	case KeyRepeated(gpu.KeyPageDown):
		m.sel = min(n-1, m.sel+10)
	case KeyJustPressed(gpu.KeyHome):
		m.sel = 0
	case KeyJustPressed(gpu.KeyEnd):
		m.sel = n - 1
	}
	m.sel %= n
	if m.page == "Font" && m.sel != FontAt() { // the highlighted font is in use at once (the pointer's too)
		SetFont(m.sel)
	}
	d, ok := 0, KeyJustPressed(gpu.KeyEnter) || KeyJustPressed(gpu.KeySpace)
	if m.page != "" && m.confirm == nil { // Left and Right step a page's settings
		switch {
		case KeyRepeated(gpu.KeyLeft):
			d, ok = -1, true
		case KeyRepeated(gpu.KeyRight):
			d, ok = 1, true
		}
	}
	if !ok {
		return false
	}
	items[m.sel].change(m, d)
	m.changed()
	return m.quit
}

// changed tells the host of a page's change.
func (m *optionsMenu) changed() {
	if m.page == "" {
		return
	}
	Settings.Fullscreen = gpu.IsFullscreen()
	if OnSettings != nil {
		OnSettings(Settings)
	}
}

// back is Esc in the menu: a page or a question goes back to the root (true), the root closes (false).
func (m *optionsMenu) back() bool {
	if m.page == "" && m.confirm == nil {
		return false
	}
	if m.page == "Font" {
		m.page = "Graphics"
		m.sel = m.fontRow()
		return true
	}
	if m.confirm != nil { // back to the row the question or the page was opened from
		m.sel = m.askSel
	} else {
		m.sel = m.pageSel
	}
	m.page, m.confirm = "", nil
	return true
}

// withExtra is the built-in rows of page with the host's inserted, then exit.
func withExtra(page string, rows []menuLine, exit menuLine) []menuLine {
	var o []menuLine
	extra := ExtraRows[page]
	for i := 0; i <= len(rows); i++ {
		for _, r := range extra {
			if r.At == i {
				o = append(o, r.line())
			}
		}
		if i < len(rows) {
			o = append(o, rows[i])
		}
	}
	return append(o, exit)
}

func (r Row) line() menuLine {
	l := menuLine{label: r.Label, change: func(_ *optionsMenu, d int) {
		if r.Change != nil {
			r.Change(d)
		}
	}}
	if r.Get != nil {
		l.value = func(*optionsMenu) string { return r.Get() }
	}
	return l
}

func flip(p *bool) func(*optionsMenu, int) { return func(*optionsMenu, int) { *p = !*p } }

// settingStep moves v by s within lo..hi: Left down, Right up, Enter up and round to lo past hi.
func settingStep(v, lo, hi, s float64, d int) float64 {
	switch {
	case d < 0:
		v -= s
	case v+s > hi+s/2:
		if d == 0 {
			return lo
		}
	default:
		v += s
	}
	return math.Round(math.Max(lo, math.Min(hi, v))/s) * s
}

// settingBar draws v in lo..hi (step s) as a slider.
func settingBar(v, lo, hi, s float64) string {
	i, n := int(math.Round((v-lo)/s)), int(math.Round((hi-lo)/s))
	return strings.Repeat("#", i+1) + strings.Repeat("-", n-i)
}

func onOff(b bool) string {
	if b {
		return "ON"
	}
	return "OFF"
}

func onOffOf(p *bool) func(*optionsMenu) string {
	return func(*optionsMenu) string { return onOff(*p) }
}

// menuOptions are the lines of the menu page shown: the root, a settings page, or a yes/no question.
func (m *optionsMenu) menuOptions() []menuLine {
	if m.confirm != nil {
		do := m.confirm.do
		return []menuLine{{label: "Yes", change: func(m *optionsMenu, _ int) { m.back(); do() }},
			{label: "No", change: func(m *optionsMenu, _ int) { m.back() }}}
	}
	exit := menuLine{label: "Exit", change: func(m *optionsMenu, _ int) { m.back() }} // a settings page's way back, as Esc
	s := &Settings.Options
	switch m.page {
	case "Graphics":
		return withExtra("Graphics", []menuLine{
			{"Fullscreen", func(*optionsMenu) string { return onOff(gpu.IsFullscreen()) }, func(*optionsMenu, int) { gpu.SetFullscreen(!gpu.IsFullscreen()) }},
			{"Dithered darkness", onOffOf(&s.Dither), flip(&s.Dither)},
			{"Torch intensity", func(*optionsMenu) string { return fmt.Sprintf("%.1f", s.Torch) }, func(_ *optionsMenu, d int) { s.Torch = settingStep(s.Torch, 0.5, 1.5, 0.5, d) }},
			{"Font", func(*optionsMenu) string { return FontNames()[FontAt()] }, func(m *optionsMenu, _ int) { m.page, m.sel = "Font", FontAt() }},
		}, exit)
	case "Font": // every font, each line in its own; the highlighted one is in use at once
		var o []menuLine
		for _, n := range FontNames() {
			o = append(o, menuLine{label: n, change: func(m *optionsMenu, _ int) { m.page = "Graphics"; m.sel = m.fontRow() }})
		}
		return o
	case "Gameplay":
		return withExtra("Gameplay", []menuLine{
			{"Smooth movement", onOffOf(&s.Smooth), flip(&s.Smooth)},
			{"Fluid movement", onOffOf(&s.Fluid), flip(&s.Fluid)},
			{"No Grid", onOffOf(&s.NoGrid), flip(&s.NoGrid)},
			{"Walk speed", func(*optionsMenu) string {
				return settingBar(s.Speed, blobber.MinSpeed, blobber.MaxSpeed, 0.25) + fmt.Sprintf(" %.2fx", s.Speed)
			}, func(_ *optionsMenu, d int) {
				s.Speed = settingStep(s.Speed, blobber.MinSpeed, blobber.MaxSpeed, 0.25, d)
			}},
		}, exit)
	}
	if rows, ok := ExtraRows[m.page]; ok && m.page != "" {
		var o []menuLine
		for _, r := range rows {
			o = append(o, r.line())
		}
		return append(o, exit)
	}
	page := func(p string) menuLine {
		return menuLine{label: p, change: func(m *optionsMenu, _ int) { m.pageSel, m.page, m.sel = m.sel, p, 0 }}
	}
	ask := func(label, q string, do func()) menuLine {
		return menuLine{label: label, change: func(m *optionsMenu, _ int) { m.askSel, m.confirm, m.sel = m.sel, &question{q, do}, 1 }}
	}
	o := []menuLine{{label: "Resume", change: func(m *optionsMenu, _ int) { m.open = false }}}
	if OnSaveLoad != nil {
		o = append(o, menuLine{label: "Save / Load", change: func(m *optionsMenu, _ int) { m.open = false; OnSaveLoad() }})
	}
	o = append(o, page("Gameplay"), page("Graphics"))
	var pages []string
	for p := range ExtraRows {
		if p != "Gameplay" && p != "Graphics" {
			pages = append(pages, p)
		}
	}
	slices.Sort(pages)
	for _, p := range pages {
		o = append(o, page(p))
	}
	if OnMainMenu != nil {
		o = append(o, ask("Back to Main Menu", "Back to the main menu? Unsaved progress is lost.", func() { m.open = false; OnMainMenu() }))
	}
	return append(o, ask(QuitRow, QuitQuestion, func() { m.quit = true }))
}

// fontRow is the Font line's row on the Graphics page.
func (m *optionsMenu) fontRow() int {
	save := m.page
	m.page = "Graphics"
	defer func() { m.page = save }()
	for i, l := range m.menuOptions() {
		if l.label == "Font" {
			return i
		}
	}
	return 0
}

// menuKeys are the Esc menu lines' hotkeys (Hotkeys), the digits 1-9 on
// lines left without a letter (as digitKeys).
func menuKeys(items []menuLine) []byte {
	var labels []string
	for _, o := range items {
		labels = append(labels, o.label)
	}
	keys, d := Hotkeys(labels), byte('1')
	for i := range keys {
		if keys[i] == 0 && d <= '9' {
			keys[i], d = d, d+1
		}
	}
	return keys
}

// Hotkeys gives each label its first letter not taken by a label above (0: none left).
func Hotkeys(labels []string) []byte {
	var keys []byte
	taken := map[byte]bool{'E': slices.Contains(labels, "Exit")} // Exit is E on every screen
	for _, l := range labels {
		var k byte
		if l == "Exit" {
			keys = append(keys, 'E')
			continue
		}
		for _, c := range []byte(strings.ToUpper(l)) {
			if c >= 'A' && c <= 'Z' && !taken[c] {
				k, taken[c] = c, true
				break
			}
		}
		keys = append(keys, k)
	}
	return keys
}

// menuItems are the lines of the menu page shown.
func (m *optionsMenu) menuItems() []string {
	var out []string
	for _, o := range m.menuOptions() {
		s := o.label
		if o.value != nil {
			s = fmt.Sprintf("%-23s%s", o.label+":", o.value(m))
		}
		out = append(out, s)
	}
	return out
}

// optionsOpen reports whether the Esc menu is up (no 3D view, as a picture scene: no menu).
func (p *Presenter) optionsOpen() bool { return p.view != nil && p.opts.open }

// openOptions opens the Esc menu at its root; while it is open,
// UpdateOptions reads the keyboard and the gamepads for it.
func (p *Presenter) openOptions() { p.opts.openAt() }

// UpdateOptions runs the Esc menu for one frame: arrows and Enter pick and
// change, Esc goes back to the root or closes it. Quit exits the program.
func (p *Presenter) UpdateOptions() {
	defer clear(padKeys)
	m := &p.opts
	if KeyJustPressed(gpu.KeyEscape) {
		if !m.back() {
			m.open = false
		}
		return
	}
	if m.update() {
		os.Exit(0)
	}
}

// pointOptions is the mouse on the Esc menu: line (-1: none) is under the
// pointer and highlights; click picks it (Enter), back goes back (Esc),
// wheel steps a setting (Left / Right).
func (p *Presenter) pointOptions(line int, click, back bool, wheel int) {
	m := &p.opts
	if back {
		if !m.back() {
			m.open = false
		}
		return
	}
	items := m.menuOptions()
	if line < 0 || line >= len(items) {
		return
	}
	if line != m.lastPoint { // only a move highlights, so the keys keep theirs
		m.sel, m.lastPoint = line, line
	}
	d := 0
	switch {
	case click:
	case wheel != 0 && m.page != "" && m.confirm == nil:
		d = 1 // wheel up: more
		if wheel < 0 {
			d = -1
		}
	default:
		return
	}
	m.sel, m.lastPoint = line, -1 // a new page highlights under the pointer once it moves
	items[line].change(m, d)
	m.changed()
	if m.quit {
		os.Exit(0)
	}
}

// drawOptions draws the Esc menu, when it is open, over what is drawn, in the game's text window.
func (p *Presenter) drawOptions(dst *gpu.Image) {
	m := &p.opts
	if !m.open {
		return
	}
	items, title := m.menuItems(), ""
	switch {
	case m.confirm != nil:
		title = m.confirm.q
	case m.page != "":
		title = strings.ToUpper(m.page)
	}
	if m.confirm == nil { // the game dims under the menu; the window dims under a question itself
		b := dst.Bounds()
		vector.DrawFilledRect(dst, 0, 0, float32(b.Dx()), float32(b.Dy()), color.NRGBA{A: 0xA0})
	}
	keys := menuKeys(m.menuOptions())
	if m.page == "Font" && FontGlyph != nil {
		MenuGlyph = FontGlyph
		defer func() { MenuGlyph = nil }()
		keys = nil // the font names' letters would pick fonts at a keystroke
	}
	p.keyWindow(dst, title, items, keys, nil, m.sel, false)
}
