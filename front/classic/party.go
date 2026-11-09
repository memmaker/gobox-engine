package classic

import (
	"cmp"
	"gobox/front"
	"gobox/game"
	"image"
	"slices"
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

// Party is the party's cards (nil: none).
func Party(party []game.Member, cards []memberCard, sel int) []front.Member {
	var out []front.Member
	marked := slices.ContainsFunc(cards, func(c memberCard) bool { return c.Chosen })
	for i, m := range party {
		m = defaultPortrait(m)
		var c memberCard
		if i < len(cards) {
			c = cards[i]
		}
		var head image.Image
		if m.PortraitHeadID != "" { // the card shows the face: the head alone, its lit part filling the square (present's face), not the whole wide head fitted
			head = poolImage("portraits", m.PortraitHeadID)
		}
		if head == nil { // no portrait: the combat icon (Buck Rogers' members have only that), else the card names the member
			head = poolImage("combat", cmp.Or(m.IconBodyID, m.IconHeadID))
		}
		out = append(out, front.Member{Head: head, Name: m.Name, HP: m.HP, MaxHP: m.MaxHP, AC: m.AC, Status: m.Status, Class: m.Class,
			Selected: c.Chosen || !marked && i == sel, Colour: nameColour(m), Powers: powers(c.Menu),
			Exp: int(int32(m.Exp)), LevelExp: m.LevelExp, NextExp: m.NextExp, Spells: m.SpellsReady, Memo: m.Memo})
	}
	return out
}

// Purse is the party's money and party items (nil: no party).
func Purse(p map[string]int, items, slots int) *front.Purse {
	if p == nil {
		return nil
	}
	return &front.Purse{Gold: p[Coin], Copper: p[game.Copper], Gems: p[game.Gems], Jewelry: p[game.Jewelry], Items: items, Slots: slots}
}

// nameColour is the roster colour of m's name (EGA number) when not
// selected: red out of action, yellow on the monsters' side, magenta
// ready for a level, else cyan.
func nameColour(m game.Member) int {
	switch {
	case !m.Okay:
		return 12
	case m.Foe:
		return 14
	case m.NextExp > 0 && int(m.Exp) >= m.NextExp:
		return 13
	}
	return 11
}
