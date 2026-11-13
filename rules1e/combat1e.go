package rules1e

import "gobox/core"

// naturalAttackMax counts as 100, so that a 20 always reaches its need.
const naturalAttackMax = 100

// AttackDie is the die of FxEngine.AttackRoll and AttackRollNoAttacker: d20, a
// natural 20 counting 100. The effects a roll triggers then change it, by the
// caller, before AttackHits or TrapHits judges it.
func AttackDie(rng *core.RNG) int {
	roll := RollDice(rng, 1, 20)
	if roll == 20 {
		return naturalAttackMax
	}
	return roll
}

// AttackMissed: a natural 1 misses before any effect sees the roll.
func AttackMissed(die int) bool { return die <= 1 }

// AttackHits is AttackRoll's judgement of a roll effects may have changed.
func AttackHits(roll, thac0, modifier, need int) bool {
	return roll >= 0 && roll+thac0+modifier >= need
}

// TrapHits is AttackRollNoAttacker's judgement: a roll with no attacker must beat
// the defender's armour value.
func TrapHits(roll, bonus, armour int) bool { return roll >= 0 && roll+bonus > armour }

// SaveFixed is what a natural save die decides by itself: 1 fails, 20 succeeds.
func SaveFixed(die int) (made, fixed bool) {
	switch die {
	case 1:
		return false, true
	case 20:
		return true, true
	}
	return false, false
}

// SaveMade is RollSavingThrow's judgement. Signed, as ruling 1: the Pascal's Byte
// total wraps below 0 and saves.
func SaveMade(target, total int) bool { return target <= total }

// StrengthIndex is CharHlp.TranslatedStrength: strength on one scale, 18/xx as
// 18..23 and 19..25, which only effects give, as 24..30.
func StrengthIndex(strength, exceptional int) int {
	switch {
	case strength < 18:
		return strength
	case strength > 18:
		return strength + 5
	}
	switch {
	case exceptional == 0:
		return 18
	case exceptional <= 50:
		return 19
	case exceptional <= 75:
		return 20
	case exceptional <= 90:
		return 21
	case exceptional <= 99:
		return 22
	}
	return 23
}

// StrengthToHit is CharHlp.StrToHitAdjust of a StrengthIndex.
func StrengthToHit(i int) int {
	switch {
	case i >= 1 && i <= 3:
		return -3
	case i == 4 || i == 5:
		return -2
	case i == 6 || i == 7:
		return -1
	case i >= 17 && i <= 19:
		return 1
	case i >= 20 && i <= 22:
		return 2
	case i >= 23 && i <= 25:
		return 3
	case i == 26 || i == 27:
		return 4
	case i >= 28 && i <= 30:
		return i - 23
	}
	return 0
}

// StrengthDamage is CharHlp.StrDamageAdjust of a StrengthIndex.
func StrengthDamage(i int) int {
	switch {
	case i == 1 || i == 2:
		return -2
	case i >= 3 && i <= 5:
		return -1
	case i == 16:
		return 1
	case i >= 17 && i <= 19:
		return i - 16
	case i >= 20 && i <= 29:
		return i - 17
	case i == 30:
		return 14
	}
	return 0
}

// TurnNeed is the turn table: the roll needed to turn undead of a kind at a column,
// negative where a success destroys. Both count from 1.
func (t *Tables) TurnNeed(undead, column int) int { return int(t.TurnUndead[undead-1][column-1]) }

// TurnColumn is the table column of a cleric level: 1..8, 9..13 as one, then the last.
func TurnColumn(level int) int {
	switch {
	case level >= 1 && level <= 8:
		return level
	case level >= 9 && level <= 13:
		return 9
	}
	return 10
}

// AttacksFromRate is CbtActs.AttacksFromRate: a rate is in halves, the odd half
// falling on odd rounds.
func AttacksFromRate(rate, round int) int { return (rate + round&1) / 2 }
