package engine

import "gobox/core"

// The context-sensitive command bar: Cast is
// offered only when the selected member could cast a spell.

// canCastNow: Cast would offer a spell (the checks of canStudy and
// CastMenu, without their messages).
func (e *Engine) canCastNow() bool {
	id := e.selected()
	if e.worldWord(core.WordNoMagic) > 0 || e.ruleset().Sheet(id).Status != core.StatusOkay || !e.ruleset().Check(core.Fit{Who: id}, nil).Yes {
		return false
	}
	list, ids := e.spellList, e.spellIDs
	any := e.buildSpellList(0)
	e.spellList, e.spellIDs = list, ids
	return any
}
