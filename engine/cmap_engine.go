package engine

import (
	"strings"

	"gobox/adventure"
	"gobox/asset"
)

// LayTiles = ovr33:0022 LoadBackpacIcons (name, first, n): block l.Block
// of l.File copied into the tile set from tile l.First on, frames 0..l.Last
// (POR's Amiga RANDCOM has two more, which would land on the tiles after it).
func (e *Engine) LayTiles(l adventure.TileLoad) {
	if l.First > 0x7F {
		return // the original stops with a fatal error
	}
	e.tilesChanged()
	im := e.LoadPic(l.File, e.tileBlock(l))
	t := e.sqrpaci[0]
	if im == nil || t == nil {
		return
	}
	// the frames change in place: the tiles' pictures follow their pixels (picRGBA)
	for f := 0; f <= l.Last && f < im.Count && l.First+f < t.Count && im.BPP == t.BPP; f++ {
		copy(t.Frame(l.First+f), im.Frame(f))
		if im.Alt != nil && t.Alt != nil && f < im.Alt.Count && l.First+f < t.Alt.Count && im.Alt.BPP == t.Alt.BPP {
			copy(t.Alt.Frame(l.First+f), im.Alt.Frame(f))
		}
	}
}

// tileBlock is l's block number in the game's present state.
func (e *Engine) tileBlock(l adventure.TileLoad) int {
	switch l.From {
	case adventure.TileBlockWalls:
		return e.mapPairA[0] & 0xFF
	case adventure.TileBlockAreaSet:
		if s := int(e.world.SymbolSet); s == 1 || s == 2 { // pod-re LoadCombatTileSet: zone0.tileSet, the layout's SymbolSet
			return s
		}
		return 0
	}
	return l.Block
}

// blankTiles is the tile set of a game without a sheet of its own (sotsb,
// pod): 0x80 blank 24x24 tiles the Tiles and combat sheets fill.
// ponytail: no DOS twin, so F12 shows the Amiga tiles; give it one if a game with Amiga art needs it.
func blankTiles() *asset.Image { return asset.NewImage(0x80, 3, 24) }

// layAreaTiles lays the area map's tiles (BACPAC) over the tile set.
func (e *Engine) layAreaTiles() {
	for _, l := range e.artNames().AreaTiles {
		e.LayTiles(l)
	}
}

// loadAreaIcon loads the area map's party icon block into slot (ICON); a
// game without one loads nothing.
func (e *Engine) loadAreaIcon(block, slot int) {
	if n := e.artNames().AreaIcon; n != "" {
		e.LoadIconPic(n, block, slot)
	}
}

// recolour = 0464:1E5C with DS:0CD2 → to: map the colours of frame 0 (the
// original takes a frame number, always passed 0, so only that frame changes).
func (e *Engine) recolour(im *asset.Image, tt *[16]byte) {
	if im == nil || im.Count == 0 {
		return
	}
	e.tilesChanged()
	e.recolour(im.Alt, tt)
	from := IconPalette
	d := im.Frame(0)
	for i, p := range d {
		for k := 0; k < 16; k++ {
			if from[k] == p {
				d[i] = tt[k]
				break
			}
		}
	}
}

// LoadIconPic = ovr33:01C8 (name, block, slot): the two poses of a combat
// icon, block and block+0x80, loaded masked on colour 0.
func (e *Engine) LoadIconPic(name string, block, slot int) {
	up := strings.ToUpper(name)
	load := func(n string, b int) *asset.Image { return e.LoadPicMasked(n, b, byte(e.artNames().IconMask)) }
	ic := &e.icons[slot]
	switch {
	case strings.HasPrefix(up, "CHEAD") || strings.HasPrefix(up, "CBODY"):
		if up[len(up)-1] == 'T' {
			block += 0x40
		}
		n := name[:len(name)-1]
		ic[0], ic[1] = load(n, block), load(n, block+0x80)
	case up == strings.ToUpper(e.artNames().CombatSprites) || up == strings.ToUpper(e.artNames().AreaIcon) ||
		e.artNames().MemberIcons != "" && up == strings.ToUpper(e.artNames().MemberIcons):
		ic[0], ic[1] = load(name, block), load(name, block+0x80)
		if up == strings.ToUpper(e.artNames().AreaIcon) {
			e.recolour(ic[0], &IconPaletteCF2)
			e.recolour(ic[1], &IconPaletteCF2)
		}
	default:
		n := e.set.File(name)
		if up == "CPIC" { // the monsters' icons: sotsb's and pod's are always CPIC1
			n = e.monsterSet().File(name)
		}
		ic[0], ic[1] = load(n, block), load(n, block+0x80)
		if !e.artNames().PlainIcons {
			e.recolour(ic[0], &IconPaletteCE2)
			e.recolour(ic[1], &IconPaletteCE2)
		}
	}
}

// freeIcon = ovr33:014E (slot).
func (e *Engine) freeIcon(n byte) { e.icons[n] = [2]*asset.Image{} }

// iconToSprite = 00BA:1892 (slot, pose, frame, flip): copy an icon into a
// frame of the DS:6E36 sprite.
func (e *Engine) iconToSprite(slot, pose, frame int, flip bool) {
	im := e.icons[slot][pose]
	if im == nil {
		return
	}
	if e.spr == nil || e.spr.BPP != im.BPP {
		e.spr = asset.NewImage(4, im.Width, im.Height)
	}
	if flip {
		im = im.Clone()
		im.FlipLeftToRight()
	}
	e.spr.Mask = im.Mask
	copy(e.spr.Frame(frame), im.Frame(0))
	if im.Alt != nil { // the DOS twin alike (F12)
		if e.spr.Alt == nil || e.spr.Alt.BPP != im.Alt.BPP {
			e.spr.Alt = asset.NewImage(4, im.Alt.Width, im.Alt.Height)
		}
		e.spr.Alt.Mask = im.Alt.Mask
		copy(e.spr.Alt.Frame(frame), im.Alt.Frame(0))
	}
}

// combatViewEnter = ovr32:0105 / combatViewLeave = ovr32:00E3: the EGA
// palette entries 0 and 8 swap for the combat screen (EGA only, DS:4A1E == 1).
func (e *Engine) combatViewEnter() {
	if e.Cfg.Graphics == 'E' {
		e.SetPaletteColor(0, 0)
		e.SetPaletteColor(8, 8)
	}
}

func (e *Engine) combatViewLeave() {
	if e.Cfg.Graphics == 'E' {
		e.SetPaletteColor(8, 0)
		e.SetPaletteColor(0, 8)
	}
}
