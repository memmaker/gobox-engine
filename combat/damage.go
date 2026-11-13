package combat

import (
	"image"
	"image/color"
	"math"
	"strconv"

	"gobox/gpu"
)

// Floating damage numbers (Graphics, Damage Numbers): the engine's EvDamage
// and EvHeal, as the host plays them, rise and fade over the unit's head.

const floatLife = 1.2 // seconds a number shows

// floater is one number in the air.
type floater struct {
	at   vec // over the head when it was hit
	img  *gpu.Image
	life float64
}

// Float starts number n over unit id's head: hit points lost, green when heal.
func (p *Presenter) Float(id, n int, heal bool) {
	at, ok := p.tops[id]
	if !ok || p.Glyph == nil || n <= 0 {
		return
	}
	c := p.Theme.Hurt
	if heal {
		c = p.Theme.Heal
	}
	p.floats = append(p.floats, floater{at: at, img: gpu.NewImageFromImage(textImage(p.Glyph, strconv.Itoa(n), c, p.Theme.TextOutline))})
}

// FloatText starts word s (MISS) over unit id's head, in the text colour.
func (p *Presenter) FloatText(id int, s string) {
	if at, ok := p.tops[id]; ok && p.Glyph != nil {
		p.floats = append(p.floats, floater{at: at, img: gpu.NewImageFromImage(textImage(p.Glyph, s, p.Theme.Letter, p.Theme.TextOutline))})
	}
}

// textImage is s in the game's font, in c on a dark outline.
func textImage(glyph func(byte) [8]byte, s string, c, outline color.RGBA) *image.RGBA {
	im := image.NewRGBA(image.Rect(0, 0, len(s)*8+2, 10))
	set := func(dx, dy int, c color.RGBA) {
		for i := 0; i < len(s); i++ {
			g := glyph(s[i])
			for y := range 8 {
				for x := range 8 {
					if g[y]&(0x80>>x) != 0 {
						im.SetRGBA(1+i*8+x+dx, 1+y+dy, c)
					}
				}
			}
		}
	}
	for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		set(d[0], d[1], outline)
	}
	set(0, 0, c)
	return im
}

// drawDamage moves, fades and draws the numbers at the UI's glyph size.
func (p *Presenter) drawDamage(dst *gpu.Image, vp gpu.Mat4, dt float64) {
	k := float64(max(1, p.K))
	live := p.floats[:0]
	for _, f := range p.floats {
		f.life += dt
		if f.life >= floatLife {
			f.img.Deallocate()
			continue
		}
		live = append(live, f)
		x, y, ok := project(dst, vp, f.at.add(vec{0, f.life * .6, 0}))
		if !ok {
			continue
		}
		a := float32(math.Min(1, 3*(1-f.life/floatLife)))
		b := f.img.Bounds()
		op := &gpu.DrawImageOptions{}
		op.GeoM.Scale(k, k)
		op.GeoM.Translate(math.Round(x-float64(b.Dx())*k/2), math.Round(y-float64(b.Dy())*k))
		op.ColorScale.Scale(a, a, a, a)
		dst.DrawImage(f.img, op)
	}
	p.floats = live
}
