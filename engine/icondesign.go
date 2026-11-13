package engine

import (
	"gobox/asset"
	"gobox/rules"
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
const ()

// Head, Weapon, Size
// iconEdit is the icon being designed (EditIcon): the character and what
// it had, put back unless kept.
type iconEdit struct {
	who  rules.CharID
	orig rules.IconPics
	anim bool
}

// iconPics is who's own combat icon (core.IconPicsFact).
func (e *Engine) iconPics(who rules.CharID) rules.IconPics {
	return e.fact(who, rules.IconPicsFact{}).(rules.IconPics)
}

// editIcon changes the designed icon by f.
func (e *Engine) editIcon(f func(p *rules.IconPics)) {
	if ed := e.iconEd; ed != nil {
		p := e.iconPics(ed.who)
		f(&p)
		e.do(rules.SetIconPics{Who: ed.who, Pics: p})
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

// iconAmiga is whether the icon is Amiga art: a colour number then shows
// as the terrain palette's slot the part's pixels take for it
// (amigaIconColour: another hue and shade than the EGA colour of that
// number, magenta is black there), not the EGA colour (IconEditor.Looks).
func (e *Engine) iconAmiga() bool {
	if im := e.icons[0xC][0]; im != nil {
		return slices.ContainsFunc(im.Data, func(v byte) bool { return v >= asset.AmigaBase })
	}
	return e.amigaUp() // no combat picture in the slot (the pool composes the icon): the art set says
}

// IconSize toggles the icon designed between small and large.
func (e *Engine) IconSize() {
	e.editIcon(func(p *rules.IconPics) {
		if p.Size == 2 {
			p.Size = 1
		} else {
			p.Size = 2
		}
	})
}

// IconColour gives part's colour shade (0 low, 1 high) colour number v.
func (e *Engine) IconColour(part, shade, v int) {
	if part >= 0 && part < rules.IconParts && shade >= 0 && shade < 2 && v >= 0 && v < 16 {
		e.editIcon(func(p *rules.IconPics) {
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
	e.editIcon(func(p *rules.IconPics) { p.PNG = nil })
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
		e.do(rules.SetIconPics{Who: ed.who, Pics: ed.orig})
	}
	e.iconEd = nil
	e.animInput = ed.anim
	if cr := e.creation; cr != nil && ed.who == cr.c { // the creation's icon: its draft takes the result, the sheet is drawn from it
		icon := e.fact(cr.c, rules.DraftIconFact{}).(rules.Icon)
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
