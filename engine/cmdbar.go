package engine

import "gobox/rules"

// The context-sensitive command bar: Cast is
// offered only when the selected member could cast a spell.

// canCast: Cast would offer member id a spell (the checks of canStudy and
// CastMenu, without their messages).
func (e *Engine) canCast(id rules.CharID) bool {
	if e.noMagic() || e.ruleset().Sheet(id).Status != rules.StatusOkay || !e.ruleset().Check(rules.Fit{Who: id}, nil).Yes {
		return false
	}
	return len(e.fact(id, rules.SpellMenuFact{Kind: 0}).(rules.SpellMenu)) > 0
}
