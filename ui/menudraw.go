package ui

import (
	"image"
	"image/color"
	"slices"
	"strings"

	"gobox/gpu"
)

// IconRow is the cells an item row with its atlas icon takes: as tall as
// wide, the text in the middle line, after the icon.
const IconRow = 3

// MenuTheme is a game's menu colours: the plain text, the hotkey, a row
// with nothing to pick, the selection's ground, the three colour spans
// (InkGold, InkHead, InkRed) and the footer on the bottom border.
type MenuTheme struct {
	Text, Hot, Off, Sel, Gold, Head, Red, Foot color.RGBA
}

// MenuStyle is how a menu is drawn: its theme, the border tiles (corner,
// side, top), the font, RowGlyph (nil: none) a font per item row (a font
// list), Tight the border without a blank column inside, Tip a tooltip
// kept wholly on the window.
type MenuStyle struct {
	Theme    MenuTheme
	Tiles    Ribbon
	Glyph    func(byte) [8]byte
	RowGlyph func(row int, c byte) [8]byte
	Tight    bool
	Tip      bool
}

// MenuCanvas is a menu's images, kept between frames: Img the box at
// scale 1 and Buf its pixels.
type MenuCanvas struct {
	Img     *gpu.Image
	Buf     []byte
	footImg *gpu.Image
	footBuf []byte
}

// Draw draws a vertical menu: a box with the ribbon border, the items in
// the font, the hotkeys in Hot, the rows off[i] in Off, the selected one on
// Sel, an icon before each row of icons that is not nil (unscaled, or
// shrunk with its aspect kept to the row's height). It goes at scale k,
// centred on cx, its top at top, and returns rects with each item's row on
// dst appended.
func (m *MenuCanvas) Draw(dst *gpu.Image, st MenuStyle, title string, items []string, keys []byte, off []bool, icons []image.Image, sel int, k float64, cx, top int, rects []image.Rectangle) []image.Rectangle {
	glyph, t := st.Glyph, st.Theme
	items = slices.Clone(items)
	grey := make([]bool, len(items))
	copy(grey, off)
	hk := make([]int, len(items)) // each row's white letter
	for i := range items {
		var key byte
		if i < len(keys) {
			key = keys[i]
		}
		items[i], hk[i] = HotText(items[i], key)
	}
	title, foot, _ := strings.Cut(title, string(rune(InkFoot)))
	rows, hot, icon := items, hk, make([]image.Image, len(items))
	first := 0 // the first item's row
	copy(icon, icons)
	if title != "" { // a long question on two lines, so the box stays narrow
		t := append(TitleLines(title), "")
		rows, hot, icon = append(t, items...), append(slices.Repeat([]int{-1}, len(t)), hk...), append(make([]image.Image, len(t)), icon...)
		grey, first = append(make([]bool, len(t)), grey...), len(t)
		if sel >= 0 {
			sel += len(t)
		}
	}
	w, ch := 0, 2                // cells: a border and a blank column each side
	at := make([]int, len(rows)) // each row's first line, in cells
	rh := func(r int) int {
		if icon[r] != nil {
			return IconRow
		}
		return 1
	}
	for r, s := range rows {
		x := VisLen(s)
		if icon[r] != nil {
			x += IconRow
		}
		w, at[r], ch = max(w, x), ch-2, ch+rh(r)
	}
	w = max(w, len(foot)) // the footer sits on the bottom border, inside its corners
	pad := 2              // a border and a blank column each side
	if st.Tight {
		pad = 1 // the border only (the creation's drop-downs, on the sheet)
	}
	cw := w + 2*pad
	if gpu.Sized(&m.Img, cw*8, ch*8) {
		m.Buf = make([]byte, cw*8*ch*8*4)
	}
	green, white, dim := t.Text, t.Hot, t.Off
	clear(m.Buf)
	for i := 3; i < len(m.Buf); i += 4 {
		m.Buf[i] = 255
	}
	set := func(x, y int, c color.RGBA) {
		o := (y*cw*8 + x) * 4
		m.Buf[o], m.Buf[o+1], m.Buf[o+2] = c.R, c.G, c.B
	}
	ink := green // a colour span goes on over a wrapped line
	for r, s := range rows {
		y0, x0 := at[r]*8+8, pad
		if r == sel && t.Sel.A != 0 {
			for y := y0; y < y0+rh(r)*8; y++ { // the text selection: a blue ground
				for x := 8; x < (cw-1)*8; x++ {
					set(x, y, t.Sel)
				}
			}
		}
		if ic := icon[r]; ic != nil { // the icon, then the text in the middle line
			b, h := ic.Bounds(), IconRow*8
			f := min(1, float64(h)/float64(b.Dx()), float64(h)/float64(b.Dy()))
			w, ih := int(f*float64(b.Dx())), int(f*float64(b.Dy()))
			for y := range ih {
				for x := range w {
					if c := color.RGBAModel.Convert(ic.At(b.Min.X+x*b.Dx()/w, b.Min.Y+y*b.Dy()/ih)).(color.RGBA); c.A > 0x80 {
						set(16+(h-w)/2+x, y0+(h-ih)/2+y, c)
					}
				}
			}
			y0, x0 = y0+8, 2+IconRow
		}
		hk := hot[r] // the hotkey: white, as on the original's bar
		col := 0
		for i := 0; i < len(s); i++ {
			if s[i] == InkGold || s[i] == InkRed || s[i] == InkHead { // a span in colour, to the next mark
				switch {
				case ink != green:
					ink = green
				case s[i] == InkGold:
					ink = t.Gold
				case s[i] == InkHead:
					ink = t.Head
				default:
					ink = t.Red
				}
				continue
			}
			g := glyph(s[i])
			if st.RowGlyph != nil && r >= first { // the Font list: each line in its font
				g = st.RowGlyph(r-first, s[i])
			}
			c := ink
			switch {
			case grey[r]:
				c = dim
			case i == hk, r == sel && t.Sel.A == 0: // no selection ground (Sel transparent): the selected row in the hotkey colour
				c = white
			}
			for y := 0; y < 8; y++ {
				for b := 0; b < 8; b++ {
					if g[y]&(0x80>>b) != 0 {
						set((col+x0)*8+b, y0+y, c)
					}
				}
			}
			col++
		}
	}
	m.Img.WritePixels(m.Buf)

	Window(m.Img, st.Tiles, gpu.GeoM{}, cw, ch, nil)
	if foot != "" { // gold on black, right on the bottom border
		fw := len(foot) * 8
		if gpu.Sized(&m.footImg, fw, 8) {
			m.footBuf = make([]byte, fw*8*4)
		}
		for i := 0; i < len(m.footBuf); i += 4 {
			m.footBuf[i], m.footBuf[i+1], m.footBuf[i+2], m.footBuf[i+3] = 0, 0, 0, 255
		}
		gold := t.Foot
		for n := 0; n < len(foot); n++ {
			g := glyph(foot[n])
			for y := 0; y < 8; y++ {
				for b := 0; b < 8; b++ {
					if g[y]&(0x80>>b) != 0 {
						o := (y*fw + n*8 + b) * 4
						m.footBuf[o], m.footBuf[o+1], m.footBuf[o+2] = gold.R, gold.G, gold.B
					}
				}
			}
		}
		m.footImg.WritePixels(m.footBuf)
		fo := &gpu.DrawImageOptions{}
		fo.GeoM.Translate(float64((cw-1)*8-fw), float64((ch-1)*8))
		m.Img.DrawImage(m.footImg, fo)
	}

	if st.Tip { // a tooltip: wholly on the window, whatever its anchor
		bw, bh, tw, th := dst.Bounds().Dx(), dst.Bounds().Dy(), int(k)*cw*8, int(k)*ch*8
		cx = min(max(cx, tw/2), bw-tw/2)
		top = max(0, min(top, bh-th))
	}
	op := &gpu.DrawImageOptions{}
	op.GeoM.Scale(k, k)
	op.GeoM.Translate(float64(cx)-k*float64(cw*8)/2, float64(top))
	dst.DrawImage(m.Img, op)
	for r := len(rows) - len(items); r < len(rows); r++ {
		x0, y0 := float64(cx)-k*float64(cw*8)/2, float64(top)+k*float64((at[r]+1)*8)
		rects = append(rects, image.Rect(int(x0+k*8), int(y0), int(x0+k*float64((cw-1)*8)), int(y0+k*float64(rh(r)*8))))
	}
	return rects
}
