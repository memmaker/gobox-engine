package engine

import (
	"gobox/core"
	"slices"
)

// NewClass = the hall's New Class (front.Game): member n (< 0: the selected
// one) changes to por class number class (< 0: not picked yet, asked).
func (e *Engine) NewClass(n, class int) error {
	return e.newClass(n, e.fact(core.NoChar, core.ClassByNumber{N: class}).(core.Class))
}

// newClass is NewClass to class (ClassNone: asked, askClass).
func (e *Engine) newClass(n int, class core.Class) error {
	if !e.inParty() || !slices.Contains(e.partyFunctions(), partyNewClass) {
		return core.ErrNotNow
	}
	id := e.actor(n)
	offers := e.classOffers(id)
	if len(offers) == 0 {
		e.say(e.name(id) + " doesn't qualify.")
		return core.ErrNotNow
	}
	for _, o := range offers {
		if o.Action.(core.ChangeClass).To == class {
			_, err := e.ruleset().Resolve(core.Training{Party: e.partyIDs(), DualClass: true}, o.Action, nil)
			return err
		}
	}
	return e.askClass(offers, n)
}

// classOffers are the rules' ChangeClass offers for member id in a hall.
func (e *Engine) classOffers(id core.CharID) []core.Offer {
	return e.ruleset().Allowed(core.Training{Party: core.Party{id}, DualClass: true}).Offers
}
