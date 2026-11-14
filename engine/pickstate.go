package engine

import (
	"slices"

	"gobox/adventure"
	"gobox/asset"
	"gobox/game"
	"gobox/rules"
)

// pickState is the grid picker and the icon designer up, as State holds
// them (game.Picker, game.IconEditor): the front lays them out.
func (e *Engine) pickState() (pk *game.Picker, ed *game.IconEditor) {
	if p := e.picker; p != nil {
		if p.state == nil {
			p.state = e.pickerState(p)
		}
		k := *p.state
		k.Cur = p.cur
		pk = &k
	}
	if d := e.iconEd; d != nil {
		ed = e.iconEditorState(d)
	}
	return pk, ed
}

// pickerState is every picture of p, as PickerView lays out a page of them.
func (e *Engine) pickerState(p *picker) *game.Picker {
	k := &game.Picker{Labels: slices.Clone(p.labels), Ask: p.ask}
	for _, id := range p.chars {
		cl, im, head := "", []byte(nil), ""
		if pt, ok := e.portrait(id); ok {
			cl, im, head = e.word(id, rules.FactClasses), pt.PNG, string(e.poolPics(id).PortraitHead)
		}
		k.Classes, k.Portraits, k.Heads = append(k.Classes, cl), append(k.Portraits, im), append(k.Heads, head)
	}
	for i := range p.n {
		var g game.Picture
		if im, mask := p.pic(i); im != nil && im.Count > 0 {
			g = framePicture(im, 0, mask)
		}
		k.Pics = append(k.Pics, g)
	}
	return k
}

// iconEditorState is the icon designed: its poses, colours and palette.
func (e *Engine) iconEditorState(edit *iconEdit) *game.IconEditor {
	c := e.iconPics(edit.who)
	amiga := e.iconAmiga()
	k := &game.IconEditor{Large: c.Size == 2, Custom: c.PNG != nil, IconPNG: c.PNG}
	for p := range 2 {
		if im := e.icons[0xC][p]; c.PNG == nil && im != nil { // a PNG: the front cuts its poses
			w := im.Width * 8
			k.Poses[p] = game.Picture{W: w, H: len(im.Frame(0)) / w, Pixels: slices.Clone(im.Frame(0)), Palette: im.Palette, Clear: im.Mask}
		}
	}
	for n := range k.Looks {
		k.Looks[n] = byte(n)
		if amiga {
			k.Looks[n] = byte(asset.AmigaBase + 16*asset.AmigaTerrain + amigaIconColour[n])
		}
	}
	for i := range adventure.IconParts {
		k.Colour[i] = [2]int{int(c.Colours[i].Low), int(c.Colours[i].High)}
	}
	return k
}
