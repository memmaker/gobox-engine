package present

import (
	"fmt"
	"image"

	"gobox/front"
	"gobox/game"
)

func cells(r image.Rectangle) front.FrontCellRect {
	return front.FrontCellRect{X1: r.Min.X, Y1: r.Min.Y, X2: r.Max.X, Y2: r.Max.Y}
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
		return sp.Portrait
	}
	own := p.sheetOwn()
	if own.PortraitHeadID == "" && pictured(p.pictures(), cells(sp.PortraitAt).Pixels()) {
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
	if len(c.Fields) > front.FieldSex {
		p.Sex = picked(c.Fields[front.FieldSex])
	}
	for i, k := range []string{"str", "int", "wis", "dex", "con", "cha"} {
		p.Stats[i] = d.Stats[k]
	}
	if p.Stats[0] == 18 && d.Str00 > 0 {
		p.Str00 = fmt.Sprintf("(%02d)", d.Str00%100)
	}
	return p
}
