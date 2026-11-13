package engine

import (
	"gobox/rules"
	"slices"
)

// NewClass = the hall's New Class (front.Game): member n (< 0: the selected
// one) changes to por class number class (< 0: not picked yet, asked).
func (e *Engine) NewClass(n, class int) error {
	return e.newClass(n, e.fact(rules.NoChar, rules.ClassByNumber{N: class}).(rules.Class))
}

// newClass is NewClass to class (ClassNone: asked, askClass).
func (e *Engine) newClass(n int, class rules.Class) error {
	if !e.inParty() || !slices.Contains(e.partyFunctions(), partyNewClass) {
		return rules.ErrNotNow
	}
	id := e.actor(n)
	offers := e.classOffers(id)
	if len(offers) == 0 {
		e.say(e.name(id) + " doesn't qualify.")
		return rules.ErrNotNow
	}
	for _, o := range offers {
		if o.Action.(rules.ChangeClass).To == class {
			_, err := e.ruleset().Resolve(rules.Training{Party: e.partyIDs(), DualClass: true}, o.Action, nil)
			return err
		}
	}
	return e.askClass(offers, n)
}

// classOffers are the rules' ChangeClass offers for member id in a hall.
func (e *Engine) classOffers(id rules.CharID) []rules.Offer {
	return e.ruleset().Allowed(rules.Training{Party: rules.Party{id}, DualClass: true}).Offers
}
