package pool

import (
	"gobox/core"
	"gobox/ega"
	"slices"
)

// The combat icon designer (extra), in place of the original's icon
// editor states (ovr17:382D: Parts/color-1/color-2/Size, then Next/Prev
// stepping): one page with both poses large and live, Head and Weapon
// opening grids of all the parts, Size toggling, and the six parts' two
// colours each as swatches that a palette sets: the 16 EGA colours the DOS
// editor stepped through, or on Amiga art the 15 terrain palette slots the
// Amiga's stepped through (1..15, in its order). It takes decisions: the
// page's parts click as OpenPicker (PickIconHead, PickWeapon), IconSize
// and IconColour; EndIcon(true) takes it, EndIcon(false) puts the old icon back. The
// bar is the presenter's.

// Layout, in cells: the poses 9 cells square at the 40x25 page, a whole
// multiple of that as far as the grid has room (iconScale); the rest
// moves along with them.
const (
	idPreviewRow  = 2 // the poses; ready, action 9s+2 to the right
	idPreviewCols = 2
	idPartCol     = 2  // the six parts, a row each, under the poses
	idSwatchCol   = 12 // colour 1; colour 2 three to the right
	idPaletteCol  = 2  // the 16 colours, 8 a row, 2 cells each, under the parts
)

// Head, Weapon, Size
// IconParts are the parts in IconColour order (the hair and the face
// share a part).
var IconParts = [6]string{"Body", "Arm", "Leg", "Hair/Face", "Shield", "Weapon"}

// iconEdit is the icon being designed (EditIcon): the character and what
// it had, put back unless kept.
type iconEdit struct {
	who  core.CharID
	orig core.IconPics
	anim bool
}

// iconPics is who's own combat icon (core.IconPicsFact).
func (e *Engine) iconPics(who core.CharID) core.IconPics {
	return e.fact(who, core.IconPicsFact{}).(core.IconPics)
}

// editIcon changes the designed icon by f.
func (e *Engine) editIcon(f func(p *core.IconPics)) {
	if ed := e.iconEd; ed != nil {
		p := e.iconPics(ed.who)
		f(&p)
		e.do(core.SetIconPics{Who: ed.who, Pics: p})
	}
}

// EditIcon readies the designer on the selected character (the creation,
// camp's Icon, member n's: n < 0 the selected one): the presenter's screen
// shows IconEditorView until EndIcon.
func (e *Engine) EditIcon(n int) {
	if e.iconEd != nil || e.partySize() == 0 {
		return
	}
	who := e.actor(n)
	e.newPage()
	e.combatViewLeave() // ovr32:00E3: the combat palette, black shown dark grey so the icon's black shows
	e.iconEd = &iconEdit{who: who, orig: e.iconPics(who), anim: e.animInput}
	e.animInput = false // camp's fire stays off the page while the editor is up
}

// iconCells are the colour number in each palette cell, row-major
// (noCell: empty), and whether the icon is Amiga art. The swatches show
// what a click gives: on Amiga art that is the terrain palette's slot the
// part's pixels take for the colour number (amigaIconColour: another hue
// and shade than the EGA colour of that number, magenta is black there),
// not the EGA colour. The DOS palette's columns put the darker of each
// pair on top; the Amiga's is its 15 slots in slot order, as its Next
// stepped them (slot 0 is left out).
func (e *Engine) iconCells() (cells [16]byte, amiga bool) {
	im := e.icons[0xC][0]
	amiga = !DOSArt.Load() // no combat picture in the slot (the pool composes the icon): the art set says
	if im != nil {
		amiga = slices.ContainsFunc(im.Data, func(v byte) bool { return v >= ega.AmigaBase })
	}
	if amiga {
		for n, s := range amigaIconColour {
			if s > 0 {
				cells[s-1] = byte(n)
			}
		}
		cells[15] = noCell
		return cells, true
	}
	lum := func(i byte) int {
		c := ega.Identity.RGBA(i)
		return int(c.R) + int(c.G) + int(c.B)
	}
	for k := 0; k < 8; k++ {
		a, b := byte(k), byte(k+8)
		if lum(a) > lum(b) {
			a, b = b, a
		}
		cells[k], cells[k+8] = a, b
	}
	return cells, false
}

const noCell = 0xFF

// IconSize toggles the icon designed between small and large.
func (e *Engine) IconSize() {
	e.editIcon(func(p *core.IconPics) {
		if p.Size == 2 {
			p.Size = 1
		} else {
			p.Size = 2
		}
	})
}

// IconColour gives part's colour shade (0 low, 1 high) colour number v.
func (e *Engine) IconColour(part, shade, v int) {
	if part >= 0 && part < core.IconParts && shade >= 0 && shade < 2 && v >= 0 && v < 16 {
		e.editIcon(func(p *core.IconPics) {
			if shade == 0 {
				p.Colours[part].Low = uint8(v)
			} else {
				p.Colours[part].High = uint8(v)
			}
		})
	}
}

// IconRemove drops the PNG icon: the game's own icon again (the save
// drops NAME.icon.png).
func (e *Engine) IconRemove() {
	e.editIcon(func(p *core.IconPics) { p.PNG = nil })
}

// EndIcon ends the designer: keep takes the icon, else the old one is
// back; the page under draws itself again.
func (e *Engine) EndIcon(keep bool) {
	ed := e.iconEd
	if ed == nil {
		return
	}
	if keep {
		e.Music(3) // Amiga 0x2F33E
	} else {
		e.do(core.SetIconPics{Who: ed.who, Pics: ed.orig})
	}
	e.iconEd = nil
	e.animInput = ed.anim
	if cr := e.creation; cr != nil && ed.who == cr.c { // the creation's icon: its draft takes the result, the sheet is drawn from it
		icon := e.fact(cr.c, core.DraftIconFact{}).(core.Icon)
		e.creationEdit(func(cr *creation) error { cr.d.D.Icon = icon; return nil })
	}
	e.clearPrompt()
	e.freeIcon(0xC) // ovr33:014E
	if e.viewMode > 1 {
		e.loadAreaIcon(0, 0xB)
		e.loadAreaIcon(1, 0xC)
	}
	e.combatViewEnter() // ovr32:0105
	switch {
	case e.campOn():
		e.DrawMainScreen()
	default:
		e.redrawNow()
	}
}

// IconPNG makes the PNG at path the icon designed, or says why not.
func (e *Engine) IconPNG(path string) {
	b, ok := e.droppedPNG(path)
	if !ok {
		return
	}
	if ed := e.iconEd; ed != nil {
		e.takeIconPNG(ed.who, b)
	} else if f := e.snap.iconDrop; f != nil { // the creation sheet takes the dropped PNG too
		f(b)
	}
}
