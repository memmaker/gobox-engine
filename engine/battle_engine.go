package engine

import (
	"image"
	"slices"
	"strconv"

	"gobox/asset"
	"gobox/rules"
)

// battleShown: the combat map is up (battle builds it).
func (e *Engine) battleShown() bool {
	return e.phase == rules.PhaseFight && e.sqrpaci[0] != nil && e.ruleset().Check(rules.BoardUp{}, nil).Yes && e.snap.cmapOn && !e.snap.sheetOn
}

// battle builds the snapshot's Battle (nil outside combat).
func (e *Engine) battle() *Battle {
	s := &e.snap
	if !e.battleShown() {
		return nil
	}
	b := &Battle{
		Marks: s.marks, Reach: s.reach, Paths: s.paths, Targets: s.targets, Turn: s.turn, Area: e.ext.aoe != 0, AreaAt: s.area, Skipped: s.skipped, AimWhy: s.aimWhy, Dungeon: s.dungeon,
		Strike: s.strike, Log: slices.Clone(e.ext.battleLog[max(0, len(e.ext.battleLog)-100):]),
	}
	g := e.fact(rules.NoChar, rules.GroundFact{}).(rules.Ground)
	b.Ground = g.Squares
	b.TilePix = e.combatTiles()
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
			u.Squares = append(u.Squares, q)
		}
		u.IconSize = image.Pt(24, 24)
		var icon Picture
		if im := e.icons[p.Icon][0]; im != nil {
			icon, u.IconSize = e.picPix(im, 0, u.Facing > 3, nil), image.Pt(im.Width*8, im.Height)
		}
		b.Units = append(b.Units, u)
		b.IconPix = append(b.IconPix, icon)
	}
	return b
}

// combatTiles are the loaded tile sets' pictures by tile number (no Pix: none).
func (e *Engine) combatTiles() []Picture {
	px := make([]Picture, 0x100)
	for t, im := range e.sqrpaci {
		for f := 0; im != nil && f < min(im.Count, 0x80); f++ {
			px[t*0x80+f] = e.tilePix(im, f, e.sqrpaci[0]) // its ground: 0 on DOS, a grey on the Amiga
		}
	}
	return px
}

// combatTileArt is tile i's name in Art().
func combatTileArt(i int) string { return "combat/tile/" + strconv.Itoa(i) }

// picPix is frame f of im (or its DOS twin) as View pixels in the screen's
// current palette (Ink; combat swaps colours 0 and 8), mirrored with
// flip: Clear pixels (and with ground, that colour) transparent, the
// values 17..AmigaBase-1 their colour register's, as the screen shows them.
//
// ground is the picture whose commonest colour is the bare ground (nil:
// none); the DOS twin (Alt) takes its twin's.
func (e *Engine) picPix(im *asset.Image, f int, flip bool, ground *asset.Image) Picture {
	p := e.pixOf(im, f, flip, e.groundIn(ground))
	if a := im.Alt; a != nil && f < a.Count {
		if ground != nil && ground.Alt != nil {
			ground = ground.Alt
		}
		alt := e.pixOf(a, f, flip, e.groundIn(ground))
		p.Alt = &alt
	}
	return p
}

// groundIn is im's ground colour (groundOf), -1 for none.
func (e *Engine) groundIn(im *asset.Image) int {
	if im == nil {
		return -1
	}
	return int(e.groundOf(im))
}

// pixOf is picPix of im itself (no twin).
func (e *Engine) pixOf(im *asset.Image, f int, flip bool, ground int) Picture {
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

// tilePix is picPix of a combat ground tile, converted once per tile set:
// the board is rebuilt every frame; laying or recolouring frames in place
// drops the conversions (tilesChanged).
func (e *Engine) tilePix(im *asset.Image, f int, ground *asset.Image) Picture {
	k := tileKey{im, f, ground}
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

// tilesChanged drops the tile conversions and ground counts: frames were
// written in place (LayTiles, recolour, recolourFrame), so the images that
// key them hold other pixels now (tileGen: the overland's art too).
func (e *Engine) tilesChanged() { e.tiles, e.grounds = nil, nil; e.tileGen++ }

type tileKey struct {
	im     *asset.Image
	f      int
	ground *asset.Image
}

// cmapShown notes the combat map is drawn: the view shows it from then on.
func (e *Engine) cmapShown() {
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
