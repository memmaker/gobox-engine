package rulesxxvc

import (
	"slices"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// The C_Skill numbers the fight menu reads.
var (
	intimidateSkill, _ = skillNo("intimidate")
	firstAidSkill, _   = skillNo("first-aid")
	treatLightSkill, _ = skillNo("treat-light-wounds")
	treatStunSkill, _  = skillNo("treat-stun-paralysis")
	jetpackSkill, _    = skillNo("use-jetpack")
)

// reach is MoveMenu's walkable squares: every free square the movement left
// pays for, the terrain's cost times 3 on a diagonal, times 2 straight
// (CombatMoveCreature), with the cheapest path. Each step is ProbeDirection's over every cell id's icon covers
// (a big unit's own cells do not block it; the costliest terrain under it is the step's); id not on the board
// is one cell.
func (r *Rules) reach(f rules.Fight, id rules.CharID, from rules.BoardPos, movement int) []rules.Reach {
	unit := rules.BoardUnit{ID: id, Size: 1}
	if i := slices.IndexFunc(f.Board.Units, func(u rules.BoardUnit) bool { return u.ID == id }); i >= 0 {
		unit = f.Board.Units[i]
	}
	type node struct {
		cost int
		path []rules.BoardPos
	}
	best := map[rules.BoardPos]node{from: {}}
	frontier := []rules.BoardPos{from}
	for len(frontier) > 0 {
		var next []rules.BoardPos
		for _, p := range frontier {
			n := best[p]
			for dir, d := range dirStep {
				to := p.Add(d)
				unit.At = p
				blocked, t, _ := r.probeFrom(f.Board, unit, dir)
				if blocked || t == 0 || terrainRow(t).cost == impassable {
					continue
				}
				cost := n.cost + stepCost(dir, t)
				if cost > movement {
					continue
				}
				if b, ok := best[to]; ok && b.cost <= cost {
					continue
				}
				best[to] = node{cost, append(append([]rules.BoardPos{}, n.path...), to)}
				next = append(next, to)
			}
		}
		frontier = next
	}
	var rs []rules.Reach
	for y := 0; y < f.Board.Size.Y; y++ { // board order, never map order
		for x := 0; x < f.Board.Size.X; x++ {
			if n, ok := best[rules.BoardPos{X: x, Y: y}]; ok && len(n.path) > 0 {
				rs = append(rs, rules.Reach{At: rules.BoardPos{X: x, Y: y}, Path: n.path})
			}
		}
	}
	return rs
}

// pathCost is a path's movement, each step's stepCost.
func pathCost(b rules.Board, from rules.BoardPos, path []rules.BoardPos) int {
	cost := 0
	for _, p := range path {
		cost += stepCost(rules.DirectionTo(from.X, from.Y, p.X, p.Y), terrainAt(b, p))
		from = p
	}
	return cost
}

// patients is viewchar.pas FindPatients in a fight: adjacent team members in
// play who are patients (isPatient).
func (r *Rules) patients(f rules.Fight, c *Character, at rules.BoardPos) []rules.CharID {
	var ps []rules.CharID
	for _, u := range f.Board.Units {
		p := r.chars[u.ID]
		if p == nil || p.Status == StFleeing || distance(at, u.At) > 1 || !r.isPatient(c, p) {
			continue
		}
		ps = append(ps, u.ID)
	}
	return ps
}

// isPatient is FindPatients' test of p for the medic c: a team member,
// dying (the medic has First Aid or Treat Light Wounds) or enabled and
// dizzy or stunned (Treat Stun).
func (r *Rules) isPatient(c, p *Character) bool {
	dying := p.Status == StDying && (c.Skills.Level(firstAidSkill) > 0 || c.Skills.Level(treatLightSkill) > 0)
	stun := p.Enabled && (hasEffect(p, EffDizzy) || hasEffect(p, EffStunned)) && c.Skills.Level(treatStunSkill) > 0
	return p.team() && (dying || stun)
}

// treat is viewchar.pas TreatPatients: the dying are bandaged (unconscious),
// the dizzy or stunned treated on a Treat Stun roll.
func (r *Rules) treat(f rules.Fight, c *Character, at rules.BoardPos, rng *rules.RNG) rules.Outcome {
	var o rules.Outcome
	for _, id := range r.patients(f, c, at) {
		o.Effects = append(o.Effects, r.treatOne(c, id, r.fight[id], rng)...)
	}
	return o
}

// treatOne is TreatPatients for one patient; s is its combat state, nil
// outside a fight.
func (r *Rules) treatOne(c *Character, id rules.CharID, s *combatState, rng *rules.RNG) []rules.Effect {
	p := r.chars[id]
	switch {
	case p.Status == StDying && (c.Skills.Level(firstAidSkill) > 0 || c.Skills.Level(treatLightSkill) > 0):
		p.Status = StUnconscious
		return []rules.Effect{rules.Narrated{Text: p.Name + " is bandaged"}}
	case hasEffect(p, EffDizzy) || hasEffect(p, EffStunned):
		if g, _ := SkillCheck([]int8{skillValue(c, treatStunSkill)}, rules.Normal, rng); g >= rules.Made {
			removeEffect(p, s, EffDizzy)
			removeEffect(p, s, EffStunned)
			return []rules.Effect{rules.Narrated{Text: p.Name + " is treated"}}
		}
	}
	return nil
}

// intimidate is cbtacts.pas IntimidateTarget: on an Intimidate roll every
// adjacent enemy that is not helpless is SCARED (10 minutes, strength 12)
// unless it saves.
func (r *Rules) intimidate(f rules.Fight, c *Character, at rules.BoardPos, rng *rules.RNG) rules.Outcome {
	var o rules.Outcome
	if g, _ := SkillCheck([]int8{skillValue(c, intimidateSkill)}, rules.Normal, rng); g < rules.Made {
		return rules.Outcome{Effects: []rules.Effect{rules.Narrated{Text: c.Name + " fails to intimidate."}}}
	}
	o.Effects = append(o.Effects, rules.Narrated{Text: c.Name + " looks real menacing."})
	for _, u := range f.Board.Units {
		e := r.chars[u.ID]
		if e == nil || !e.Enabled || e.team() == c.team() || distance(at, u.At) > 1 || helpless(e) {
			continue
		}
		if r.addEffect(e, c, Effect{ID: EffScared, Minutes: 10, Strength: 12}, rollSave(e, 0, 0, rng)) {
			o.Effects = append(o.Effects, rules.Narrated{Text: e.Name + " is intimidated"})
		}
	}
	return o
}

// jetpack is cbtacts.pas CombatUseJetpack: on a Use Jetpack roll the flight
// lands on the chosen free square; a failure tumbles through eleven d24
// neighbour rolls (the 5x5 around, not the centre) and lands on the last free
// one. Either way the movement is gone and the status UsedJetpack.
// Guards attack at take-off and on landing (GuardersAttack, no line needed
// while it flies), and after a failed flight's landing.
func (r *Rules) jetpack(f rules.Fight, c *Character, s *combatState, at, to rules.BoardPos, rng *rules.RNG) (rules.Outcome, error) {
	s.movement = 0
	f.Board.Units = slices.Clone(f.Board.Units)
	u := slices.IndexFunc(f.Board.Units, func(u rules.BoardUnit) bool { return u.ID == f.Actor })
	guards := func(o *rules.Outcome) error {
		g, err := r.guardersAttack(f, f.Actor, c, rng)
		o.Effects = append(o.Effects, g.Effects...)
		return err
	}
	if g, _ := SkillCheck([]int8{skillValue(c, jetpackSkill)}, rules.Normal, rng); g >= rules.Made && r.free(f, to) {
		c.Status = StUsingJetpack
		var o rules.Outcome
		if err := guards(&o); err != nil {
			return o, err
		}
		if c.Enabled {
			o.Effects = append(o.Effects, rules.Walked{Unit: unitOf(f.Actor), Path: []rules.BoardPos{to}})
			f.Board.Units[u].At = to
			if err := guards(&o); err != nil {
				return o, err
			}
		}
		if c.Enabled {
			c.Status = StUsedJetpack
		}
		return o, nil
	}
	c.Status = StUsingJetpack
	o := rules.Outcome{Effects: []rules.Effect{rules.Narrated{Text: c.Name + " fails at use jetpack"}}}
	p, landed, path := at, at, []rules.BoardPos{}
	for range 11 {
		var cand rules.BoardPos
		for {
			roll := rules.RNGDice(rng, 1, 24) - 1
			if roll >= 12 {
				roll++
			}
			cand = rules.BoardPos{X: p.X + roll%5 - 2, Y: p.Y + roll/5 - 1}
			if cand.X >= 0 && cand.Y >= 0 && cand.X < f.Board.Size.X && cand.Y < f.Board.Size.Y {
				break
			}
		}
		p = cand
		path = append(path, p)
		if r.free(f, p) {
			landed = p
		}
	}
	if i := indexOf(path, landed); i >= 0 {
		o.Effects = append(o.Effects, rules.Walked{Unit: unitOf(f.Actor), Path: path[:i+1]})
	}
	f.Board.Units[u].At = landed
	if err := guards(&o); err != nil {
		return o, err
	}
	if c.Enabled {
		c.Status = StUsedJetpack
	}
	endActions(s)
	return o, nil
}

func indexOf(ps []rules.BoardPos, p rules.BoardPos) int {
	for i := len(ps) - 1; i >= 0; i-- {
		if ps[i] == p {
			return i
		}
	}
	return -1
}

// hasJetpack is FindReadiedItem(cr, C_Equip.JETPACK).
func (r *Rules) hasJetpack(c *Character) bool {
	_, _, _, ok := r.readied(c, xxvc.SlotJetpack)
	return ok
}

// walk is CombatMoveCreature along path, a step at a time: the step's cost
// off the movement (all of it when it costs more), the attacks drawn reset,
// then GuardersAttack; a mover downed or helpless stops.
func (r *Rules) walk(f rules.Fight, id rules.CharID, c *Character, s *combatState, at rules.BoardPos, path []rules.BoardPos, rng *rules.RNG) (rules.Outcome, error) {
	var o rules.Outcome
	f.Board.Units = slices.Clone(f.Board.Units)
	u := slices.IndexFunc(f.Board.Units, func(u rules.BoardUnit) bool { return u.ID == id })
	for _, p := range path {
		cost := stepCost(rules.DirectionTo(at.X, at.Y, p.X, p.Y), terrainAt(f.Board, p))
		s.movement = max(0, s.movement-cost)
		at = p
		f.Board.Units[u].At = p
		s.attacksDrawn = 0
		o.Effects = append(o.Effects, rules.Walked{Unit: unitOf(id), Path: []rules.BoardPos{p}})
		g, err := r.guardersAttack(f, id, c, rng)
		if err != nil {
			return o, err
		}
		o.Effects = append(o.Effects, g.Effects...)
		o.Effects = append(o.Effects, r.considerTerrain(f.Board, id, c, s, rng)...)
		if !c.Enabled || helpless(c) {
			s.movement = 0
			break
		}
	}
	return o, nil
}

// guardersAttack is cbtacts.pas GuardersAttack: each guarding enemy a line
// reaches within 60 (CountTargets order), not helpless, with a ranged weapon
// or adjacent, whose weapon's line reaches the mover (or the mover flies a
// jetpack), attacks it once and stops guarding, until the mover is down.
func (r *Rules) guardersAttack(f rules.Fight, id rules.CharID, c *Character, rng *rules.RNG) (rules.Outcome, error) {
	var o rules.Outcome
	at, _ := unitAt(f, id)
	for _, t := range r.targetList(f, id, c, 60, false) {
		if !c.Enabled {
			break
		}
		g, gs := r.chars[t.id], r.fight[t.id]
		if gs == nil || !gs.guarding || helpless(g) {
			continue
		}
		gat, _ := unitAt(f, t.id)
		if !r.hasRangedWeapon(g) && distance(gat, at) > 1 {
			continue
		}
		if _, _, ok, _ := walkLine(f.Board, gat, at, r.readiedWeaponRange(g), true, false); !ok && c.Status != StUsingJetpack {
			continue
		}
		a, err := r.attack(f, t.id, id, rng)
		if err != nil {
			return o, err
		}
		o.Effects = append(o.Effects, a.Effects...)
		gs.guarding = false
	}
	return o, nil
}

// hasRangedWeapon is CharHlp.HasRangedWeapon: a readied weapon of range over 1.
func (r *Rules) hasRangedWeapon(c *Character) bool {
	_, _, ty, ok := r.readied(c, xxvc.SlotWeapon)
	return ok && ty.Range > 1
}
