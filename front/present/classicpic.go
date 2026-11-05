// classicpic.go: the game's own pixels (Picture.Pix) in the page's ink.

package present

import (
	"image"
	"image/color"

	"gobox/ega"
	"gobox/front"
)

type pixKey struct {
	pix  string
	w    int
	mask int
	ink  [16]color.RGBA
	pal  *color.RGBA // the palette's first entry (a game shares one slice), nil for EGA
}

var pixPics = map[pixKey]image.Image{}

// egaInk is the EGA's own colours, for the pictures given in them.
var egaInk = func() (k [16]color.RGBA) {
	for i := range k {
		k[i] = ega.RGBA(i)
	}
	return k
}()

// pixImage is pc's pixels in ink (nil without pixels), cached.
func pixImage(pc front.Picture, ink [16]color.RGBA) image.Image {
	if pc.Pix == nil {
		return nil
	}
	if len(pixPics) > 4096 { // ponytail: whole cache dropped when big
		clear(pixPics)
	}
	k := pixKey{string(pc.Pix), pc.PixW, -1, ink, nil}
	if pc.Palette != nil {
		k.pal = &pc.Palette[0]
	}
	if pc.Mask != nil {
		k.mask = int(*pc.Mask)
	}
	if m, ok := pixPics[k]; ok {
		return m
	}
	var m *image.RGBA
	if pc.Palette != nil {
		m = front.PalettedRGBA(pc.Pix, pc.PixW, pc.Palette)
	} else {
		m = front.IndexedRGBA(pc.Pix, pc.PixW, k.mask, &ink)
	}
	pixPics[k] = m
	return m
}

// picImage is pc as an image: its pixels in ink, else the DAX picture its
// File, Block and Frame name (EGA colours), else nil.
func picImage(pc front.Picture, ink [16]color.RGBA) image.Image {
	if pc.Img != nil {
		return pc.Img
	}
	if ink == egaInk && pc.Palette == nil {
		if m := poolImage(pc); m != nil {
			return m
		}
	}
	if pc.Pix != nil {
		return pixImage(pc, ink)
	}
	if pc.File != "" {
		return daxMasked(pc.File, pc.Block, pc.Frame, refMask(pc))
	}
	return nil
}

// indexedPictures is ps with the pictures given as pixels converted in ink.
func indexedPictures(ps []front.Picture, ink [16]color.RGBA) []front.Picture {
	var out []front.Picture
	for i, pc := range ps {
		if pc.Pix == nil || pc.Img != nil {
			continue
		}
		if out == nil {
			out = append([]front.Picture(nil), ps...)
		}
		out[i].Img = pixImage(pc, ink)
	}
	if out == nil {
		return ps
	}
	return out
}

// pixView is v with the pictures given as pixels (the ...Pix fields)
// coloured into their images, and an unset Ink the EGA's; the game's own structs are copied, not changed.
func pixView(v front.View) front.View {
	if v.Ink == ([16]color.RGBA{}) { // unset: the EGA's own
		v.Ink = egaInk
	}
	ink := v.Ink
	if v.BorderPix[0].Pix != nil {
		for i, pc := range v.BorderPix {
			if v.Border[i] == nil {
				v.Border[i] = pixImage(pc, egaInk)
			}
		}
		for i, pc := range v.BorderCapsPix {
			if v.BorderCaps[i] == nil {
				v.BorderCaps[i] = pixImage(pc, egaInk)
			}
		}
	}
	if m := v.AreaMap; m != nil && m.SquarePix != nil && m.Squares == nil {
		c := *m
		c.Squares = make([][]image.Image, len(m.SquarePix))
		for y, row := range m.SquarePix {
			c.Squares[y] = make([]image.Image, len(row))
			for x, pc := range row {
				c.Squares[y][x] = picImage(pc, ink)
			}
		}
		c.Marker = picImage(m.MarkerPix, ink)
		v.AreaMap = &c
	}
	v.Battle = pixBattle(v.Battle, ink)
	if w := v.Walls; w != nil && w.FrontPix != nil && w.Fronts == nil {
		c := *w
		c.Fronts = make(map[int]image.Image, len(w.FrontPix))
		for k, pc := range w.FrontPix {
			c.Fronts[k] = picImage(pc, ink)
		}
		v.Walls = &c
	}
	if w := v.Walls; w != nil && w.Backdrop != nil && w.Backdrop.Img == nil {
		c, b := *w, *w.Backdrop
		b.Img = picImage(b, ink)
		c.Backdrop = &b
		v.Walls = &c
	}
	if pixParty(v.Party) {
		v.Party = append([]front.Member(nil), v.Party...)
		for i := range v.Party {
			if m := &v.Party[i]; m.Head == nil {
				m.Head = pixImage(m.HeadPix, egaInk)
			}
		}
	}
	if c := v.Camp; c != nil && len(c.Bestiary) > 0 && c.Bestiary[0].Icon == nil {
		cc := *c
		cc.Bestiary = append([]front.Beast(nil), c.Bestiary...)
		for i := range cc.Bestiary {
			b := &cc.Bestiary[i]
			if b.Icon == nil {
				b.Icon = pixImage(b.IconPix, ink)
			}
			if b.Pic == nil {
				b.Pic = pixImage(b.PicPix, ink)
			}
		}
		v.Camp = &cc
	}
	return v
}

func pixParty(ms []front.Member) bool {
	for _, m := range ms {
		if m.Head == nil && m.HeadPix.Pix != nil {
			return true
		}
	}
	return false
}

// pixBattle is b with its pictures given as pixels coloured in ink.
func pixBattle(b *front.Battle, ink [16]color.RGBA) *front.Battle {
	if b == nil || b.TilePix == nil && b.IconPix == nil && b.EffectPix == nil {
		return b
	}
	c := *b
	if b.Tiles == nil && b.TilePix != nil {
		c.Tiles = make([]image.Image, len(b.TilePix))
		for i, pc := range b.TilePix {
			c.Tiles[i] = picImage(pc, ink)
		}
	}
	c.Units = append([]front.Unit(nil), b.Units...)
	for i := range c.Units {
		if u := &c.Units[i]; u.Icon == nil && i < len(b.IconPix) {
			u.Icon = picImage(b.IconPix[i], ink)
		}
	}
	c.Effects = append([]front.Effect(nil), b.Effects...)
	for i := range c.Effects {
		if f := &c.Effects[i]; f.Pic == nil && i < len(b.EffectPix) {
			f.Pic = picImage(b.EffectPix[i], ink)
		}
	}
	return &c
}
