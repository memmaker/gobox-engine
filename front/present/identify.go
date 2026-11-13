package present

import (
	"gobox/front"
	"image"
	"image/color"
	"image/draw"

	"gobox/ega"
	"gobox/gpu"
	"gobox/gpu/vector"
	"gobox/ui"
)

// The shop's Id (play.Trading.Known): what it revealed in a pop-up of
// its own over the dimmed screen, the icon large, the name and every line
// of what the item does in colour. Any key or click closes it.

// identified is the reveal still to show (nil: none).
func (p *Presenter) identified() *front.InvItem {
	s := p.state()
	sh := shopOf(s)
	if sh == nil || sh.Identified == nil {
		p.camp.idShut = false // the next reveal is a new one
		return nil
	}
	if p.camp.idShut && p.camp.idSeen == *sh.Identified {
		return nil
	}
	it := itemOf(partyOf(s), *sh.Identified)
	if it == nil {
		return nil
	}
	x := invItem(*it)
	x.Look = lines(it.Text)
	return &x
}

// identDone closes the reveal's pop-up if one is up; it reports whether one was.
func (p *Presenter) identDone() bool {
	if p.identified() == nil {
		return false
	}
	p.camp.idSeen, p.camp.idShut = *shopOf(p.state()).Identified, true
	return true
}

// drawIdentified draws the reveal's pop-up over dst.
func (p *Presenter) drawIdentified(dst *gpu.Image, glyph func(byte) [8]byte, it *front.InvItem) {
	const title, hint, iconH = "* Identified! *", "Splendid!", 2 * iconSlot // the icon at a whole 2x, so every atlas pixel stays square and even
	w := max(len(title), len(it.Name), len(hint))
	for _, l := range it.Look {
		w = max(w, len(l)+2)
	}
	cw := w + 4
	ch := 2 + 2 + iconH/8 + 1 + 2 + len(it.Look) + 2 // the border, the title and a blank, the icon and a blank, the name and a blank, the lines, a blank and the hint
	m := image.NewRGBA(image.Rect(0, 0, cw*8, ch*8))
	draw.Draw(m, m.Rect, image.Black, image.Point{}, draw.Src)
	centre := func(s string) int { return (cw*8 - 8*len(s)) / 2 }
	sparkle := []int{inkMagenta, inkYellow, inkCyan, inkWhite, inkGreen, inkRed} // the title's letters in turn: light magenta, yellow, cyan, white, green, red
	x0 := centre(title)
	for i := 0; i < len(title); i++ {
		ui.DrawText(m, glyph, title[i:i+1], x0+8*i, 8, ega.RGBA(sparkle[i%len(sparkle)]))
	}
	y := 24
	if ic := p.inv.iconImage(it.Name, it.Words); ic != nil { // the icon on the pop-up's black
		b := ic.Bounds()
		iw, ih := iconH, iconH // any atlas cell, scaled to the box
		at := image.Rect((cw*8-iw)/2, y, (cw*8+iw)/2, y+ih)
		for py := range ih {
			for px := range iw {
				if c := ic.At(b.Min.X+px*b.Dx()/iw, b.Min.Y+py*b.Dy()/ih); alpha(c) {
					m.Set(at.Min.X+px, at.Min.Y+py, c)
				}
			}
		}
	}
	y += iconH + 8
	ui.DrawText(m, glyph, it.Name, centre(it.Name), y, ega.RGBA(inkYellow))
	y += 16
	for _, l := range it.Look {
		ui.DrawText(m, glyph, "*", 16, y, col(inkMagenta))
		ui.DrawText(m, glyph, l, 32, y, col(inkCyan))
		y += 8
	}
	ui.DrawText(m, glyph, hint, centre(hint), y+8, ega.RGBA(inkWhite))

	img := upload(&p.idImg, m)
	window(img, p.border.tiles, gpu.GeoM{}, cw, ch, nil)
	b := dst.Bounds()
	vector.DrawFilledRect(dst, 0, 0, float32(b.Dx()), float32(b.Dy()), color.NRGBA{A: 0xA0})
	k := fitScale(scale(dst), 1, func(k float64) bool { // K, as far as it fits
		return k*float64(cw*8) <= float64(b.Dx()) && k*float64(ch*8) <= float64(b.Dy())
	})
	op := &gpu.DrawImageOptions{}
	op.GeoM.Scale(k, k)
	op.GeoM.Translate((float64(b.Dx())-k*float64(cw*8))/2, (float64(b.Dy())-k*float64(ch*8))/2)
	dst.DrawImage(img, op)
}

// alpha: c is not see-through.
func alpha(c interface{ RGBA() (r, g, b, a uint32) }) bool {
	_, _, _, a := c.RGBA()
	return a > 0x8000
}
