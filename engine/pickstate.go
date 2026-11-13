package engine

import (
	"image"
	"slices"

	"gobox/asset"
	"gobox/game"
	"gobox/rules"
)

// pickState is the grid picker and the icon designer up, as State holds
// them (game.Picker, game.IconEditor): the front lays them out.
func (e *Engine) pickState(s *facts) {
	if p := e.picker; p != nil {
		if p.state == nil || p.stateDOS != DOSArt.Load() { // the pictures follow the art set
			p.state, p.stateDOS = e.pickerState(p), DOSArt.Load()
		}
		k := *p.state
		k.Cur = p.cur
		s.Picker = &k
	}
	if ed := e.iconEd; ed != nil {
		s.IconEditor = e.iconEditorState(ed)
	}
}

// pickerState is every picture of p, as PickerView lays out a page of them.
func (e *Engine) pickerState(p *picker) *game.Picker {
	k := &game.Picker{Cell: image.Pt(p.cw, p.ch), Labels: slices.Clone(p.labels), Ask: p.ask}
	for _, id := range p.chars {
		cl, im := "", image.Image(nil)
		if pt, ok := e.portrait(id); ok {
			cl, im = e.word(id, rules.FactClasses), e.pngImage(pt.PNG)
		}
		k.Classes, k.Portraits = append(k.Classes, cl), append(k.Portraits, im)
	}
	for i := range p.n {
		var g game.Picture
		if im, mask := p.pic(i); im != nil && im.Count > 0 {
			lp := e.pictureOf(0, 0, im, 0, mask)
			g = gamePicture(lp)
			g.File, g.Block, g.Frame = lp.File, lp.Block, lp.Frame
		}
		k.Pics = append(k.Pics, g)
	}
	return k
}

// iconEditorState is the icon designed: its poses, colours and palette.
func (e *Engine) iconEditorState(edit *iconEdit) *game.IconEditor {
	c := e.iconPics(edit.who)
	cells, amiga := e.iconCells()
	k := &game.IconEditor{Large: c.Size == 2, Custom: c.PNG != nil}
	for p := range 2 {
		if c.PNG != nil {
			k.Icons[p] = iconPose(e.pngImage(c.PNG), p, false)
		} else if im := e.icons[0xC][p]; im != nil {
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
	for i, v := range cells {
		k.Cells[i] = int(v)
		if v == noCell {
			k.Cells[i] = -1
		}
	}
	for i := range IconParts {
		k.Colour[i] = [2]int{int(c.Colours[i].Low), int(c.Colours[i].High)}
	}
	return k
}
