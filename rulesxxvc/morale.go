package rulesxxvc

import (
	"gobox/rules"
	"gobox/rules/xxvc"
)

// nonPlayerMorale is C_Npc: the npc byte's morale is its low seven bits, read
// doubled; above maxMorale it counts as none.
const maxMorale = 102

// checkMorale is combatai.pas CheckMorale: a companion tests its morale
// against its wounds, then the hostiles' remaining hit points against the
// fight's moraleBase (FightTerms.Morale), which a TEAM companion always
// fails. Failing both, it flees when no enemy is faster (halved movement),
// else with INT over 5 it surrenders (TEAM: fleeing; hostile: unconscious)
// and its turn ends. MORALE_CHECK has no handlers in BR (fxengine.pas), so
// the two TriggerEffectsWithCondition calls do nothing.
func (r *Rules) checkMorale(f rules.Fight, c *Character, s *combatState) bool {
	s.fleeing = false
	if !c.Companion {
		return false
	}
	r.morale = c.Morale * 2
	if r.morale > maxMorale {
		r.morale = 0
	}
	if r.morale >= 100-c.HP*100/c.MaxHP && r.morale != 0 {
		return false
	}
	r.morale = r.hpPercent
	if r.morale >= 100-int(f.Terms.Morale) && r.morale != 0 && c.Side == rules.SideMonsters {
		return false
	}
	if r.highestEnemyMovement(f, c) <= r.RecalcMovement(c)/2 {
		s.fleeing = true
		return false
	}
	if score(c, rules.Intelligence) <= 5 {
		return false
	}
	c.Status = StUnconscious
	if c.team() {
		c.Status = StFleeing
	}
	c.Enabled = false // PutOutOfPlayWithStatus 'Surrenders'
	endActions(s)
	return true
}

// hostileHpPercent is cbtacts.pas UpdateHostileHpPercent: the enabled
// hostiles' hit points as a percentage, in steps of 5, of all the hostiles'
// maximum; 100 with none (Word sums: no overflow below 3277 hit points).
// Kept in r.hpPercent: the reference updates it only at the fight's start
// and each round's end, so hits within a round do not move it.
func (r *Rules) hostileHpPercent(f rules.Fight) int {
	hp, maxHP := 0, 0
	for _, id := range r.fighters(f) {
		if c := r.chars[id]; c.Side == rules.SideMonsters {
			if c.Enabled {
				hp += c.HP
			}
			maxHP += c.MaxHP
		}
	}
	if maxHP == 0 {
		return 100
	}
	return 20 * hp / maxHP * 5
}

// highestEnemyMovement is cbtacts.pas HighestEnemyMovement: the fastest
// enabled enemy's RecalcMovement, halved.
func (r *Rules) highestEnemyMovement(f rules.Fight, c *Character) int {
	best := 0
	for _, id := range r.fighters(f) {
		if e := r.chars[id]; e.Enabled && e.team() != c.team() {
			best = max(best, r.RecalcMovement(e)/2)
		}
	}
	return best
}

// stepsOn is AiMoveStep's gate: a player character, a non-player one while
// the hostiles' hit points stay within morale plus d100, or a hostile; and
// fleeing, armoured or no medic.
func (r *Rules) stepsOn(f rules.Fight, c *Character, s *combatState, rng *rules.RNG) bool {
	willing := !c.NonPlayer || r.hpPercent <= r.morale+rules.RNGDice(rng, 1, 100) || c.Side == rules.SideMonsters
	_, _, _, armoured := r.readied(c, xxvc.SlotArmor)
	return willing && (s.fleeing || armoured || c.Career != xxvc.Medic)
}

// fleeDirection is AiMoveStep's flight: away from the party's facing, the
// other way for the TEAM side (Byte arithmetic, mod 8).
func fleeDirection(facing rules.Facing, team bool) int {
	d := uint8(facing) - (uint8(facing)+2)%4/2
	if team {
		d += 4
	}
	return int(d % 8)
}
