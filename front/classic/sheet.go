package classic

import (
	"fmt"
	"image"

	"gobox/front"
	"gobox/game"
)

func cells(r image.Rectangle) front.CellRect {
	return front.CellRect{X1: r.Min.X, Y1: r.Min.Y, X2: r.Max.X, Y2: r.Max.Y}
}

// Sheet is v with the sheet s holds: what it offers (nil: none), the page
// (nil: none) and the portrait and icon boxes.
func Sheet(v front.View, s game.State) front.View {
	v.Sheet, v.SheetPage = nil, nil
	if h := s.Sheet; h != nil {
		v.Sheet = &front.Sheet{Powers: powers(h.Powers)}
		for _, r := range h.Reactions {
			v.Sheet.Reactions = append(v.Sheet.Reactions, front.Reaction{ID: r.ID, Name: r.Name, Policy: front.Policy(r.Policy)})
		}
	}
	v.Portrait, v.IconBox = cells(s.Portrait), cells(s.IconBox)
	p := s.SheetPage
	if p == nil && s.Creation != nil && s.Mode == game.ModeCreate { // the creation sheet: gobox lays it out from the draft
		p = draftPage(s.Creation)
	}
	if p != nil {
		k := &front.SheetPage{Name: p.Name, NameColour: p.NameColour, NPC: p.NPC, Spells: p.Spells, Sex: p.Sex, Race: p.Race, Age: p.Age,
			Alignment: p.Alignment, Class: p.Class, Stats: p.Stats, Str00: p.Str00, Creating: p.Creating, Levels: p.Levels,
			Exp: p.Exp, Next: p.Next, AC: p.AC, HP: p.HP, MaxHP: p.MaxHP, THAC0: p.THAC0, Damage: p.Damage, Move: p.Move,
			Saves: p.Saves, Mods: p.Mods, Status: p.Status, Weapon: p.Weapon, Armour: p.Armour, Encumbrance: p.Encumbrance,
			Coins: p.Coins, Effects: p.Effects, PortraitAt: cells(p.PortraitAt), Portrait: p.Portrait, Icons: p.Icons}
		for _, r := range p.Rules {
			k.Rules = append(k.Rules, front.Stat(r))
		}
		if k.Portrait == nil {
			m := subject(s)
			k.Portrait = poolID("portraits", m.PortraitHeadID, m.PortraitBodyID)
		}
		v.SheetPage = k
	}
	return v
}

// draftPage is the creation sheet's page from c's draft (the rules publish
// no page while creating).
func draftPage(c *game.Creation) *game.SheetPage {
	d := c.Draft
	p := &game.SheetPage{Name: d.Name, Race: d.Race, Alignment: d.Alignment, Class: d.Class, Creating: true,
		AC: d.AC, HP: d.HP, MaxHP: d.MaxHP, THAC0: d.THAC0, Move: d.Move, Status: d.Status}
	if d.ClassText != "" {
		p.Class = d.ClassText
	}
	for _, f := range c.Fields {
		if f.Name == "gender" && f.Pick >= 0 && f.Pick < len(f.Options) {
			p.Sex = f.Options[f.Pick]
		}
	}
	for i, k := range []string{"str", "int", "wis", "dex", "con", "cha"} {
		p.Stats[i] = d.Stats[k]
	}
	if p.Stats[0] == 18 && d.Str00 > 0 {
		p.Str00 = fmt.Sprintf("(%02d)", d.Str00%100)
	}
	return p
}
