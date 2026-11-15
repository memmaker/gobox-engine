package rulesxxvc

import (
	"gobox/rules"
)

// Difficulty is the game's difficulty setting (c_diff.pas).
type Difficulty int

// difficulty is the player's core.Options.Difficulty as BR's level.
func difficulty(d rules.Difficulty) Difficulty {
	switch d {
	case rules.DifficultyEasiest:
		return Novice
	case rules.DifficultyEasier:
		return Cadet
	case rules.DifficultyHarder:
		return Veteran
	case rules.DifficultyHardest:
		return Ace
	}
	return Trooper
}

const (
	Novice Difficulty = iota
	Cadet
	Trooper
	Veteran
	Ace
)

// naturalAttackMax is what a natural 20 counts as, so that it always reaches.
const naturalAttackMax = 100

// AttackDie is fxengine.pas AttackRoll's die: d20, a natural 20 counting 100.
// missed: a natural 1, before any effect sees the roll. The effects
// (ATTACK_ROLL_ATTACKER, COND_ATTACK_ROLL_DEF) then change the roll, by the
// caller, before AttackHits judges it.
func AttackDie(r *rules.RNG) (roll int, missed bool) {
	switch d := rules.RNGDice(r, 1, 20); d {
	case 1:
		return d, true
	case 20:
		return naturalAttackMax, false
	default:
		return d, false
	}
}

// Thac0Modifier is AttackRoll's side modifier: the zone's party modifier
// shifted by the difficulty for the team, the zone's monster modifier else.
func Thac0Modifier(team bool, d Difficulty, party, monster int) int {
	if !team {
		return monster
	}
	switch d {
	case Novice:
		return party + 2
	case Cadet:
		return party + 1
	case Veteran:
		return party - 1
	case Ace:
		return party - 2
	}
	return party
}

// AttackHits is AttackRoll's judgement of a roll effects may have changed.
func AttackHits(roll, attackBonus, modifier, target int) bool {
	return roll >= 0 && roll+attackBonus+modifier >= target
}

// ProjectileRangePenalty is cbtacts.pas ProjectileRangePenalty: +5 beyond
// three quarters of the weapon's range, +2 beyond half, none for range <= 1.
func ProjectileRangePenalty(distance, weaponRange int) int {
	switch {
	case weaponRange <= 1:
		return 0
	case distance > weaponRange*3>>2:
		return 5
	case distance > weaponRange>>1:
		return 2
	}
	return 0
}

// RangePenalty is cbtacts.pas ApplyRangePenalty's raise of the target value:
// +5 for a guarding shooter, else +2 past blocked terrain plus the
// projectile penalty.
func RangePenalty(guardingRanged, blocked bool, distance, weaponRange int) int {
	if guardingRanged {
		return 5
	}
	p := ProjectileRangePenalty(distance, weaponRange)
	if blocked {
		p += 2
	}
	return p
}

// ArmorValue is RollAttacks' armour value: the defender's rear armour
// (armorClass[1]) on a rear attack (IsRearAttack), else the front; a rogue
// attacking from behind (asked for, or a rear attack) adds 2.
func ArmorValue(front, rear int, rearAttack, fromBehind, rogue bool) int {
	a := front
	if rearAttack {
		a = rear
	}
	if rogue && (rearAttack || fromBehind) {
		a += 2
	}
	return a
}

// SaveDie is fxengine.pas RollSavingThrow's die: d20; a natural 1 fails and a
// natural 20 saves (fixed), else the roll plus the modifier goes to the
// effects (COND_SAVING_THROW), by the caller, before SaveMade judges it.
func SaveDie(r *rules.RNG, modifier int8) (roll int, made, fixed bool) {
	switch d := rules.RNGDice(r, 1, 20); d {
	case 1:
		return d, false, true
	case 20:
		return d, true, true
	default:
		return int(uint8(d + int(modifier))), false, false // saveRoll is a Byte
	}
}

// SaveMade is RollSavingThrow's judgement: the character's save value for the
// kind is at most the roll.
func SaveMade(save, roll int) bool { return save <= roll }
