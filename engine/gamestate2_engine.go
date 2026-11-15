package engine

import (
	"image"
	"strconv"

	"gobox/game"
	"gobox/rules"
)

// headArt is the Data.Art name of the head c chose ("" none).
func (e *Engine) headArt(cID rules.CharID) string {
	if e.charHead(cID).Pix == nil {
		return ""
	}
	return "head/" + strconv.Itoa(e.fact(cID, rules.PortraitFact{}).(rules.Portrait).Head)
}

var verbs = map[PowerKind]string{
	PowerView: "view", PowerItems: "items", PowerSpells: "cast", PowerMagic: "magic",
	PowerHeal: "heal", PowerCure: "cure", PowerPlan: "memorize", PowerScribe: "scribe",
	PowerIcon: "icon", PowerDrop: "drop",
}

func gameEntries(ps []Power) []game.Entry {
	if ps == nil {
		return nil
	}
	out := make([]game.Entry, len(ps))
	for i, p := range ps {
		out[i] = game.Entry{Name: p.Name, Verb: verbs[p.Kind]}
	}
	return out
}

func gameRect(r CellRect) image.Rectangle { return image.Rect(r.X1, r.Y1, r.X2, r.Y2) }

// gameSheetPage is p as a game.SheetPage, nil for none.
func gameSheetPage(p *SheetPage) *game.SheetPage {
	if p != nil {
		k := &game.SheetPage{Name: p.Name, NPC: p.NPC, Spells: p.Spells, Sex: p.Sex, Race: p.Race, Age: p.Age,
			Alignment: p.Alignment, Class: p.Class, Stats: p.Stats, Str00: p.Str00, Creating: p.Creating, Levels: p.Levels,
			Exp: p.Exp, Next: p.Next, AC: p.AC, RearAC: p.RearAC, HP: p.HP, MaxHP: p.MaxHP, THAC0: p.THAC0, Damage: p.Damage, Move: p.Move,
			Saves: p.Saves, Mods: p.Mods, Status: p.Status, Weapon: p.Weapon, Armour: p.Armour, Encumbrance: p.Encumbrance,
			Coins: p.Coins, Effects: p.Effects, DualClass: p.DualClass, HasPortrait: p.PortraitAt != (CellRect{}), Portrait: p.Portrait, IconPNG: p.IconPNG}
		for _, r := range p.Rules {
			k.Rules = append(k.Rules, r)
		}
		return k
	}
	return nil
}
