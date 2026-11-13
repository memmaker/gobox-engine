package present

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
func Sheet(v front.View, h *game.Sheet, portrait, iconBox image.Rectangle, page *game.SheetPage, drafted bool, draft *game.Creation, own game.Member) front.View {
	v.Sheet, v.SheetPage = nil, nil
	if h != nil {
		v.Sheet = &front.Sheet{Powers: powers(h.Powers)}
		for _, r := range h.Reactions {
			v.Sheet.Reactions = append(v.Sheet.Reactions, front.Reaction{ID: r.ID, Name: r.Name, Policy: front.Policy(r.Policy)})
		}
	}
	v.Portrait, v.IconBox = cells(portrait), cells(iconBox)
	p := page
	if drafted { // the creation sheet: gobox lays it out from the draft, with the rules' figures when they give a page
		p = draftPage(draft, page)
	}
	if p != nil {
		k := &front.SheetPage{Name: p.Name, NameColour: nameColour(own), NPC: p.NPC, Spells: p.Spells, Sex: p.Sex, Race: p.Race, Age: p.Age,
			Alignment: p.Alignment, Class: p.Class, Stats: p.Stats, Str00: p.Str00, Creating: p.Creating, Levels: p.Levels,
			Exp: p.Exp, Next: p.Next, AC: p.AC, RearAC: p.RearAC, HP: p.HP, MaxHP: p.MaxHP, THAC0: p.THAC0, Damage: p.Damage, Move: p.Move,
			Saves: p.Saves, Mods: p.Mods, Status: p.Status, Weapon: p.Weapon, Armour: p.Armour, Encumbrance: p.Encumbrance,
			Coins: p.Coins, Effects: p.Effects, PortraitAt: cells(p.PortraitAt), Portrait: p.Portrait, Icons: p.Icons}
		for _, r := range p.Rules {
			k.Rules = append(k.Rules, front.Stat(r))
		}
		if k.Portrait == nil && (own.PortraitHeadID != "" || !pictured(v, k.PortraitAt.Pixels())) { // a default fills an empty box only
			m := subject(own)
			k.Portrait = poolID("portraits", m.PortraitHeadID, m.PortraitBodyID)
		}
		if drafted {
			k.NameColour = 11 // the member being made: as a well one's
			draftRows(k, draft)
		}
		v.SheetPage = k
	}
	return v
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

// draftRows puts c's own fields, abilities and figures on k as given:
// adnd2-br the fields one per row, the figures in order.
func draftRows(k *front.SheetPage, c *game.Creation) {
	if rowFields(c) {
		k.FieldRows = []string{}
		for _, f := range c.Fields {
			k.FieldRows = append(k.FieldRows, picked(f))
		}
	}
	for _, st := range c.Stats {
		k.StatRows = append(k.StatRows, front.Stat{Name: st.Label, Value: fmt.Sprintf("%2d", st.Value) + st.Extra})
	}
	if c.Derived != nil {
		k.Derived = []front.Stat{}
		for _, d := range c.Derived {
			k.Derived = append(k.Derived, front.Stat{Name: d.Label, Value: d.Value})
		}
	}
}
