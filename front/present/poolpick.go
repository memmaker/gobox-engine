// poolpick.go: the portrait and combat icon pickers over the whole shared
// pool. When the game publishes no Picker.Pics for one of them, gobox fills
// its cells from sprites.Names and sends the pick as Create (at creation) or
// SetIcon, the game only storing the opaque pool name.

package present

import (
	"image"

	"gobox/front"
	"gobox/front/classic"
	"gobox/game"
	"gobox/sprites"
)

// poolPick is a pool picker up: the Field it picks for, kind and names in the pool.
type poolPick struct {
	field, kind string
	names       []string
	cell        image.Point
	tiles       bool // the bodies by original ref (classic.TilePicker)
}

var poolPicks = map[int]poolPick{
	front.PickHead:     {"portraithead", "portraits", nil, image.Pt(11, 5), false},
	front.PickBody:     {"portraitbody", "portraits", nil, image.Pt(11, 6), false},
	front.PickIconHead: {"iconhead", "combat", nil, image.Pt(11, 5), false},
	front.PickWeapon:   {"iconbody", "combat", nil, image.Pt(11, 6), false},
}

// poolPrefix is what the names of a field start with.
var poolPrefix = map[string]string{"portraithead": "head_", "portraitbody": "body_", "iconhead": "combat_head_", "iconbody": "combat_body_"}

func (g classicGame) OpenPicker(kind int) bool {
	g.p.pool = nil
	ok := g.Game.OpenPicker(kind)
	if k := g.r.State().Picker; k != nil && len(k.Pics) > 0 { // the game's own pictures
		return ok
	}
	pp, has := poolPicks[kind]
	if !has {
		return ok
	}
	if names, _ := classic.BodyTiles(g.r.State()); kind == front.PickWeapon && len(names) > 0 {
		pp.names, pp.tiles = names, true
		g.p.pool = &pp
		return true
	}
	for _, n := range sprites.Names(series, pp.kind) {
		if len(n) > len(poolPrefix[pp.field]) && n[:len(poolPrefix[pp.field])] == poolPrefix[pp.field] {
			pp.names = append(pp.names, n)
		}
	}
	if len(pp.names) == 0 {
		return ok
	}
	g.p.pool = &pp
	return true
}

func (g classicGame) PickerView(page int) *front.Picker {
	if pp := g.p.pool; pp != nil {
		s := g.r.State()
		if pp.tiles {
			return classic.TilePicker(s, page, g.p.grid[0], g.p.grid[1])
		}
		s.Picker = &game.Picker{Cell: pp.cell, Pics: make([]game.Picture, len(pp.names))}
		for i := range s.Picker.Pics {
			s.Picker.Pics[i].W = 1 // present, no pixels: classic.Picker lays the cell out, the PNG fills it
		}
		v := classic.Picker(s, page, g.p.grid[0], g.p.grid[1])
		if v != nil {
			v.Portraits = make([]image.Image, len(pp.names))
			for i, n := range pp.names {
				v.Portraits[i] = sprites.Image(series, pp.kind, n)
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
	s := g.r.State()
	var d game.Decision = game.SetIcon{Member: s.Selected, Field: pp.field, ID: pp.names[i]}
	if classic.Drafting(s) {
		d = game.Create{Field: pp.field, Text: pp.names[i]}
	}
	if g.r.Apply(d).Err != nil {
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
