// poolpick.go: the portrait and combat icon pickers over the whole shared
// pool. When the game publishes no Picker.Pics for one of them, gobox fills
// its cells from sprites.Names and sends the pick as Create (at creation) or
// SetIcon, the game only storing the opaque pool name.

package present

import (
	"image"

	"gobox/front"
	"gobox/game"
	"gobox/sprites"
)

// poolPick is a pool picker up: the Field it picks for, kind and names in the pool.
type poolPick struct {
	field, kind string
	names       []string
	cell        image.Point
	tiles       bool // the bodies by original ref (TilePicker)
	amiga       bool // names lists the Amiga set (F12 when it was listed)
}

var poolPicks = map[int]poolPick{
	front.FrontPickHead:     {"portraithead", "portraits", nil, image.Pt(11, 5), false, false},
	front.FrontPickBody:     {"portraitbody", "portraits", nil, image.Pt(11, 6), false, false},
	front.FrontPickIconHead: {"iconhead", "combat", nil, image.Pt(11, 5), false, false},
	front.FrontPickWeapon:   {"iconbody", "combat", nil, image.Pt(11, 6), false, false},
}

// poolPrefix is what the names of a field start with.
var poolPrefix = map[string]string{"portraithead": "head_", "portraitbody": "body_", "iconhead": "combat_head_", "iconbody": "combat_body_"}

func (g classicGame) OpenPicker(kind int) bool {
	g.p.pool = nil
	ok := g.Game.OpenPicker(kind)
	pp, has := poolPicks[kind]
	if !has { // the game's own pictures; a portrait or icon part comes from the shared pool, whatever the game offers
		return ok
	}
	if names, _ := BodyTiles(g.Game.Subject()); kind == front.FrontPickWeapon && len(names) > 0 {
		pp.names, pp.tiles = names, true
		g.p.pool = &pp
		return true
	}
	pp.list()
	if len(pp.names) == 0 {
		return ok
	}
	g.p.pool = &pp
	return true
}

// list fills names with the pool's pictures of pp's field in the art shown:
// F12's Amiga set alone, never mixed with the DOS one.
func (pp *poolPick) list() {
	prefix := poolPrefix[pp.field]
	if pp.amiga = amigaShown(); pp.amiga {
		prefix = "amiga_" + prefix
	}
	pp.names = pp.names[:0]
	for _, n := range sprites.Names(series, pp.kind) {
		if len(n) > len(prefix) && n[:len(prefix)] == prefix {
			pp.names = append(pp.names, n)
		}
	}
}

func (g classicGame) PickerView(page int) *front.FrontPicker {
	if pp := g.p.pool; pp != nil {
		if !pp.tiles && pp.amiga != amigaShown() { // F12 while it is up: the other set's list
			pp.list()
		}
		if pp.tiles {
			return g.Game.TilesOf(page, g.r.grid.X, g.r.grid.Y)
		}
		k := &game.Picker{Cell: pp.cell, Pics: make([]game.Picture, len(pp.names))}
		for i := range k.Pics {
			k.Pics[i].W = 1 // present, no pixels: Picker lays the cell out, the PNG fills it
		}
		v := g.Game.PickerOf(k, page, g.r.grid.X, g.r.grid.Y)
		if v != nil {
			v.Portraits = make([]image.Image, len(pp.names))
			for i, n := range pp.names {
				switch pp.field { // a combat icon shows whole: the picked part over the member's other one
				case "iconhead":
					v.Portraits[i] = PickIcon(g.Game.Subject(), n, "")
				case "iconbody":
					v.Portraits[i] = PickIcon(g.Game.Subject(), "", n)
				default:
					v.Portraits[i] = sprites.Image(series, pp.kind, n)
				}
			}
		}
		return v
	}
	return g.pickerView(page)
}

func (g classicGame) Pick(i int) bool {
	pp := g.p.pool
	if pp == nil {
		return g.Game.Pick(i)
	}
	if i < 0 || i >= len(pp.names) {
		return false
	}
	if g.Game.SetChoice(pp.field, pp.names[i]) != nil {
		return false
	}
	g.p.pool = nil
	g.Game.ClosePicker()
	return true
}

func (g classicGame) ClosePicker() {
	g.p.pool = nil
	g.Game.ClosePicker()
}
