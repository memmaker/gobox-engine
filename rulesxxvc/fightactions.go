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
	if len(a.Aimed) > 0 { // the turn bar's Attack button: a foe (or a ground aim) can be aimed at
		offer(rules.Attack{})
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
	switch a := act.(type) { // the set-up: no actor yet (combatst.pas CombatStart)
	case rules.BuildGround:
		return rules.Outcome{}, r.buildGround(f, a, rng)
	case rules.PlaceFight:
		return r.placeAllCreatures(f, a.Walls), nil
	case rules.TakeLoot:
		return rules.Outcome{}, r.takeLoot(a)
	case rules.FightPlaced:
		return rules.Outcome{}, nil // CombatStart's last rule, UpdateHostileHpPercent, BeginFight already ran
	}
	own := ownsBoard(f)
	f.Board = r.board(f.Board)
	o, err := r.resolveAction(f, act, rng)
	if err == nil {
		// The turn is used when the actor has no actions left (EndActions: initiative 0), or it waited (Delay).
		_, waited := act.(rules.Delay)
		if st := r.fight[f.Actor]; st != nil {
			o.TurnOver = o.Confirm == "" && (st.initiative <= 0 || waited)
		}
		r.layFallen(&f.Board, &o)
		if own {
			r.keepOnGround(o.Effects)
		}
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
		switch a := act.Action.(type) {
		case rules.Attack:
			return r.attack(f, f.Actor, a.Target, rng)
		case rules.Flee: // the "Flee:" asked of a player was answered Yes
			return r.flee(f, c, s, rng), nil
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
	case rules.Quick: // Combat.TurnOnAuto: the actor (All: every party member in play) goes to quick-fight, the computer steering
		ids := []rules.CharID{f.Actor}
		if act.All {
			ids = r.fightParty(f)
		}
		for _, id := range ids {
			if q := r.chars[id]; q != nil && q.team() {
				q.Side = rules.SideAuto
				if qs := r.fight[id]; qs != nil && r.chars[qs.target] != nil && r.chars[qs.target].team() {
					qs.target = rules.NoChar // an ally target is dropped
				}
			}
		}
		return rules.Outcome{}, nil
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
	case rules.Flee: // combat.pas MoveMenu: a step off the map asks "Flee:" (default No) before AttemptFlee
		return rules.Outcome{Confirm: fleeQuestion}, nil
	case rules.Maneuver:
		if act.ID == monsterSpecial && s.specialDue {
			s.specialDue = false
			return r.monsterSpecial(f, f.Actor, c, s, rng), nil
		}
	}
	return rules.Outcome{}, fmt.Errorf("Resolve %T in a fight: %w", act, ErrNotPorted)
}

// fleeQuestion is combat.pas MoveMenu's YesNoMenu('Flee:', DEFAULT_NO).
const fleeQuestion = "Flee:"

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
		o.Effects = append(o.Effects, rules.Narrated{Text: c.Name + " " + msgGotAway}, rules.Fled{Unit: f.Actor})
	} else {
		o.Effects = append(o.Effects, rules.Narrated{Text: "Escape is blocked"})
	}
	endActions(s)
	return o
}
