package ui

import (
	"gobox/game"
)

// kinds are the actions the menu verbs of game.Entry stand for.
var kinds = map[string]ActKind{
	"view": ActMember, "items": ActInventory, "cast": ActSpells, "magic": ActMagic,
	"heal": ActHeal, "cure": ActCure, "memorize": ActPlan, "scribe": ActScribe,
	"icon": ActIconEdit, "drop": ActDropMember,
}

func powers(es []game.Entry) []Power {
	if es == nil {
		return nil
	}
	out := make([]Power, len(es))
	for i, e := range es {
		out[i] = Power{Name: e.Name, Kind: kinds[e.Verb]}
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
