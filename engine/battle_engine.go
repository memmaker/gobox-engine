package engine

import (
	"image"
	"slices"

	"gobox/asset"
	"gobox/rules"
)

// battleShown: the combat map is up (battle builds it).
func (e *Engine) battleShown() bool {
	return e.worldWord(rules.WordBeginFlag) == 5 && e.sqrpaci[0] != nil && e.ruleset().Check(rules.BoardUp{}, nil).Yes && e.snap.cmapOn && !e.snap.sheetOn
}

// battle builds the snapshot's Battle (nil outside combat).
func (e *Engine) battle() *Battle {
	e.present()
	s := &e.snap
	if !e.battleShown() {
		return nil
	}
	b := &Battle{
		Cursor: s.cursor, CursorOn: s.cursorOn,
		Marks: s.marks, Reach: s.reach, Paths: s.paths, Targets: s.targets, Turn: s.turn, Area: e.ext.aoe != 0, AreaAt: s.area, Skipped: s.skipped, AimWhy: s.aimWhy, Dungeon: s.dungeon, View: e.cview,
		Strike: s.strike, Log: slices.Clone(e.ext.battleLog[max(0, len(e.ext.battleLog)-100):]),
	}
	g := e.fact(rules.NoChar, rules.GroundFact{}).(rules.Ground)
	for _, d := range g.Footprints[e.ccursor.size] {
		b.CursorAt = append(b.CursorAt, s.cursor.Add(image.Pt(d.X, d.Y)))
	}
	b.Ground = g.Squares
	ground := e.groundOf(e.art(e.sqrpaci[0], 0)) // the shown art's (F12): 0 on DOS, a grey on the Amiga
	b.Floor = e.pal.RGBA(ground)
	b.TilePix = make([]Picture, 0x100)
	for t, im := range e.sqrpaci {
		for f := 0; im != nil && f < min(im.Count, 0x80); f++ {
			b.TilePix[t*0x80+f] = e.tilePix(im, f, int(ground))
		}
	}
	active := rules.NoChar
	if e.sel < e.partySize() {
		active = e.partyIDs()[e.sel]
	}
	for _, p := range e.fact(rules.NoChar, rules.UnitsFact{}).(rules.Units) {
		i := p.Number
		switch p.ID {
		case s.striker[0]:
			b.Strike.From = i
		case s.striker[1]:
			b.Strike.To = i
		}
		u := Unit{ID: i, X: p.At.X, Y: p.At.Y, Size: p.Size, Party: p.Party,
			Facing: p.Facing, Active: p.ID == active, Name: p.Name, HP: p.HP, MaxHP: p.MaxHP, AC: p.AC, Hit: s.hits[i], Dmg: p.Damage}
		switch {
		case !p.Active:
			u.Status = e.statusName(p.ID)
		case p.Helpless:
			u.Status = "Helpless"
		}
		if p.ID == e.ext.panel {
			b.Panel = i
		}
		for _, q := range p.Squares {
			u.Squares = append(u.Squares, image.Pt(q.X, q.Y))
		}
		u.IconSize = image.Pt(24, 24)
		var icon Picture
		if im := e.icons[p.Icon][0]; im != nil {
			icon, u.IconSize = e.picPix(im, 0, u.Facing > 3, -1), image.Pt(im.Width*8, im.Height)
		}
		if im := e.iconPNG(p.ID, 0, u.Facing > 3); im != nil { // the dropped PNG (iconpng.go)
			u.Icon, icon = im, Picture{}
		}
		b.Units = append(b.Units, u)
		b.IconPix = append(b.IconPix, icon)
	}
	return b
}

// picPix is frame f of im (or its DOS twin) as View pixels in the screen's
// current palette (Ink; combat swaps colours 0 and 8), mirrored with
// flip: Clear pixels (and with ground, that colour) transparent, the
// values 17..AmigaBase-1 their colour register's, as the screen shows them.
func (e *Engine) picPix(im *asset.Image, f int, flip bool, ground int) Picture {
	im = e.art(im, f)
	w := im.Width * 8
	d := im.Frame(f)
	out := make([]uint8, len(d))
	for i, p := range d {
		x := i % w
		if flip {
			x = w - 1 - x
		}
		switch {
		case p == asset.Clear || int(p) == ground:
			p = asset.Clear
		case p > asset.Clear && p < asset.AmigaBase:
			p &= 15
		}
		out[i-i%w+x] = p
	}
	m := uint8(asset.Clear)
	return Picture{Pix: out, PixW: w, Mask: &m, Palette: im.Palette}
}

// tilePix is picPix of a combat ground tile, converted once: the board is
// rebuilt every frame and its tiles never change (unlike icons, which
// recolour in place).
func (e *Engine) tilePix(im *asset.Image, f, ground int) Picture {
	k := tileKey{e.art(im, f), f, ground}
	if p, ok := e.tiles[k]; ok {
		return p
	}
	if e.tiles == nil {
		e.tiles = map[tileKey]Picture{}
	}
	p := e.picPix(im, f, false, ground)
	e.tiles[k] = p
	return p
}

// groundOf is groundColour(im), counted once per tile set.
func (e *Engine) groundOf(im *asset.Image) uint8 {
	if c, ok := e.grounds[im]; ok {
		return c
	}
	if e.grounds == nil {
		e.grounds = map[*asset.Image]uint8{}
	}
	c := groundColour(im)
	e.grounds[im] = c
	return c
}

type tileKey struct {
	art       *asset.Image // the shown art (DOS or Amiga)
	f, ground int
}

// cmapShown notes the combat map is drawn: the view shows it from then on.
func (e *Engine) cmapShown() {
	e.snap.cursorOn = false // a fresh page: drawCursor puts it back
	e.snap.cmapOn = true
}

// groundColour is the pixel value the combat tiles paint the bare ground
// in: their most common one (0 on DOS, a slot's grey on the Amiga).
func groundColour(im *asset.Image) uint8 {
	var n [256]int
	for _, p := range im.Data {
		n[p]++
	}
	best := 0
	for c := range n {
		if c != asset.Clear && n[c] > n[best] {
			best = c
		}
	}
	return uint8(best)
}
