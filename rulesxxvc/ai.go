package rulesxxvc

import (
	"slices"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// maxAiSteps is AiMoveAttack's rounds limit: past it the turn finishes.
const maxAiSteps = 20

// aiDirOffsets is gdconsts.pas aiDirOffsets: per approach pattern (1..6) the
// turns, in eighths, AiMoveStep tries in order (8 straight on).
var aiDirOffsets = [6][5]int{
	{8, 7, 6, 1, 2}, {8, 1, 2, 7, 6}, {7, 1, 8, 6, 2}, {1, 7, 8, 2, 6}, {8, 7, 6, 5, 4}, {8, 1, 2, 3, 4},
}

// Decide is combatai.pas CombatAiControl, one action per call. The turn
// start rolls the approach pattern and CheckMorale; each pass of
// CombatAiControl's loop runs PrepareAiAction mode 1 (the target, then
// EquipWeaponAndArmor) before AiMoveAttack, whose rounds the following calls
// play: the flight steps, an attack on a target in reach, else one AiMoveStep
// toward it, else FinishRound. An attack spends all attacks or downs its
// target (RollAttacks); a downed one with attacks left stops AiMoveAttack and
// starts the next pass.
// An area weapon in reach fires at SelectAreaTarget's cell (Attack At).
// ponytail: medics' FindPatients is not ported.
func (r *Rules) Decide(f rules.Fight, id rules.CharID, rng *rules.RNG) rules.Action {
	f.Board = r.board(f.Board)
	c, s := r.chars[id], r.fight[id]
	if c == nil || s == nil {
		return rules.EndTurn{}
	}
	if !s.aiBegun {
		s.aiBegun = true
		if !c.Enabled || c.Status == StDodging || c.Status == StSprinting {
			endActions(s)
		}
		if s.approach == 0 || s.approach > 4 || rules.RNGDice(rng, 1, 4) == 1 {
			if rules.RNGDice(rng, 1, 8) == 8 {
				s.approach = rules.RNGDice(rng, 1, 2) + 4
			} else {
				s.approach = rules.RNGDice(rng, 1, 4)
			}
		}
		if s.initiative > 0 && r.checkMorale(f, c, s) {
			return rules.EndTurn{}
		}
	}
	if s.aiPrepare { // a CombatAiControl pass: PrepareAiAction(cr, 1), AiMoveAttack begins
		s.aiPrepare, s.aiSteps, s.lastMoveDir, s.stuck = false, 0, centerDir, 0
		if r.prepareAiAction(f, id, c, s, rng) {
			return rules.EndTurn{} // a heavy weapon readied ends the turn
		}
		if hasEffect(c, EffSpitAcid) || hasEffect(c, EffStrikeAdjacent) {
			s.specialDue = true // AiMoveAttack's MONSTER_SPECIAL, resolved as its own action
			return rules.Maneuver{ID: monsterSpecial}
		}
	}
	s.aiSteps++
	if s.initiative <= 0 || !c.Enabled || s.aiSteps > maxAiSteps {
		return r.finishRound(c, s)
	}
	at, ok := unitAt(f, id)
	if !ok {
		return r.finishRound(c, s)
	}
	// AiMoveAttack: a fleeing one spends its movement (while any is left,
	// half steps) on AiMoveStep first, so it runs even with no target;
	// AiMoveStep's own gate failing finishes the round there.
	if s.fleeing && s.movement > 0 {
		if s.movement/2 == 0 || !r.stepsOn(f, c, s, rng) {
			return r.finishRound(c, s)
		}
		s.approach = rules.RNGDice(rng, 1, 2)
		if act := r.aiStep(f, id, c, s, at, rng); act != nil {
			return act
		}
		return r.Decide(f, id, rng) // the flight loop goes on
	}
	// isTargetInReach: a line to the target within the weapon's range, and
	// either little movement left, the target nearer than the range, or an
	// area weapon.
	reach := r.readiedWeaponRange(c)
	target, inReach := s.target, false
	if t := r.chars[target]; t != nil && t.Enabled && t.team() != c.team() {
		if tat, ok := unitAt(f, target); ok {
			if _, d, ok, _ := walkLine(f.Board, at, tat, reach, true, false); ok && d/2 <= reach &&
				(s.movement < 4 || reach > d || r.hasAreaWeapon(c)) {
				inReach = true
			}
		}
	}
	if !inReach {
		if s.movement > 4 {
			reach = 1
		}
		if near := r.pickTarget(f, id, c, reach, rng); near != rules.NoChar {
			target, inReach = near, true // CanTarget: an enemy
		} else if r.prepareTarget(f, id, c, s, rng) && s.movement/2 > 0 && r.stepsOn(f, c, s, rng) {
			// PrepareAiAction mode 0, then AiMoveStep
			if s.fleeing {
				s.approach = rules.RNGDice(rng, 1, 2)
			}
			if act := r.aiStep(f, id, c, s, at, rng); act != nil {
				return act
			}
			return r.Decide(f, id, rng)
		} else {
			return r.finishRound(c, s)
		}
	}
	if r.hasAreaWeapon(c) { // SelectAreaTarget mode 1
		_, tt, ty, _ := r.readied(c, xxvc.SlotWeapon)
		if a, _ := r.selectAreaTarget(f, id, c, s, tt, ty, rng); a.weight > 0 {
			return rules.Attack{At: &a.at} // AttackGroundTarget, then EndActions
		}
		if r.equipWeaponAndArmor(f, id, c, s, false, rng) {
			return rules.EndTurn{} // a heavy weapon readied
		}
		return r.Decide(f, id, rng) // AiMoveAttack's loop goes on with the weapon readied
	}
	s.target, s.aiPrepare = target, true // MakeAttack; it ends the turn or downs the target: the next pass
	return rules.Attack{Target: target}
}

// hasAreaWeapon is CharHlp.HasAreaWeapon: the readied weapon is a rocket
// launcher, grenade or plasma weapon.
func (r *Rules) hasAreaWeapon(c *Character) bool {
	_, _, ty, ok := r.readied(c, xxvc.SlotWeapon)
	return ok && ty.WeaponClass >= xxvc.RocketLauncher
}

// prepareTarget is PrepareAiAction's target part: one on its own side or down
// is dropped (CanTarget is the side test) and another picked among all
// enemies a line reaches; on the second pass (distMode 1) any line counts.
func (r *Rules) prepareTarget(f rules.Fight, id rules.CharID, c *Character, s *combatState, rng *rules.RNG) bool {
	if t := r.chars[s.target]; t != nil && t.Enabled && t.team() != c.team() {
		return true
	}
	s.target = rules.NoChar
	for _, anyLine := range []bool{false, true} {
		if ts := r.targetList(f, id, c, 255, anyLine); len(ts) > 0 {
			s.target = ts[rules.RNGDice(rng, 1, len(ts))-1].id
			return true
		}
	}
	return false
}

// prepareAiAction is PrepareAiAction mode 1: prepareTarget, then with a
// target EquipWeaponAndArmor mode 1. It reports a heavy weapon readied.
func (r *Rules) prepareAiAction(f rules.Fight, id rules.CharID, c *Character, s *combatState, rng *rules.RNG) bool {
	return r.prepareTarget(f, id, c, s, rng) && r.equipWeaponAndArmor(f, id, c, s, true, rng)
}

// finishRound is FinishRound: a helpless one or one without initiative ends,
// the others guard.
func (r *Rules) finishRound(c *Character, s *combatState) rules.Action {
	if helpless(c) || s.initiative <= 0 {
		return rules.EndTurn{}
	}
	return rules.Guard{}
}

// pathTarget is TTarget: a combatant BuildTargetList reached, its line's
// distance (half steps: 2 straight, 3 diagonal) and direction.
type pathTarget struct {
	id   rules.CharID
	dist int
	dir  int
}

// targetList is pathing.pas BuildTargetList(x, y, rng, 255, size) as
// CountTargets filters it: every enabled enemy of id a line from one of id's
// icon cells reaches within rng (the shortest kept, its direction the first
// IsDirToward finds), nearest first (SortTargets). anyLine is distMode 1.
func (r *Rules) targetList(f rules.Fight, id rules.CharID, c *Character, rng int, anyLine bool) []pathTarget {
	i := slices.IndexFunc(f.Board.Units, func(u rules.BoardUnit) bool { return u.ID == id })
	if i < 0 {
		return nil
	}
	return r.lineTargets(f.Board, cellsOf(f.Board.Units[i]), rng, !anyLine, func(e *Character) bool { return e.team() != c.team() })
}

// lineTargets is BuildTargetList's walk from the cells mine: every enabled
// combatant keep accepts that a line reaches within rng, nearest first.
func (r *Rules) lineTargets(b rules.Board, mine []rules.BoardPos, rng int, stopsAtWalls bool, keep func(*Character) bool) []pathTarget {
	var ts []pathTarget
	for _, u := range b.Units {
		if e := r.chars[u.ID]; e == nil || !e.Enabled || !keep(e) {
			continue
		}
		best := pathTarget{id: u.ID, dist: 255}
		found := false
		for _, tc := range cellsOf(u) {
			for _, mc := range mine {
				_, d, ok, _ := walkLine(b, mc, tc, rng, stopsAtWalls, false)
				found = found || ok
				if ok && d < best.dist {
					best.dist = d
					best.dir = 0
					for !rules.InArc(b.Size.X, b.Size.Y, mc.X, mc.Y, tc.X, tc.Y, best.dir) {
						best.dir++
					}
				}
			}
		}
		if found {
			ts = append(ts, best)
		}
	}
	rules.NearSort(ts, func(t pathTarget) (byte, byte) { return byte(min(t.dist, 255)), byte(t.dir) })
	return ts
}

// pickTarget is CountTargets(cr, rng) and its RollDice(1, count) pick.
func (r *Rules) pickTarget(f rules.Fight, id rules.CharID, c *Character, rng int, dice *rules.RNG) rules.CharID {
	ts := r.targetList(f, id, c, rng, false)
	if len(ts) == 0 {
		return rules.NoChar
	}
	return ts[rules.RNGDice(dice, 1, len(ts))-1].id
}

// aiStep is AiMoveStep: toward the target, or away from the fight when
// fleeing, the approach's turns tried in order (AiCanStep over
// ProbeDirection: on the board, passable, nobody there, its stepCost below
// the movement left); a fleeing one at the edge flees. Finding no way, or
// reversing the last step, counts as stuck and turns the approach; the
// second time this pass the target is dropped and sought again
// (PrepareAiAction mode 1), the third ends the movement. A nil action is no
// step this round: AiMoveAttack's loop goes on.
func (r *Rules) aiStep(f rules.Fight, id rules.CharID, c *Character, s *combatState, at rules.BoardPos, rng *rules.RNG) rules.Action {
	var base int
	if s.fleeing {
		base = fleeDirection(f.Terms.PartyFacing, c.team())
	} else if tat, ok := unitAt(f, s.target); ok {
		base = rules.DirectionTo(at.X, at.Y, tat.X, tat.Y)
	}
	step := -1
	for _, turn := range aiDirOffsets[s.approach-1] { // AiCanStep
		dir := (base + turn) % 8
		blocked, t, _ := r.probeDirection(f.Board, id, dir)
		if t == 0 {
			if s.fleeing { // the edge: AttemptFlee
				return rules.Flee{}
			}
			continue
		}
		if terrainRow(t).cost != impassable && !blocked && stepCost(dir, t) < s.movement {
			step = dir
			break
		}
	}
	if step < 0 || s.lastMoveDir == (step+4)%8 {
		s.stuck++
		s.approach = s.approach%len(aiDirOffsets) + 1
		if s.stuck > 1 {
			s.target = rules.NoChar
			if s.stuck > 2 {
				s.movement = 0
				return nil
			}
			if heavy := r.prepareAiAction(f, id, c, s, rng); heavy {
				return rules.EndTurn{}
			}
			if s.target == rules.NoChar {
				return r.finishRound(c, s)
			}
		}
	}
	if step < 0 {
		return nil
	}
	s.lastMoveDir = step
	return rules.Move{To: at.Add(dirStep[step])}
}

// centerDir is C_Dir.CENTER, no direction.
const centerDir = 8

// dirStep is C_Dir's eight directions, north first, clockwise.
var dirStep = [8]rules.BoardPos{{X: 0, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: -1, Y: 1}, {X: -1, Y: 0}, {X: -1, Y: -1}}

// free: on the board, passable and nobody standing there.
func (r *Rules) free(f rules.Fight, p rules.BoardPos) bool {
	t := terrainAt(f.Board, p)
	if t == 0 || terrainRow(t).cost == impassable {
		return false
	}
	_, taken := r.occupant(f.Board, p)
	return !taken
}

func unitAt(f rules.Fight, id rules.CharID) (rules.BoardPos, bool) {
	for _, u := range f.Board.Units {
		if u.ID == id {
			return u.At, true
		}
	}
	return rules.BoardPos{}, false
}

func distance(a, b rules.BoardPos) int {
	d := b.Sub(a)
	return max(rules.Abs(d.X), rules.Abs(d.Y))
}
