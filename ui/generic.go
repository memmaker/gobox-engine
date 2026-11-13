package ui

import (
	"image"
	"image/color"
	"math"
	"strings"

	"gobox/gpu"
)

// drawGeneric draws a page that has no screen of its own (a script's
// picture and text, a story page) over the whole window from its data:
// its pictures (the art, never the engine's frames or fills) as large as
// whole times their pixels fit over the text, centred; the text pane and
// the text over the view window in the remaster's boxes; its question as
// every page's. It returns the window.
func (p *Presenter) drawGeneric(dst *gpu.Image, note Note) Rect {
	dst.Fill(color.Black)
	b := dst.Bounds()
	cx := b.Dx() / 2
	k := scale(dst)
	var art []Picture
	var box image.Rectangle
	for _, pc := range p.pictures() {
		if pc.Img != nil && pc.File != "" {
			art, box = append(art, pc), box.Union(pc.At)
		}
	}
	if (p.flags().Full && (p.pane() == nil || !p.pane().Shown()) || p.introUp() != nil) && len(art) > 0 { // a whole-screen picture, a title's: fitted, its aspect kept, centred, sharp (bars only on a screen of another shape)
		area := b
		var strip image.Rectangle
		if p.flags().CreditsPic { // the game's credits picture: the picture above a K-sized row of the remaster credit, so it covers none of the picture's text
			strip = creditStrip(b)
			area.Max.Y = strip.Min.Y
		}
		s := math.Min(float64(area.Dx())/float64(box.Dx()), float64(area.Dy())/float64(box.Dy()))
		op := &gpu.DrawImageOptions{}
		op.GeoM.Scale(s, s)
		op.GeoM.Translate((float64(area.Dx())-s*float64(box.Dx()))/2, (float64(area.Dy())-s*float64(box.Dy()))/2)
		dst.DrawImage(pictureBox(&p.pg.pic, art, box), op)
		if p.flags().CreditsPic {
			p.drawCreditStrip(dst, note, strip)
		}
		p.drawPageMenu(dst, note)
		return Rect{0, 0, float64(b.Dx()), float64(b.Dy())}
	}
	txt := paneCells(p.pane()).Framed()
	th := 0.0
	if p.pane() != nil && p.pane().Shown() {
		th = k * float64(txt.Pixels().Dy())
	}
	gap := int(k * 8)
	if len(p.viewText()) > 0 { // text over the view window (WinText): in its box, in the art's place
		vw := viewWindow.Framed()
		r := vw.Pixels()
		vk := math.Max(1, math.Min(k, math.Floor((float64(b.Dy())-th-3*float64(gap))/float64(r.Dy()))))
		drawBox(dst, viewTextImage(&p.pg.pic, p.viewText(), r.Inset(8), note), p.border.val.tiles, vw, vk, cx, gap+int(vk*float64(r.Dy())/2))
	} else if pic := innerWindow.Framed(); len(art) > 0 && box.In(pic.Pixels().Inset(8)) { // a view-window picture (a treasure, a face): in the view window's frame, as over the 3D view
		r := pic.Pixels()
		s := math.Max(1, math.Floor(math.Min(float64(b.Dx()-2*gap)/float64(r.Dx()), (float64(b.Dy())-th-3*float64(gap))/float64(r.Dy()))))
		s = math.Min(s, math.Max(1, math.Floor(k)))
		drawBox(dst, pictureBox(&p.pg.pic, art, r.Inset(8)), p.border.val.tiles, pic, s, cx, gap+int(s*float64(r.Dy())/2))
	} else if fr := (FrontCellRect{X1: box.Min.X / 8, Y1: box.Min.Y / 8, X2: (box.Max.X - 1) / 8, Y2: (box.Max.Y - 1) / 8}).Framed(); len(art) > 0 && fr.Pixels().Inset(8) == box { // a cell-aligned picture (SotSB's big pictures): in a frame, as the original's screen
		r := fr.Pixels()
		s := fitScale(k, Host.Pic, func(s float64) bool { // the picture box: whole steps of K, as far as it fits over the text (its ribbon at a whole cell)
			return s*float64(r.Dx()) <= float64(b.Dx()-2*gap) && s*float64(r.Dy()) <= float64(b.Dy())-th-2*float64(gap)
		})
		drawBox(dst, pictureBox(&p.pg.pic, art, box), p.border.val.tiles, fr, s, cx, gap+int(s*float64(r.Dy())/2))
	} else if len(art) > 0 {
		s := math.Max(1, math.Floor(math.Min(float64(b.Dx()-2*gap)/float64(box.Dx()), (float64(b.Dy())-th-3*float64(gap))/float64(box.Dy()))))
		img := pictureBox(&p.pg.pic, art, box)
		op := &gpu.DrawImageOptions{}
		op.GeoM.Scale(s, s)
		op.GeoM.Translate(float64(cx)-s*float64(box.Dx())/2, float64(gap))
		dst.DrawImage(img, op)
	}
	if th > 0 { // the text along the bottom
		p.paneBox(dst, note, k, cx, b.Dy()-int(th/2)) // on the bottom edge
		p.menuFloor = b.Dy() - int(th)                // its question goes above the text, not over it
		defer func() { p.menuFloor = 0 }()
	}
	p.drawPageMenu(dst, note)
	if t := p.typingAsk(); t != nil && t.Prompt != "" { // a typed answer (a name after the sheet): its question and line, never an invisible question
		p.drawTyping(dst, note, math.Max(1, math.Floor(k)), cx, 0)
	}
	if m := p.askMenu(); m != nil && len(m.Commands) == 1 && !p.choosing() && p.typingAsk() == nil && p.askList() == nil && (strings.TrimSpace(m.Prompt) != "" || len(art) == 0 && th == 0) { // "press Enter": nothing to pick, the prompt alone above the text; a bare "press Enter" not over a page's picture or text (as over the 3D view)
		mk := math.Max(1, math.Floor(k))
		text := strings.TrimSpace(m.Prompt)
		if text == "" {
			text = m.Commands[0].Text
		}
		p.drawMenu(dst, note.Glyph, "", []string{text}, nil, nil, -1, mk, cx, int(float64(b.Dy()-int(th))/2)-int(mk*12))
	}
	return Rect{0, 0, float64(b.Dx()), float64(b.Dy())}
}

// creditStrip is where the remaster credit goes over a credits picture on
// window b: the bottom rows of cells, 8*K pixels each, as wide as the grid.
func creditStrip(b image.Rectangle) image.Rectangle {
	k := int(Scale(b))
	cols := b.Dx() / (8 * k)
	h := len(remasterRows(cols)) * 8 * k
	w := cols * 8 * k
	return image.Rect(b.Min.X+(b.Dx()-w)/2, b.Max.Y-h, b.Min.X+(b.Dx()+w)/2, b.Max.Y)
}

// drawCreditStrip draws the remaster credit, yellow on black, at K into strip.
func (p *Presenter) drawCreditStrip(dst *gpu.Image, note Note, strip image.Rectangle) {
	k := int(scale(dst))
	cols := strip.Dx() / (8 * k)
	rows := remasterRows(cols)
	m := image.NewRGBA(image.Rect(0, 0, cols*8, len(rows)*8))
	for y, s := range rows {
		for i := range s {
			glyphAt(m, note.Font(s[i], true), ((cols-len(s))/2+i)*8, y*8, egaInk[inkYellow], egaInk[inkBlack])
		}
	}
	gpu.Sized(&p.pg.credit, m.Rect.Dx(), m.Rect.Dy())
	p.pg.credit.WritePixels(m.Pix)
	if drawnAt != nil {
		drawnAt(drawnPage, dst, float64(k))
	}
	op := &gpu.DrawImageOptions{}
	op.GeoM.Scale(float64(k), float64(k))
	op.GeoM.Translate(float64(strip.Min.X), float64(strip.Min.Y))
	dst.DrawImage(p.pg.credit, op)
}
