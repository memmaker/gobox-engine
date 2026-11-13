package ui

import (
	"fmt"
	"image"

	"gobox/game"
)

func cells(r image.Rectangle) CellRect {
	return CellRect{X1: r.Min.X, Y1: r.Min.Y, X2: r.Max.X, Y2: r.Max.Y}
}

// sheetPortraitBox is the sheet's portrait box: 11x11 cells against the
// right border at the top (the head 5 rows, the body 6), as the original's
// SheetPortrait (ovr29:04F9); zero when the member has no portrait.
func sheetPortraitBox(sp *game.SheetPage) CellRect {
	if sp == nil || !sp.HasPortrait {
		return CellRect{}
	}
	return CellRect{X1: 28, Y1: 1, X2: 38, Y2: 11}
}

// sheetOwn is the member whose sheet is up: the draft being made, else the selected one.
func (p *Presenter) sheetOwn() game.Member {
	s, r := p.state(), p.eng.r
	return OwnSubject(r.modeOf(s), creationOf(s), partyOf(s), r.selectedOf(s))
}

// sheetDraft is the draft the creation sheet shows (nil: a member's sheet).
func (p *Presenter) sheetDraft() *game.Creation {
	s := p.state()
	if c := creationOf(s); Drafting(p.eng.r.modeOf(s), c) {
		return c
	}
	return nil
}

// sheetPortrait is the PNG dropped on the sheet, else the pool's head and
// body of its member in an empty portrait box (v's pictures are the game's:
// a default fills an empty box only).
func (p *Presenter) sheetPortrait(sp *game.SheetPage) image.Image {
	if sp == nil || sp.Portrait != nil {
		if sp == nil {
			return nil
		}
		return ownPNG(sp.Portrait)
	}
	own := p.sheetOwn()
	if own.PortraitHeadID == "" && pictured(p.pictures(), sheetPortraitBox(sp).Pixels()) {
		return nil
	}
	m := subject(own)
	return poolID("portraits", m.PortraitHeadID, m.PortraitBodyID)
}

// draftPage is the creation sheet's page from c's draft (the rules publish
// no page while creating).
func draftPage(c *game.Creation, rules *game.SheetPage) *game.SheetPage {
	d := c.Draft
	p := &game.SheetPage{Name: d.Name, Race: d.Race, Alignment: d.Alignment, Class: d.Class, Creating: true,
		AC: d.AC, RearAC: d.AC, HP: d.HP, MaxHP: d.MaxHP, THAC0: d.THAC0, Move: d.Move, Status: d.Status, Age: d.Age}
	if r := rules; r != nil { // the rules' own figures: one source, never a 0 the game did figure
		p.Age, p.AC, p.RearAC, p.HP, p.MaxHP, p.THAC0, p.Damage, p.Move = r.Age, r.AC, r.RearAC, r.HP, r.MaxHP, r.THAC0, r.Damage, r.Move
		p.Saves, p.Mods, p.Rules, p.Levels, p.Exp, p.Next, p.Portrait = r.Saves, r.Mods, r.Rules, r.Levels, r.Exp, r.Next, r.Portrait
	}
	if d.ClassText != "" {
		p.Class = d.ClassText
	}
	if len(c.Fields) > FieldSex {
		p.Sex = picked(c.Fields[FieldSex])
	}
	for i, k := range []string{"str", "int", "wis", "dex", "con", "cha"} {
		p.Stats[i] = d.Stats[k]
	}
	if p.Stats[0] == 18 && d.Str00 > 0 {
		p.Str00 = fmt.Sprintf("(%02d)", d.Str00%100)
	}
	return p
}
