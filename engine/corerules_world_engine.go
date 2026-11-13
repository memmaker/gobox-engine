package engine

import "gobox/rules"

// partyIDs is the party as the ruleset's ids, in marching order.
func (e *Engine) partyIDs() rules.Party {
	return e.fact(rules.NoChar, rules.PartyFact{}).(rules.Party)
}

// trainingOpen: the training hall is open (core.HallOpen).
func (e *Engine) trainingOpen() bool { open, _ := e.hallOpen(); return open }

// moveMember puts member from at place to (camp's order).
func (e *Engine) moveMember(from, to int) error {
	id := rules.NoChar
	if p := e.partyIDs(); from >= 0 && from < len(p) {
		id = p[from]
	}
	_, err := e.ruleset().Resolve(nil, rules.MoveMember{Who: id, To: int32(to)}, nil)
	return err
}
