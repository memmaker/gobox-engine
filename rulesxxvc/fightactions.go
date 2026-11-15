package rulesxxvc

import (
	"fmt"
	"slices"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// endActions is charhlp.pas EndActions: initiative, guarding and movement go.
func endActions(s *combatState) {
	s.initiative, s.guarding, s.movement = 0, false, 0
}

// allowedFight is combat.pas CombatMenu's choices for the actor.
// Jetpack needs areaKind Venus or Mars wilderness (core.AreaWilderness).
func (r *Rules) allowedFight(f rules.Fight) rules.Allowed {
	f.Board = r.board(f.Board)
	c, s := r.chars[f.Actor], r.fight[f.Actor]
	var a rules.Allowed
	if c == nil || s == nil {
		return a
	}
	offer := func(x rules.Action) { a.Offers = append(a.Offers, rules.Offer{Action: x}) }
	for _, u := range f.Board.Units {
		if d := r.chars[u.ID]; d != nil && d.Enabled && d.team() != c.team() {
			a.Aimed = append(a.Aimed, rules.Aimed{Action: rules.Attack{Target: u.ID}, Squares: []rules.BoardPos{u.At}, Targets: 1})
		}
	}
	at, _ := unitAt(f, f.Actor)
	a.Aimed = append(a.Aimed, r.groundAims(f, c, at, a.Aimed)...)
	if s.movement > 0 {
		a.Moves = r.reach(f, at, s.movement)
	}
	if c.Skills.Level(intimidateSkill) > 0 && c.Status != StDodging && c.Status != StFailedZeroG &&
		len(r.targetList(f, f.Actor, c, 1, false)) > 0 {
		offer(rules.Intimidate{})
	}
	offer(rules.Quick{})
	if ps := r.patients(f, c, at); len(ps) > 0 {
		offer(rules.Aid{Target: ps[0]})
	}
	if s.movement > 0 && c.Status != StDodging {
		if c.Status != StSprinting {
			if c.Status != StUsedJetpack && c.Status != StFailedZeroG {
				offer(rules.Sprint{})
			}
			offer(rules.Dodge{})
		}
		if c.Status != StUsedJetpack && r.hasJetpack(c) && f.Terms.Area == rules.AreaWilderness {
			for y := 0; y < f.Board.Size.Y; y++ {
				for x := 0; x < f.Board.Size.X; x++ {
					if p := (rules.BoardPos{X: x, Y: y}); r.free(f, p) {
						a.Aimed = append(a.Aimed, rules.Aimed{Action: rules.Jetpack{To: p}, Squares: []rules.BoardPos{p}})
					}
				}
			}
		}
	}
	offer(rules.Delay{})
	if c.Status != StSprinting && c.Status != StDodging {
		offer(rules.Guard{})
	}
	offer(rules.Flee{})
	offer(rules.EndTurn{})
	return a
}

// groundAims is ManualAimCursor's acceptance for a loaded area weapon
// (HasLoadedAreaWeapon: an area weapon with a quantity): every cell a WalkLine
// from the actor reaches within the weapon range (ReadiedWeaponRange, 0 as 1;
// distance div 2), not impassable, with canAttack (CombatOptionAllowed: not
// sprinting). Cells an enemy's Attack{Target} already covers are left to it.
func (r *Rules) groundAims(f rules.Fight, c *Character, at rules.BoardPos, aimed []rules.Aimed) []rules.Aimed {
	ci, _, _, ok := r.readied(c, xxvc.SlotWeapon)
	if !ok || !r.hasAreaWeapon(c) || ci.Ref.Count <= 0 || c.Status == StSprinting {
		return nil
	}
	covered := map[rules.BoardPos]bool{}
	for _, x := range aimed {
		for _, p := range x.Squares {
			covered[p] = true
		}
	}
	rng := r.readiedWeaponRange(c)
	var out []rules.Aimed
	for y := 0; y < f.Board.Size.Y; y++ {
		for x := 0; x < f.Board.Size.X; x++ {
			p := rules.BoardPos{X: x, Y: y}
			if covered[p] || terrainRow(terrainAt(f.Board, p)).cost == impassable {
				continue
			}
			if _, dist, reached, _ := walkLine(f.Board, at, p, 255, true, false); reached && dist/2 <= rng {
				out = append(out, rules.Aimed{Action: rules.Attack{At: &p}, Squares: []rules.BoardPos{p}})
			}
		}
	}
	return out
}

// resolveFight is ManualControl's keys and their cbtacts.pas rules, on the
// arena as the rules hold it; the fallen leave corpses.
func (r *Rules) resolveFight(f rules.Fight, act rules.Action, rng *rules.RNG) (rules.Outcome, error) {
	f.Board = r.board(f.Board)
	o, err := r.resolveAction(f, act, rng)
	if err == nil {
		r.layFallen(&f.Board, &o)
	}
	return o, err
}

func (r *Rules) resolveAction(f rules.Fight, act rules.Action, rng *rules.RNG) (rules.Outcome, error) {
	c, s := r.chars[f.Actor], r.fight[f.Actor]
	if c == nil || s == nil {
		return rules.Outcome{}, fmt.Errorf("Resolve %T: actor %d not in the fight", act, f.Actor)
	}
	switch act := act.(type) {
	case rules.Confirmed:
		if a, ok := act.Action.(rules.Attack); ok {
			return r.attack(f, f.Actor, a.Target, rng)
		}
	case rules.Attack:
		if act.At != nil { // CbtActs.AttackGroundTarget
			return r.groundAttack(f, f.Actor, c, s, *act.At, rng)
		}
		if r.hasAreaWeapon(c) { // AimMenu: an area weapon fires at the target's cell, no ally question
			at, ok := unitAt(f, act.Target)
			if !ok {
				return rules.Outcome{}, fmt.Errorf("Attack %d on %d: not on the board", f.Actor, act.Target)
			}
			return r.groundAttack(f, f.Actor, c, s, at, rng)
		}
		if d := r.chars[act.Target]; d != nil && d.team() == c.team() && act.Target != f.Actor {
			return rules.Outcome{Confirm: "Attack Ally?"}, nil // CheckAttackAlly
		}
		return r.attack(f, f.Actor, act.Target, rng)
	case rules.Move: // MoveMenu, CombatMoveCreature per step
		at, _ := unitAt(f, f.Actor)
		i := slices.IndexFunc(r.reach(f, at, s.movement), func(x rules.Reach) bool { return x.At == act.To })
		if i < 0 {
			return rules.Outcome{}, fmt.Errorf("Move %d to %v: not reachable with the movement left", f.Actor, act.To)
		}
		return r.walk(f, f.Actor, c, s, at, r.reach(f, at, s.movement)[i].Path, rng)
	case rules.Intimidate: // 'I'; ponytail: the turn ending after it is assumed (EndActions)
		at, _ := unitAt(f, f.Actor)
		o := r.intimidate(f, c, at, rng)
		endActions(s)
		return o, nil
	case rules.Aid: // 'A': TreatPatients, the turn over (AiMoveAttack sets initiative 0)
		at, _ := unitAt(f, f.Actor)
		o := r.treat(f, c, at, rng)
		endActions(s)
		return o, nil
	case rules.Jetpack: // 'J'
		at, _ := unitAt(f, f.Actor)
		return r.jetpack(f, c, s, at, act.To, rng)
	case rules.Sprint: // 'S'
		if c.Status != StSprinting {
			c.Status = StSprinting
			s.movement *= 2
		}
		return rules.Outcome{}, nil
	case rules.Dodge: // 'D'
		c.Status = StDodging
		endActions(s)
		return rules.Outcome{}, nil
	case rules.Guard: // CbtActs.Guard
		endActions(s)
		s.guarding = true
		return rules.Outcome{}, nil
	case rules.Delay: // 'W': initiative 1, after the others still to act
		s.initiative = 1
		r.initiative = append(r.initiative, f.Actor)
		return rules.Outcome{}, nil
	case rules.EndTurn: // 'E'
		endActions(s)
		return rules.Outcome{}, nil
	case rules.Flee:
		return r.flee(f, c, s, rng), nil
	case rules.Maneuver:
		if act.ID == monsterSpecial && s.specialDue {
			s.specialDue = false
			return r.monsterSpecial(f, f.Actor, c, s, rng), nil
		}
	}
	return rules.Outcome{}, fmt.Errorf("Resolve %T in a fight: %w", act, ErrNotPorted)
}

// flee is cbtacts.pas AttemptFlee: no enabled enemy, or the fastest enemy's
// movement div 2 below the actor's, gets away; equal: 1 in 2.
// ponytail: CountTargets(255) as every enabled enemy.
func (r *Rules) flee(f rules.Fight, c *Character, s *combatState, rng *rules.RNG) rules.Outcome {
	own, fastest, enemies := r.RecalcMovement(c)/2, 0, 0
	for _, id := range r.fighters(f) {
		if e := r.chars[id]; e.Enabled && e.team() != c.team() {
			enemies++
			fastest = max(fastest, r.RecalcMovement(e)/2)
		}
	}
	escaped := enemies == 0 || fastest < own || fastest == own && rules.RNGDice(rng, 1, 2) == 1
	var o rules.Outcome
	if escaped {
		c.Status, c.Enabled = StFleeing, false // PutOutOfPlayWithStatus
		removeIncapacitating(c, s)
		o.Effects = append(o.Effects, rules.Fled{Unit: f.Actor})
	} else {
		o.Effects = append(o.Effects, rules.Narrated{Text: "Escape is blocked"})
	}
	endActions(s)
	return o
}
