package present

import (
	"gobox/front"
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
func (p *Presenter) drawGeneric(dst *gpu.Image, v front.View, note Note) Rect {
	dst.Fill(color.Black)
	b := dst.Bounds()
	cx := b.Dx() / 2
	k := scale(dst)
	var art []front.Picture
	var box image.Rectangle
	for _, pc := range v.Pictures {
		if pc.Img != nil && pc.File != "" {
			art, box = append(art, pc), box.Union(pc.At)
		}
	}
	if (v.Full || v.Intro != nil) && len(art) > 0 { // a whole-screen picture, a title's: fitted, its aspect kept, centred, sharp (bars only on a screen of another shape)
		if in := v.Intro; in != nil && titleBackdrop > 0 && p.intro.Phase == front.IntroBackdrop && p.mm.title == nil { // the picture the main menu comes in over
			m := image.NewRGBA(image.Rect(0, 0, box.Dx(), box.Dy()))
			drawPictures(m, art, box.Min)
			p.mm.title = m
		}
		s := math.Min(float64(b.Dx())/float64(box.Dx()), float64(b.Dy())/float64(box.Dy()))
		op := &gpu.DrawImageOptions{}
		op.GeoM.Scale(s, s)
		op.GeoM.Translate((float64(b.Dx())-s*float64(box.Dx()))/2, (float64(b.Dy())-s*float64(box.Dy()))/2)
		dst.DrawImage(pictureBox(&p.pg.pic, art, box), op)
		p.drawPageMenu(dst, note, v)
		return Rect{0, 0, float64(b.Dx()), float64(b.Dy())}
	}
	txt := paneBox.Framed()
	th := 0.0
	if v.Pane != nil && v.Pane.Shown() {
		th = k * float64(txt.Pixels().Dy())
	}
	gap := int(k * 8)
	if len(v.ViewText) > 0 { // text over the view window (WinText): in its box, in the art's place
		vw := viewWindow.Framed()
		r := vw.Pixels()
		vk := math.Max(1, math.Min(k, math.Floor((float64(b.Dy())-th-3*float64(gap))/float64(r.Dy()))))
		drawBox(dst, viewTextImage(&p.pg.pic, v.ViewText, r.Inset(8), note), p.border.tiles, vw, vk, cx, gap+int(vk*float64(r.Dy())/2))
	} else if len(art) > 0 {
		s := math.Max(1, math.Floor(math.Min(float64(b.Dx()-2*gap)/float64(box.Dx()), (float64(b.Dy())-th-3*float64(gap))/float64(box.Dy()))))
		img := pictureBox(&p.pg.pic, art, box)
		op := &gpu.DrawImageOptions{}
		op.GeoM.Scale(s, s)
		op.GeoM.Translate(float64(cx)-s*float64(box.Dx())/2, float64(gap))
		dst.DrawImage(img, op)
	}
	if th > 0 { // the text along the bottom
		p.paneBox(dst, note, v, k, cx, b.Dy()-gap-int(th/2))
		p.menuFloor = b.Dy() - gap - int(th) // its question goes above the text, not over it
		defer func() { p.menuFloor = 0 }()
	}
	p.drawPageMenu(dst, note, v)
	if m := askMenu(v); m != nil && len(m.Items) == 1 && !choosing(v) && typing(v) == nil && askList(v) == nil { // "press Enter": nothing to pick, the prompt alone above the text
		mk := math.Max(1, math.Floor(k))
		text := strings.TrimSpace(m.Prompt)
		if text == "" {
			text = m.Items[0].Text
		}
		p.drawMenu(dst, note.Glyph, "", []string{text}, nil, nil, -1, mk, cx, int(float64(b.Dy()-int(th))/2)-int(mk*12))
	}
	return Rect{0, 0, float64(b.Dx()), float64(b.Dy())}
}
