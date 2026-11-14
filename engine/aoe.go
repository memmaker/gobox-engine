package engine

// Area of effect preview (extra): while the player aims a spell, the
// combatants it would hit get the dotted frame Move and Aim use (markReach)
// and are counted in the prompt, its squares go to the snapshot (area).
// The test is the cast's own: buildNear round the square, radius kind&7
// for chooseSpellTargets's kinds 9..E, 0 (the square itself) for single
// targets; a bolt (bolts) walks boltWalk dry: the squares it crosses are
// the area, the ones it strikes the marks.

import (
	"gobox/rules"
	"slices"
)

// aoeMarks marks who the spell aimed at (x,y) would hit (the rules'
// AreaFact); aoeUnmark takes the marks off again.
func (e *Engine) aoeMarks(x, y int) {
	if e.ext.aoe == 0 {
		return
	}
	var spell rules.SpellID
	if e.ext.aoeSpell != 0 {
		spell = e.spellID(e.ext.aoeSpell)
	}
	e.as(e.sel)
	a := e.fact(e.selected(), rules.AreaFact{Spell: spell, At: rules.BoardPos{X: x, Y: y}, Area: e.ext.aoe}).(rules.AreaPreview)
	e.snap.area, e.snap.skipped = slices.Clone(a.Area), slices.Clone(a.Skipped)
	marks := map[rules.BoardPos]bool{}
	for _, p := range a.Hit {
		if p.X != x || p.Y != y { // the aim's cursor marks its own square
			marks[p] = true
		}
	}
	e.snap.marks = nil
	for p := range marks {
		e.snap.marks = append(e.snap.marks, p)
	}
}

func (e *Engine) aoeUnmark() { e.snap.marks, e.snap.area, e.snap.skipped = nil, nil, nil }
