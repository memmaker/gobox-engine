// poolpick.go: the portrait and combat icon pickers over the whole shared
// pool. When the game publishes no Picker.Pics for one of them, gobox fills
// its cells from sprites.Names and sends the pick as Create (at creation) or
// SetIcon, the game only storing the opaque pool name.

package ui

import (
	"image"
	"slices"

	"gobox/game"
)

// poolPick is a pool picker up: the Field it picks for, kind and names in the pool.
type poolPick struct {
	field, kind string
	names       []string
	cell        image.Point
	tiles       bool // the bodies by original ref (TilePicker)
}

var poolPicks = map[int]poolPick{
	PickHead:     {"portraithead", "portraits", nil, image.Pt(11, 5), false},
	PickBody:     {"portraitbody", "portraits", nil, image.Pt(11, 6), false},
	PickIconHead: {"iconhead", "combat", nil, image.Pt(11, 5), false},
	PickWeapon:   {"iconbody", "combat", nil, image.Pt(11, 6), false},
}

// poolPrefix is what the names of a field start with.
var poolPrefix = map[string]string{"portraithead": "head_", "portraitbody": "body_", "iconhead": "combat_head_", "iconbody": "combat_body_"}

func (g classicGame) OpenPicker(kind int) bool {
	g.p.pool = nil
	pp, has := poolPicks[kind]
	if !has { // the game's own pictures; a portrait or icon part comes from the shared pool, the game is not asked
		return g.Game.OpenPicker(kind)
	}
	if names, _ := BodyTiles(g.Game.Subject()); kind == PickWeapon && len(names) > 0 {
		pp.names, pp.tiles = names, true
		g.p.pool = &pp
		return true
	}
	pp.list(g.Game.Subject())
	if len(pp.names) == 0 {
		return g.Game.OpenPicker(kind)
	}
	g.p.pool = &pp
	return true
}

// list fills names with the pool's entry IDs of pp's field: art-neutral,
// each drawn in the art shown, so F12 while it is up only redraws. An icon
// part is composed with own's other one, so only the ready poses of its
// body's size class, in pictures of that body's palette, are listed (again
// with every layout: F12 while it is up shows the other art's).
func (pp *poolPick) list(own game.Member) {
	pp.names = poolEntries(pp.kind, poolPrefix[pp.field])
	if pp.kind != "combat" {
		return
	}
	_, body := memberIcon(subject(own))
	if body == "" {
		body = "combat_body_000" // as PickIcon
	}
	class, ready := poolParts()[poolEntryID(body)]
	bp, _ := poolPicture(pp.kind, body).(*image.Paletted)
	pp.names = slices.DeleteFunc(pp.names, func(n string) bool {
		part, ok := poolParts()[n]
		im, _ := poolPicture(pp.kind, n).(*image.Paletted)
		return !ok || !ready || part.large != class.large || im == nil || bp == nil || len(im.Palette) != len(bp.Palette)
	})
}

func (g classicGame) PickerView(page int) *FrontPicker {
	if pp := g.p.pool; pp != nil {
		if pp.tiles {
			return g.Game.TilesOf(page, g.r.grid.X, g.r.grid.Y)
		}
		if was := pp.names; pp.kind == "combat" { // the parts that compose in the art shown now
			if pp.list(g.Game.Subject()); len(pp.names) == 0 {
				pp.names = was
			}
		}
		var v *FrontPicker
		for sc := max(1, g.r.grid.Y/pp.cell.Y); sc >= 1; sc-- { // the largest whole scale that keeps every picture on one page
			k := &game.Picker{Pics: make([]game.Picture, len(pp.names))}
			for i := range k.Pics { // present, no pixels: the cell's size, the PNG fills it
				k.Pics[i].W, k.Pics[i].H = sc*pp.cell.X*8, sc*pp.cell.Y*8
			}
			if v = g.Game.PickerOf(k, page, g.r.grid.X, g.r.grid.Y); v == nil || v.Pages == 1 {
				break
			}
		}
		if v != nil {
			v.Portraits = make([]image.Image, len(pp.names))
			for i, n := range pp.names {
				switch pp.field { // a combat icon shows whole: the picked part over the member's other one
				case "iconhead":
					v.Portraits[i] = PickIcon(g.Game.Subject(), n, "")
				case "iconbody":
					v.Portraits[i] = PickIcon(g.Game.Subject(), "", n)
				default:
					v.Portraits[i] = poolPicture(pp.kind, n)
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
	return true
}

func (g classicGame) ClosePicker() {
	if g.p.pool != nil { // gobox's own: the game has none up
		g.p.pool = nil
		return
	}
	g.Game.ClosePicker()
}
