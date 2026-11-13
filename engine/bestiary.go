package engine

import "gobox/rules1e"

// Bestiary (extra): the monster kinds the party has met (e.ext.met, kept
// with the save), published in camp (Camp.Bestiary); the kinds are rules1e.BeastOf.

// metIcon is where a met kind's combat icon is: CPIC file and block.
type metIcon struct {
	File  string
	Block int
}

// meetMonster records the kind of a monster loaded for a fight (addMonsters).
func (e *Engine) meetMonster(name, file string, block int) {
	b := rules1e.BeastOf(name)
	if b == nil {
		return
	}
	if e.ext.met == nil {
		e.ext.met = map[string]metIcon{}
	}
	e.ext.met[b.Name] = metIcon{file, block}
}
