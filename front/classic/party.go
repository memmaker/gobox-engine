package classic

import (
	"gobox/front"
	"gobox/game"
	"gobox/sprites"
)

// kinds are the actions the menu verbs of game.Entry stand for.
var kinds = map[string]front.ActKind{
	"view": front.ActMember, "items": front.ActInventory, "cast": front.ActSpells, "magic": front.ActMagic,
	"heal": front.ActHeal, "cure": front.ActCure, "memorize": front.ActPlan, "scribe": front.ActScribe,
	"icon": front.ActIconEdit, "drop": front.ActDropMember,
}

func powers(es []game.Entry) []front.Power {
	if es == nil {
		return nil
	}
	out := make([]front.Power, len(es))
	for i, e := range es {
		out[i] = front.Power{Name: e.Name, Kind: kinds[e.Verb]}
	}
	return out
}

// Party is the party's cards s holds (nil: none), with the heads from art
// (Data.Art).
func Party(s game.State, art map[string]game.Picture) []front.Member {
	var out []front.Member
	for _, m := range s.Party {
		if m.Dropped == nil && m.PortraitHeadID != "" { // the card shows the face: the head alone
			m.Dropped = sprites.Image("pool", "portraits", m.PortraitHeadID)
		}
		out = append(out, front.Member{Name: m.Name, HP: m.HP, MaxHP: m.MaxHP, AC: m.AC, Status: m.Status, Class: m.Class,
			Selected: m.Chosen, Colour: m.Colour, HeadPix: pix(art[m.Head]), Portrait: m.Dropped, Powers: powers(m.Card),
			Exp: int(int32(m.Exp)), LevelExp: m.LevelExp, NextExp: m.NextExp, Spells: m.SpellsReady, Memo: m.Memo})
	}
	return out
}

// Purse is the party's money and party items s holds (nil: no party).
func Purse(s game.State) *front.Purse {
	p := s.Purse
	if p == nil {
		return nil
	}
	return &front.Purse{Gold: p[game.Gold], Copper: p[game.Copper], Gems: p[game.Gems], Jewelry: p[game.Jewelry], Items: s.PoolItems, Slots: s.PartyCap}
}
