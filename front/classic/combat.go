package classic

import (
	"slices"

	"gobox/front"
	"gobox/game"
)

// turnAct is the combat menu's action a Decision does (false: not a menu entry).
func turnAct(d game.Decision) (front.ActKind, bool) {
	switch d.(type) {
	case game.Attack:
		return front.ActAttack, true
	case game.Ranged:
		return front.ActRanged, true
	case game.Guard:
		return front.ActGuard, true
	case game.Items:
		return front.ActItems, true
	case game.Cast:
		return front.ActCast, true
	case game.Maneuver:
		return front.ActManeuver, true
	case game.TurnUndead:
		return front.ActTurn, true
	case game.Bandage:
		return front.ActBandage, true
	case game.Wait:
		return front.ActWait, true
	case game.EndTurn:
		return front.ActEnd, true
	case game.Quick:
		return front.ActQuick, true
	case game.Flee:
		return front.ActFlee, true
	}
	return 0, false
}

// Battle is b (the game's) with what s.Combat gives in its place: the
// combat menu (Turn), each combatant's name, hit points, armor class,
// damage and status, the squares a move reaches, the foes an attack or aim
// takes and the aim (marks, area, why not). b's map, icons and effects stay.
func Battle(b *front.Battle, s game.State) *front.Battle {
	c := s.Combat
	if b == nil || c == nil {
		return b
	}
	o := *b
	o.Units = slices.Clone(b.Units)
	for i, u := range o.Units {
		if k := slices.IndexFunc(c.Units, func(g game.Unit) bool { return g.ID == u.ID }); k >= 0 {
			g := c.Units[k]
			o.Units[i].Party, o.Units[i].Active, o.Units[i].Name = !g.Foe, g.Active, g.Name
			o.Units[i].HP, o.Units[i].MaxHP, o.Units[i].AC = g.HP, g.MaxHP, g.AC
			o.Units[i].Hit, o.Units[i].Dmg, o.Units[i].Status = g.Hit, g.Dmg, g.Status
		}
	}
	o.Reach = nil
	for _, r := range c.Reach {
		o.Reach = append(o.Reach, r.At)
	}
	o.Targets, o.Panel, o.Marks, o.Area = slices.Clone(c.Targets), c.Panel, slices.Clone(c.Marks), c.Area
	o.AreaAt, o.Skipped, o.AimWhy = slices.Clone(c.AreaAt), slices.Clone(c.Skipped), c.Why
	o.Turn = nil
	if t := c.Turn; t != nil {
		o.Turn = &front.Turn{Unit: t.Unit, Aim: t.Aim, Casting: t.Spell, Moves: t.Moves}
		if b.Turn != nil { // the weapon icons and the rule set's maneuvers are the game's
			o.Turn.Weapon, o.Turn.Missile, o.Turn.Maneuvers = b.Turn.Weapon, b.Turn.Missile, b.Turn.Maneuvers
		}
		for _, op := range t.Options {
			if act, ok := turnAct(op.Do); ok {
				o.Turn.Can = append(o.Turn.Can, act)
			}
		}
		for _, sp := range c.Spells {
			o.Turn.Spells = append(o.Turn.Spells, front.Spell{ID: sp.ID, Name: sp.Name})
		}
	}
	return &o
}
