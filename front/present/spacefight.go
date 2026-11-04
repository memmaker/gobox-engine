package present

// The Buck Rogers ship fight (game.ScreenSpaceFight), laid out on the page's
// cells as the original draws it: BR1's ship page and BR2's
// pO099_27CF (the fight's mode), the message window pR653_118A / pR602_11F6, the bar
// pO0AF_10CD.

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"
	"time"

	"gobox/front"
	"gobox/game"
	"gobox/input"
)

// sfCells is the SHIPS picture table (BR1 DS 23EC, BR2 d253E: the same
// cells): entry k is SHIPS block 10+k at its cell; 1..15 are the blast
// frames (relative), 16 the weapons piece, 17..22 the own system icons,
// 23..28 the foe's.
var sfCells = [29]image.Point{
	{0, 11}, {0, 0}, {2, 0}, {1, 0}, {0, 2}, {0, 1}, {0, 2}, {1, 2}, {0, 3}, {0, 3}, {1, 3}, {0, 3}, {0, 4}, {2, 4}, {0, 3}, {1, 4},
	{26, 14}, {26, 1}, {26, 3}, {26, 5}, {26, 7}, {26, 9}, {26, 11}, {1, 12}, {1, 13}, {1, 14}, {1, 15}, {1, 16}, {1, 17},
}

// The bars' two pixel rows by system (a0E5E, a0E64), and the system labels (a11BF).
var (
	sfBarHi  = [6]int{15, 11, 12, 14, 9, 13}
	sfBarLo  = [6]int{7, 3, 4, 12, 1, 5}
	sfLabels = [6]string{"Hull", "Sens", "Ctrl", "Life", "Fuel", "Engn"}
)

// toneInk is each game.Tone's EGA colour.
var toneInk = [...]int{game.ToneText: 10, game.ToneOwn: 11, game.ToneBad: 12, game.TonePrompt: 13, game.ToneWarn: 14, game.ToneLoud: 15}

// sfPause is the overflow pause's text (BR1 pR653_1291, DS 0EDE).
const sfPause = "Press button or enter to continue."

const (
	sfBlinkMs    = 70 // the blink's half-step (1F4A / 1E24)
	sfBlinkSteps = 10
)

// sfRow is one row of the message window.
type sfRow struct {
	Tone game.Tone
	Text string
}

// sfWrap is s cut as 11F6 cuts it: longer than the window's 38 columns, at
// the last space before column 37 (the space dropped), else at 36.
func sfWrap(s string) []string {
	if len(s) <= 38 {
		return []string{s}
	}
	v := strings.LastIndexByte(s[:37], ' ') + 1 // 1-based, 0: none
	if v == 0 {
		v = 37
	}
	return append([]string{s[:v-1]}, sfWrap(s[v:])...)
}

// sfPages are the window's pages of 3 rows for log: a page ends where the
// original pauses (the window full with a row to come, a Clear message on a
// window that holds text). pad: the last message is the crew menu's
// question, put on the last row (O036 / O04D: blanks until row 21).
func sfPages(log []game.Message, pad bool) [][3]sfRow {
	pages := [][3]sfRow{{}}
	cur, holds := 0, false
	page := func() {
		if holds {
			pages = append(pages, [3]sfRow{})
		}
		cur, holds = 0, false
	}
	write := func(r sfRow) {
		if cur > 2 {
			page()
		}
		pages[len(pages)-1][cur] = r
		cur, holds = cur+1, true
	}
	for i, m := range log {
		if m.Clear {
			page()
		}
		if pad && i == len(log)-1 {
			if cur > 2 {
				page()
			}
			for cur < 2 {
				write(sfRow{})
			}
		}
		for _, s := range sfWrap(m.Text) {
			write(sfRow{m.Tone, s})
		}
	}
	return pages
}

// sfUI is what the presenter keeps of the fight's screen: the message
// window's page, the blink's start and the bar's highlight.
type sfUI struct {
	log      []game.Message
	page     int
	blinkSig string
	blinkAt  time.Time
	barSig   string
	sel      int // the bar's highlight (0-based)
}

// sfShown is the fight's screen as it shows now.
type sfShown struct {
	Pages [][3]sfRow
	Page  int
	More  bool // the window pauses: the pause text, a key shows the next page
	Blink int  // the blink's half-step 0..10 (odd: the alternate colour)
	Sel   int
}

func sfOf(v front.View) *game.SpaceFight {
	if s := v.Screen; s != nil && s.Kind == game.ScreenSpaceFight {
		return s.SpaceFight
	}
	return nil
}

// sfBar is the bar's item string as 10CD takes it: each item followed by a
// blank, '@' after the first row's last.
func sfBar(a *front.Ask) string {
	var b strings.Builder
	for i, it := range a.Items {
		b.WriteString(it.Text)
		if i == a.Break-1 {
			b.WriteByte('@')
		} else {
			b.WriteByte(' ')
		}
	}
	return b.String()
}

// sfState brings the presenter's state up to v's fight and returns what shows.
func (p *Presenter) sfState(v front.View, sf *game.SpaceFight) sfShown {
	u := &p.sf
	if len(sf.Log) < len(u.log) || fmt.Sprint(sf.Log[:len(u.log)]) != fmt.Sprint(u.log) { // not more of the same: a new window
		u.page = 0
	}
	u.log = append(u.log[:0], sf.Log...)
	pad := v.Ask != nil && v.Ask.Break > 0
	sh := sfShown{Pages: sfPages(sf.Log, pad)}
	u.page = min(u.page, len(sh.Pages)-1)
	sh.Page, sh.More = u.page, u.page < len(sh.Pages)-1
	if b := sf.Blink; b.System > 0 || b.Slot > 0 {
		if sig := fmt.Sprint(*sf); sig != u.blinkSig {
			u.blinkSig, u.blinkAt = sig, time.Now()
		}
		sh.Blink = min(int(time.Since(u.blinkAt).Milliseconds()/sfBlinkMs), sfBlinkSteps)
	} else {
		u.blinkSig, sh.Blink = "", sfBlinkSteps
	}
	if a := v.Ask; a != nil {
		sig := sfBar(a)
		if a.Kind == front.AskConfirm {
			sig = a.Prompt + "\x00yn"
		}
		if sig != u.barSig { // 00AB: a new item string starts on its first item (Yes/No on the default)
			u.barSig, u.sel = sig, 0
			if a.Kind == front.AskConfirm && !a.Yes {
				u.sel = 1
			}
		}
	}
	sh.Sel = u.sel
	return sh
}

// sfKey takes key k on the fight's screen: the window's pause, the bar
// (10CD: Left/Right move, Up/Down switch rows, Enter takes the highlight,
// a hotkey its item, Esc nothing), Yes/No (2137: Esc ignored) and the
// explicit pause. false: not the screen's key.
func (p *Presenter) sfKey(k uint16, v front.View) bool {
	sf := sfOf(v)
	if sf == nil {
		return false
	}
	sh := p.sfState(v, sf)
	if sh.More {
		p.sf.page++
		return true
	}
	a := v.Ask
	if a == nil {
		return false
	}
	act, u := actOf(k), &p.sf
	upper := byte(k)
	if k < 256 && upper >= 'a' && upper <= 'z' {
		upper -= 32
	}
	switch {
	case a.Kind == front.AskConfirm:
		switch {
		case act == input.Left || act == input.Right:
			u.sel = 1 - u.sel
		case act == input.Enter && u.sel == 0, k < 256 && upper == 'Y':
			p.send(front.ActYes, 0)
		case act == input.Enter, k < 256 && upper == 'N':
			p.send(front.ActNo, 0)
		}
		return true
	case a.Kind == front.AskMenu && len(a.Items) == 0, a.Kind == front.AskKey:
		p.send(front.ActOK, 0)
		return true
	case a.Kind == front.AskMenu:
		n := len(a.Items)
		switch act {
		case input.Left:
			u.sel = (u.sel + n - 1) % n
			return true
		case input.Right:
			u.sel = (u.sel + 1) % n
			return true
		case input.Up, input.Down: // 0ED3, 1-based
			if r := a.Break; r > 0 {
				s := u.sel + 1
				if s <= r {
					s = min(s+r, n)
				} else {
					s = min(s-r, r)
				}
				u.sel = s - 1
			}
			return true
		case input.Enter:
			p.send(front.ActChoose, u.sel)
			return true
		case input.Esc: // 0: the game asks again
			return true
		}
		if k < 256 {
			for i, it := range a.Items {
				if sfHotkey(it) == upper {
					p.send(front.ActChoose, i)
					return true
				}
			}
		}
		return k < 256 && upper >= 'A' && upper <= 'Z'
	}
	return false
}

// sfHotkey is the item's key: its own, else its first capital (10CD).
func sfHotkey(it front.MenuItem) byte {
	if k := it.Key; k != 0 {
		if k >= 'a' && k <= 'z' {
			k -= 32
		}
		return k
	}
	for i := 0; i < len(it.Text); i++ {
		if c := it.Text[i]; c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' {
			return c
		}
	}
	return 0
}

// put is displayString: s at (x, y) in EGA fg on bg.
func (c *cellPage) put(x, y, fg, bg int, s string) {
	for i := 0; i < len(s) && x+i < c.cols && y < c.rows; i++ {
		if x+i >= 0 && y >= 0 {
			c.cells[y][x+i] = cell{true, true, s[i], egaInk[fg], egaInk[bg]}
		}
	}
}

// fillCells clears cells x1..x2, y1..y2 to EGA colour bg.
func (c *cellPage) fillCells(x1, y1, x2, y2, bg int) {
	r := image.Rect(x1*8, y1*8, x2*8+8, y2*8+8)
	c.pics = append(c.pics, front.Picture{At: r, Fill: true, Colour: egaInk[bg]})
	c.unset(r)
}

// fillPx fills pixel rectangle r with EGA colour n.
func (c *cellPage) fillPx(r image.Rectangle, n int) {
	c.pics = append(c.pics, front.Picture{At: r, Fill: true, Colour: egaInk[n]})
}

// shipsPic draws table entry k's SHIPS picture at cell at (plus off).
func (c *cellPage) shipsPic(k int, off image.Point) {
	m := daxMasked("SHIPS", 10+k, 0, -1)
	if m == nil {
		return
	}
	at := sfCells[k].Add(off).Mul(8)
	c.pics = append(c.pics, front.Picture{File: "SHIPS", Block: 10 + k, Img: m, At: m.Bounds().Sub(m.Bounds().Min).Add(at)})
}

// gamePic is p as an image (its pixels in ink, else its DAX block).
func gamePic(p game.Picture, ink [16]color.RGBA) image.Image {
	if p.Pixels != nil {
		f := pix0(p)
		return pixImage(f, ink)
	}
	if p.File == "" {
		return nil
	}
	return daxMasked(p.File, p.Block, p.Frame, -1)
}

func pix0(p game.Picture) front.Picture {
	f := front.Picture{Pix: p.Pixels, PixW: p.W}
	if p.Clear >= 0 {
		m := uint8(p.Clear)
		f.Mask = &m
	}
	return f
}

// rjust is s right-justified n wide (1E1D, 1620).
func rjust(n, w int) string {
	s := strconv.Itoa(n)
	return strings.Repeat(" ", max(0, w-len(s))) + s
}

// spaceFight lays the fight's screen out on c.
func (c *cellPage) spaceFight(sf *game.SpaceFight, sh sfShown, ask *front.Ask) {
	alt := func(on bool, fg, other int) int {
		if on && sh.Blink%2 == 1 && sh.Blink < sfBlinkSteps {
			return other
		}
		return fg
	}
	// 2358 / 2461: the view at its cell, the foe in the picture window
	c.shipsPic(0, image.Point{})
	if m := gamePic(sf.Foe, c.ink); m != nil {
		c.pics = append(c.pics, front.Picture{File: sf.Foe.File, Block: sf.Foe.Block, Img: m, At: image.Rect(0, 0, 192, 88)})
	}
	if b := sf.Blast; b != nil && b.K >= 1 && b.K <= 15 {
		c.shipsPic(b.K, b.At)
	}
	// 22E2 / 23EB
	c.put(1, 1, 12, 0, sf.Status)
	c.put(1, 2, 12, 0, "Range: "+strconv.Itoa(max(sf.Range, 0)))
	for _, own := range []bool{false, true} {
		sys, k0 := sf.Enemy, 23
		if own {
			sys, k0 = sf.Own, 17
			for i := range 6 { // 2144 / 2252: the rows cleared, the ticks
				c.fillCells(29, 2*i+1, 38, 2*i+1, 8)
				c.fillCells(29, 2*i+2, 38, 2*i+2, 0)
				for k := range 10 {
					c.fillPx(image.Rect(232+8*k, (2*i+2)*8, 233+8*k, (2*i+2)*8+8), 8)
				}
			}
		} else if sf.SensorsOut {
			c.fillCells(1, 12, 24, 17, 8)
			c.put(4, 14, 12, 8, "Sensors Destroyed")
			continue
		}
		for i := range 6 {
			c.shipsPic(k0+i, image.Point{})
		}
		for i := range 6 { // 1E60 / 1F82
			now, mx := sys.Now[i], sys.Max[i]
			fg, other, bg, col, row := 10, 12, 0, 10, i+12
			if own {
				fg, bg, col, row = 11, 8, 34, 2*i+1
			}
			if now == 0 {
				fg, other = 12, 0
			}
			if own {
				c.put(29, row, fg, bg, sfLabels[i]+":")
			}
			b := sf.Blink
			c.put(col, row, alt(b.System == i+1 && b.Own == own, fg, other), bg, rjust(now, 5))
			n := 0
			if mx >= 100 {
				n = (now + 15) >> 4 * 80 / (mx >> 4)
			} else if mx != 0 {
				n = now * 80 / mx
			}
			fill := n
			n = min(max(n, 0), 79)
			bx, by := 120, 8*i+96
			if own {
				bx, by = 232, 16*i+16
			}
			if fill > 0 {
				c.fillPx(image.Rect(bx, by+3, bx+n, by+4), sfBarHi[i])
				c.fillPx(image.Rect(bx, by+4, bx+n, by+5), sfBarLo[i])
			}
			for x := n; x < 80; x++ {
				e := 0
				if x&7 == 0 {
					e = 8
				}
				c.fillPx(image.Rect(bx+x, by+3, bx+x+1, by+5), e)
			}
		}
	}
	// 23F4 / 2508: the compact weapons, then the piece
	c.fillCells(29, 14, 38, 17, 8)
	c.put(29, 14, 10, 8, "Weapons:")
	for i, w := range sf.Weapons[:min(3, len(sf.Weapons))] {
		if w.Mounts <= 0 {
			continue
		}
		fg := 10
		if w.Working != w.Mounts {
			fg = 14
		}
		fg = alt(sf.Blink.Slot == i+1, fg, 12)
		name := w.Short
		if len(name) > 8 {
			name = name[:8]
		}
		c.put(29, 15+i, fg, 8, name)
		c.put(37, 15+i, fg, 8, rjust(w.Working, 2))
	}
	c.shipsPic(16, image.Point{})
	// the message window (1,19)-(38,21)
	c.fillCells(1, 19, 38, 21, 0)
	for i, r := range sh.Pages[sh.Page] {
		c.put(1, 19+i, toneInk[r.Tone], 0, r.Text)
	}
	// the prompt rows
	c.fillCells(0, 23, 39, 24, 0)
	switch {
	case sh.More, ask != nil && (ask.Kind == front.AskMenu && len(ask.Items) == 0 || ask.Kind == front.AskKey):
		c.put(0, 24, 15, 0, sfPause)
	case ask != nil && ask.Kind == front.AskConfirm:
		c.bar(ask.Prompt, "Yes No", sh.Sel)
	case ask != nil && ask.Kind == front.AskMenu:
		c.bar(ask.Prompt, sfBar(ask), sh.Sel)
	}
}

// bar draws 10CD's bar: the prompt fg13 at column 0, then the items, their
// capitals fg15, the rest fg10, the highlighted item 0 on 15; '@' starts
// the second row (the bar then on rows 23-24).
func (c *cellPage) bar(prompt, items string, sel int) {
	row := 24
	if strings.Contains(items, "@") {
		row = 23
	}
	c.put(0, row, 13, 0, prompt)
	isKey := func(b byte) bool { return b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' }
	var start, end []int // 0122: each item from its capital to two before the next (the last to the end)
	for i := 0; i < len(items); i++ {
		if isKey(items[i]) {
			if len(start) > 0 {
				end = append(end, i-2)
			}
			start = append(start, i)
		}
	}
	end = append(end, len(items)-1)
	x := len(prompt)
	for i := 0; i < len(items); i++ {
		ch := items[i]
		if ch == '@' {
			row, x = row+1, len(prompt)
			continue
		}
		switch {
		case sel < len(start) && i >= start[sel] && i <= end[sel]:
			c.put(x, row, 0, 15, string(ch))
		case isKey(ch):
			c.put(x, row, 15, 0, string(ch))
		default:
			c.put(x, row, 10, 0, string(ch))
		}
		x++
	}
}
