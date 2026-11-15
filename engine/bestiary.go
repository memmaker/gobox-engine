package engine

import "gobox/rules"

// Bestiary (extra): the monster kinds the party has met (e.ext.met, kept
// with the save), published in camp (Camp.Bestiary); the kinds are rules.BeastKindFact.

// metIcon is where a met kind's combat icon is: CPIC file and block.
type metIcon struct {
	File  string
	Block int
}

// meetMonster records the kind of a monster loaded for a fight (addMonsters).
func (e *Engine) meetMonster(name, file string, block int) {
	kind := e.fact(rules.NoChar, rules.BeastKindFact{Monster: name}).(rules.BeastKind).Name
	if kind == "" {
		return
	}
	if e.ext.met == nil {
		e.ext.met = map[string]metIcon{}
	}
	e.ext.met[kind] = metIcon{file, block}
}
