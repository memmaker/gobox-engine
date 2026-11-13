package present

import (
	"gobox/front"
	"gobox/game"
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

// expOf is m's experience as the cards and the sheet show it (a signed count: a level drained below 0).
func expOf(m game.Member) int { return int(int32(m.Exp)) }
