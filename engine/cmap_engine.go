package engine

import (
	"image"

	"strings"

	"gobox/asset"
	"gobox/core"
	"gobox/core/goldbox"
)

// drawPage = 0464:0D71 draws a frame at page cell (x,y), clipped to the
// page (which starts a cell in). masked skips the image's mask colour.
// Outside combat only (pictures, animations, the area map, the icon
// editor): the combat page is gone.
func (e *Engine) drawPage(x, y int, im *asset.Image, frame int, masked bool) {
	if im == nil || frame >= im.Count {
		return
	}
	m := -1
	if masked {
		m = im.Mask
	}
	e.picture(8+x*8, 8+y*8, im.Width*8, im.Height, im.Frame(frame), m, im, frame)
	e.pg.layers[len(e.pg.layers)-1].clip = image.Rect(8, 8, 8+pageCells*8, 8+pageCells*8)
}

// drawTile = ovr33:00F1 (x, y, tile, page): a ground tile from the tile
// sets (DS:5E52, tiles from 0x80 in DS:5E56); the area map's icons.
func (e *Engine) drawTile(x, y int, t byte) {
	im, f := e.sqrpaci[0], int(t)
	if t > 0x7F {
		im, f = e.sqrpaci[1], int(t&0x7F)
	}
	e.drawPage(x, y, im, f, false)
}

// drawIcon = ovr33:04FC (x, y, dir, pose, slot): a combat icon on page
// square (x,y), mirrored when facing directions 4..7 (the area map's party
// marker, the icon editor).
func (e *Engine) drawIcon(x, y int, dir, pose, slot byte) {
	im := e.icons[slot][pose&1]
	if im == nil {
		return
	}
	if dir > 3 {
		im = im.Clone()
		im.FlipLeftToRight()
	}
	e.drawPage(x*3, y*3, im, 0, true)
}

// LayTiles = ovr33:0022 LoadBackpacIcons (name, first, n): block l.Block
// of l.File copied into the tile set from tile l.First on, frames 0..l.Last
// (POR's Amiga RANDCOM has two more, which would land on the tiles after it).
func (e *Engine) LayTiles(l TileLoad) {
	if l.First > 0x7F {
		return // the original stops with a fatal error
	}
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
func (e *Engine) tileBlock(l TileLoad) int {
	switch l.From {
	case TileBlockWalls:
		return e.mapPairA[0] & 0xFF
	case TileBlockAreaSet:
		if s := e.worldWord(core.WordAreaScreen); s == 1 || s == 2 {
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
	load := func(n string, b int) *asset.Image { return e.LoadPicMasked(n, b, 0) }
	ic := &e.icons[slot]
	switch {
	case strings.HasPrefix(up, "CHEAD") || strings.HasPrefix(up, "CBODY"):
		if up[len(up)-1] == 'T' {
			block += 0x40
		}
		n := name[:len(name)-1]
		ic[0], ic[1] = load(n, block), load(n, block+0x80)
	case up == strings.ToUpper(e.artNames().CombatSprites) || up == strings.ToUpper(e.artNames().AreaIcon):
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
		e.recolour(ic[0], &IconPaletteCE2)
		e.recolour(ic[1], &IconPaletteCE2)
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

// hitColours = 00BA:19DE (slot): the four frames of the sprite from an icon.
func (e *Engine) hitColours(slot int) {
	e.iconToSprite(slot, 0, 0, false)
	e.iconToSprite(slot, 0, 1, true)
	e.iconToSprite(slot, 1, 2, true)
	e.iconToSprite(slot, 1, 3, false)
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

// drawCursor = ovr32:0127 (x, y): the cursor on (x, y), with its icon when
// cmap.cursor is set (the snapshot's Cursor; the view draws it).
func (e *Engine) drawCursor(x, y int) {
	e.snap.cursor, e.snap.cursorOn = image.Pt(x, y), e.ccursor.icon
}

// combatCursor is the combat map's cursor: with its icon (cmap.cursor),
// over size's footprint (cmap.cursorSize).
type combatCursor struct {
	icon bool
	size int
}

// freshCursor is the cursor of a fresh combat map.
var freshCursor = combatCursor{size: 1}

// eraseCursor = ovr32:02D5 (x, y).
func (e *Engine) eraseCursor(x, y int) { e.snap.cursorOn = false }

// combatView is the combat view window (extra): 12 x 7 squares, scrolled as
// the original scrolled its 7x7 (recentre); the battle view's camera follows it.
var combatView = goldbox.View{W: 12, H: 7}

// inView = ovr32:06E9 on map squares: (x, y) lies in the view at cview.
func (e *Engine) inView(x, y int) bool { return combatView.In(x-e.cview.X, y-e.cview.Y) }

// recentre = ovr32:07D4 (x, y, margin): scroll the view so (x,y) is at
// least margin squares inside it (0xFF: always centre).
func (e *Engine) recentre(x, y, r int) bool {
	w := combatView.W
	half := w / 2
	cx, cy := e.cview.X+half, e.cview.Y+3
	rr := r
	if r == 0xFF {
		rr = 0
	}
	if r != 0xFF && x >= cx-rr && x <= cx+rr && y >= cy-rr && y <= cy+rr {
		return false
	}
	if x < cx-rr {
		for x < cx && cx > half {
			cx--
		}
	} else if x > cx+rr {
		for x > cx && cx < 0x31-(w-1-half) {
			cx++
		}
	}
	if y < cy-rr {
		for y < cy && cy > 3 {
			cy--
		}
	} else if y > cy+rr {
		for y > cy && cy < 0x15 {
			cy++
		}
	}
	e.cview.X, e.cview.Y = cx-half, cy-3
	e.cmapShown()
	return true
}

// combatCentre = ovr32:0996 (x, y, margin, dir): bring the square next to
// (x,y) in direction dir (8 = itself) into view and put the cursor there.
func (e *Engine) combatCentre(x, y, r, dir int) {
	nx, ny := x+goldbox.StepX[dir], y+goldbox.StepY[dir]
	e.recentre(nx, ny, r)
	e.eraseCursor(x, y)
	if !e.inView(nx, ny) {
		nx = min(max(nx, 0), 0x31)
		ny = min(max(ny, 0), 0x18)
	}
	e.drawCursor(nx, ny)
}

// combatRedrawChar = ovr32:1298 (char, margin, cursor): centre on the
// combatant, with the cursor icon over all its squares when asked.
func (e *Engine) combatRedrawChar(cID core.CharID, r int, cursor byte) {
	e.ccursor = combatCursor{icon: cursor != 0, size: int(e.fact(cID, core.SizeFact{}).(core.Number))}
	at := e.at(cID)
	e.combatCentre(at.X, at.Y, r, 8)
	e.ccursor = freshCursor
}

const pageCells = 21 // the drawing page, 21x21 cells (outside combat)
