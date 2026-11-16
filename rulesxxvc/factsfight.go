package rulesxxvc

import (
	"fmt"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// The fight's Facts: the board the ruleset holds (r.ground) and the fighters' turn records, as the engine's
// state build and turn flow ask them. A fact that is AD&D's alone answers XXVc's none: the zero value.

// factFight answers the Facts of the fight topic; ok false for any other.
func (r *Rules) factFight(id rules.CharID, f rules.Fact) (v rules.Value, ok bool, err error) {
	switch f.(type) {
	case rules.GroundFact:
		return r.groundFact(), true, nil
	case rules.UnitsFact:
		return r.unitsFact(), true, nil
	case rules.UnitCharFact:
		who := charOf(f.(rules.UnitCharFact).Unit)
		if _, ok := r.boardUnit(who); !ok {
			who = rules.NoChar
		}
		return rules.UnitChar{Who: who}, true, nil
	}
	c := r.chars[id]
	switch f := f.(type) {
	case rules.UnitFact:
		if c == nil {
			return nil, true, errNoFightChar
		}
		return rules.Number(int(unitOf(id))), true, nil
	case rules.AtFact:
		if u, ok := r.boardUnit(id); ok {
			return u.At, true, nil
		}
		return rules.BoardPos{}, true, nil
	case rules.SizeFact:
		if u, ok := r.boardUnit(id); ok {
			return rules.Number(int(u.Size)), true, nil
		}
		return rules.Number(0), true, nil
	case rules.IconFact:
		if c == nil {
			return nil, true, errNoFightChar
		}
		return rules.Number(c.Order), true, nil
	case rules.FightCardFact:
		if c == nil {
			return nil, true, errNoFightChar
		}
		return r.fightCard(id, c), true, nil
	case rules.WeaponFact:
		if c == nil {
			return nil, true, errNoFightChar
		}
		// XXVc has one weapon slot: the Pool's Launcher/Missiles alternative (Ranged) is none.
		if ci, _, _, armed := r.readied(c, xxvc.SlotWeapon); armed && !f.Ranged {
			return rules.Handles{ci.Handle}, true, nil
		}
		return rules.Handles{}, true, nil
	case rules.WeaponRulesFact:
		// AD&D's weapon rule lines (speed, vs. size): XXVc's weapon is its ItemCard: none.
		return rules.Lines(nil), true, nil
	case rules.TargetWhyFact:
		// Why a target cannot be hit this turn, worded for the status row: XXVc gives no reason, only a line or none.
		return rules.Text(""), true, nil
	case rules.MayActFact:
		t := r.chars[f.Target]
		return rules.Yes(c != nil && t != nil && t.Enabled && t.team() != c.team()), true, nil
	case rules.ManeuversFact:
		// AD&D 4e-style maneuver cards: XXVc's choices come through Allowed (Sprint, Dodge, Aid ...): none.
		return rules.ManeuverCards(nil), true, nil
	case rules.SpellAimFact:
		return rules.SpellAim{}, true, nil // no spells in XXVc: the zero aim
	case rules.StepFact:
		return r.stepFact(id, c, f.Dir), true, nil
	case rules.HitFact:
		return r.hitFact(id, f.Target)
	case rules.RangeToFact:
		a, aok := r.boardUnit(id)
		b, bok := r.boardUnit(f.Target)
		if !aok || !bok {
			return nil, true, fmt.Errorf("RangeToFact %d to %d: not on the board", id, f.Target)
		}
		return rules.Number(distance(a.At, b.At)), true, nil
	}
	return nil, false, nil
}

// errNoFightChar: a fight fact asked of nobody.
var errNoFightChar = fmt.Errorf("fight fact: no such character")

// boardUnit is id's unit on the ruleset's ground.
func (r *Rules) boardUnit(id rules.CharID) (rules.BoardUnit, bool) {
	for _, u := range r.ground.Units {
		if u.ID == id {
			return u, true
		}
	}
	return rules.BoardUnit{}, false
}

// fightCard is the turn record: Delay 1 while the fighter acts this round (it rolled an initiative and is in play),
// Moves the half-squares left (combatState.movement). Casting is 0: XXVc has no spells begun a round before.
func (r *Rules) fightCard(id rules.CharID, c *Character) rules.FightCard {
	s := r.fight[id]
	card := rules.FightCard{Auto: c.Side == rules.SideAuto}
	if s != nil {
		card.Delay, card.Moves, card.Attacked = 1, s.movement, s.hasAttacked
	}
	return card
}

// groundFact is the arena as squares: the terrain's number as the tile, a wall where no move can pay (cost
// IMPASSABLE). Height is 0: XXVc's lines are classed by terrain, not by heights; Crown false.
func (r *Rules) groundFact() rules.Ground {
	var g rules.Ground
	b := r.board(rules.Board{})
	for y := range g.Squares {
		for x := range g.Squares[y] {
			t := terrainAt(b, rules.BoardPos{X: x, Y: y})
			g.Squares[y][x] = rules.Square{Tile: int(t), Wall: terrainRow(t).cost == impassable}
		}
	}
	for size := 1; size <= len(iconCells); size++ {
		g.Footprints[size] = rules.Squares(iconCells[size-1])
	}
	return g
}

// unitsFact is the units on the ground in placement order (the party list's). Number is the unit number the
// Effects carry (unitOf).
func (r *Rules) unitsFact() rules.Units {
	var out rules.Units
	for _, u := range r.ground.Units {
		c := r.chars[u.ID]
		if c == nil || u.Size == 0 {
			continue
		}
		out = append(out, rules.Unit{ID: u.ID, Number: int(unitOf(u.ID)), At: u.At, Size: int(u.Size), Facing: int(u.Facing),
			Party: c.team(), Helpless: hasEffect(c, EffHelpless), Active: c.Enabled, Icon: c.Order, Name: c.Name,
			HP: c.HP, MaxHP: c.MaxHP, AC: c.ArmorClass[0], Damage: diceText(c.Damage[0]), Squares: rules.Squares(cellsOf(u))})
	}
	return out
}

// stepFact is what lies one step off the actor's square in direction dir: the unit there, the ground, whether a foe.
func (r *Rules) stepFact(id rules.CharID, c *Character, dir int) rules.StepInfo {
	b := r.board(rules.Board{})
	from, ok := r.boardUnit(id)
	if !ok {
		return rules.StepInfo{}
	}
	to := from.At.Add(dirStep[dir&7])
	t := terrainAt(b, to)
	info := rules.StepInfo{Ground: int32(t)}
	if who, ok := r.occupant(b, to); ok {
		info.Unit = int32(unitOf(who))
		info.Foe = r.chars[who].team() != c.team()
	}
	return info
}

// hitFact is the chance in percent that the attacker's roll hits the target from the front: the d20 values that
// hit, as attack rolls them (AttackDie: 1 always misses, 20 is naturalAttackMax), times five.
func (r *Rules) hitFact(aid, did rules.CharID) (rules.Value, bool, error) {
	a, d, s := r.chars[aid], r.chars[did], r.fight[aid]
	if a == nil || d == nil || s == nil {
		return nil, true, errNoFightChar
	}
	b := r.board(rules.Board{})
	aat, _ := unitAt(rules.Fight{Board: b}, aid)
	dat, _ := unitAt(rules.Fight{Board: b}, did)
	rc := *d // recalculated on a copy: a fact changes nothing
	r.RecalcStats(&rc)
	armor, mod := r.hitNeed(b, a, &rc, s, false, r.readiedWeaponRange(a), aat, dat)
	hits := 0
	for die := 2; die <= 20; die++ {
		roll := die
		if die == 20 {
			roll = naturalAttackMax
		}
		if helpless(d) || AttackHits(roll, a.AttackBonus, mod, armor) {
			hits++
		}
	}
	return rules.Number(hits * 5), true, nil
}
