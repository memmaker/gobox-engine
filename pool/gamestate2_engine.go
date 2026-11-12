package pool

import (
	"image"
	"strconv"

	"gobox/core"
	"gobox/game"
)

// headArt is the Data.Art name of the head c chose ("" none).
func (e *Engine) headArt(cID core.CharID) string {
	if e.charHead(cID).Pix == nil {
		return ""
	}
	return "head/" + strconv.Itoa(e.fact(cID, core.PortraitFact{}).(core.Portrait).Head)
}

// fillState2 adds to s the facts of the window, the journal, what the game
// waits on, the party's cards and the sheet, from the engine; members and
// quests are its cards and journal, pick the power pick owed.
func (e *Engine) fillState2(s *facts, members []Member, quests []Quest, pick *PowerPick) {
	for i := range s.Quests {
		q := quests[i]
		s.Quests[i].Giver, s.Quests[i].Place, s.Quests[i].Brief = q.Giver, q.Place, q.Text
	}
	for _, l := range e.loreView() {
		s.Lore = append(s.Lore, game.Lore{Title: l.Title, Place: l.Place, Text: l.Text, ID: l.ID})
	}
	if s.Door != nil {
		d := e.ext.door.opts
		s.Door.Bash, s.Door.Pick, s.Door.Knock = d.Bash, d.Pick, d.Knock
	}
	if p := pick; p != nil {
		k := &game.PowerPick{Name: p.Name, What: p.What}
		for _, m := range p.Options {
			k.Options = append(k.Options, game.PowerOption{ID: int(m.ID), Name: m.Name, Tip: m.Tip, Aim: int(m.Aim), Range: m.Range, Area: m.Area, Move: m.Move})
		}
		s.PowerPick = k
	}
	chars := e.partySide()
	for i, m := range members {
		g := &s.Party[i]

		g.LevelExp, g.NextExp, g.SpellsReady, g.Memo = m.LevelExp, m.NextExp, m.Spells, m.Memo
		if i < len(chars) {
			g.Page = gameSheetPage(e.sheetPage(chars[i])) // BUG-CLASSES 36: gobox's look-only sheet (a fight's, a shop's View)
		}
	}
	s.PoolItems = int(e.ruleset().Check(core.PoolCount{}, nil).Value)
	var sh *Sheet
	if e.sheetUp() {
		sh = e.sheetView()
	}
	if sh != nil {
		k := &game.Sheet{Powers: gameEntries(sh.Powers)}
		for _, r := range sh.Reactions {
			k.Reactions = append(k.Reactions, game.Reaction{ID: r.ID, Name: r.Name, Policy: int(r.Policy)})
		}
		s.Sheet = k
	}
	s.SheetPage = gameSheetPage(e.snap.sheet)
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
			Coins: p.Coins, Effects: p.Effects, PortraitAt: gameRect(p.PortraitAt), Portrait: p.Portrait, Icons: p.Icons}
		for _, r := range p.Rules {
			k.Rules = append(k.Rules, r)
		}
		return k
	}
	return nil
}
