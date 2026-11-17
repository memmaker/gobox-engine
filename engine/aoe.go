package engine

// Area of effect preview (extra): while the player aims a spell, the
// combatants it would hit get the dotted frame Move and Aim use (markReach)
// and are counted in the prompt, its squares go to the snapshot (area).
// The test is the cast's own: buildNear round the square, radius kind&7
// for chooseSpellTargets's kinds 9..E, 0 (the square itself) for single
// targets; a bolt (bolts) walks boltWalk dry: the squares it crosses are
// the area, the ones it strikes the marks.

import (
	"gobox/game"
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
	a := e.areaAt(spell, e.ext.aoe, rules.BoardPos{X: x, Y: y})
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

// areaAt is the rules' word (AreaFact, the cast's own test) on who an area
// of aoe (radius+1) of spell ("" a maneuver's) aimed at at takes.
func (e *Engine) areaAt(spell rules.SpellID, aoe int, at rules.BoardPos) rules.AreaPreview {
	e.as(e.sel)
	return e.fact(e.selected(), rules.AreaFact{Spell: spell, At: at, Area: aoe}).(rules.AreaPreview)
}

// Preview is game.Previewer: what the engine's own aim (Aim) or an aimed
// spell or maneuver (its one square in Ats) would take, as areaAt; empty
// for an aim with no area.
func (e *Engine) Preview(a game.Action) game.Preview {
	if !e.battleUp() {
		return game.Preview{}
	}
	var spell rules.SpellID
	aoe, at, why := 0, rules.BoardPos{}, ""
	switch a := a.(type) {
	case game.Aim:
		if e.cs.aim == nil || e.ext.aoe == 0 {
			return game.Preview{}
		}
		if e.ext.aoeSpell != 0 {
			spell = e.spellID(e.ext.aoeSpell)
		}
		aoe, at = e.ext.aoe, a.At
	case game.Cast:
		if len(a.Ats) != 1 {
			return game.Preview{}
		}
		spell = e.spellID(a.Spell)
		aim := e.fact(e.selected(), rules.SpellAimFact{Spell: spell}).(rules.SpellAim)
		if aim.Kind == rules.SpellAimArea {
			aoe = aim.Area
		}
		at = a.Ats[0]
		if b := e.fightState(); b != nil && b.Turn != nil {
			why = e.aimWhy(b, aimOf{rng: aim.Range, foesOnly: aim.FoesOnly, ground: aim.Kind == rules.SpellAimArea}, at)
		}
	case game.DoManeuver:
		if len(a.Ats) != 1 {
			return game.Preview{}
		}
		for _, m := range e.maneuvers(e.selected()) {
			if int(m.ID) == a.ID && m.Aim == rules.AimArea {
				aoe = m.Area + 1 // as maneuver's aim
			}
		}
		at = a.Ats[0]
	}
	if aoe == 0 {
		return game.Preview{Why: why}
	}
	p := e.areaAt(spell, aoe, at)
	return game.Preview{Area: slices.Clone(p.Area), Hit: slices.Clone(p.Hit), Skipped: slices.Clone(p.Skipped), Why: why}
}

func (e *Engine) aoeUnmark() { e.snap.marks, e.snap.area, e.snap.skipped = nil, nil, nil }
