package engine

import (
	"gobox/rules"
	"slices"
)

// newClass = the hall's Change Class: member n (< 0: the selected one)
// changes to class (ClassNone: asked, askClass) once sure (asked with what
// the change costs, the offer's Note).
func (e *Engine) newClass(n int, class rules.Class, sure bool) error {
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
			if !sure {
				err := e.askSure(o.Note+" Change class?", pendingCall{kind: callClass, member: n, class: class})
				e.ext.pq.q.Default = 0 // it cannot be undone: the highlight starts on No
				return err
			}
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
