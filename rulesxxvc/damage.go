package rulesxxvc

import (
	"gobox/rules"
	"gobox/rules/xxvc"
)

// WeaponDamage is cbtacts.pas CalcWeaponDamage's roll (BR1 = BR2): the dice
// plus the modifier, at least 0, in effectDamage's Byte; a backstab multiplies
// it by (level+7) div 4 for a rogue, by 2 for anyone else. The effect triggers
// and Shielded follow it.
func WeaponDamage(r *rules.RNG, d rules.Dice, backstab, rogue bool, level int) int {
	computed := rules.RNGDice(r, int(d.N), int(d.Sides)) + int(d.Bonus)
	var dmg uint8
	if computed > 0 {
		dmg = uint8(computed)
	}
	if backstab {
		if rogue {
			dmg *= uint8((level + 7) / 4)
		} else {
			dmg <<= 1
		}
	}
	return int(dmg)
}

// Defences are what CalcWeaponDamage tests of the defender and the shooter's cell.
type Defences struct {
	SawChaff, SawMist bool // the attacker's cell, or earlier in the fight
	ECM, Fields, Body bool // the defender's armorFlags
}

// Shielded is CalcWeaponDamage's tail: a rocket stops on chaff, on ECM half
// the time, on fields three times in four (rolled in that order, Pascal's
// short-circuit or); a laser on aerosol mist; a microwave on body armor.
func Shielded(r *rules.RNG, c xxvc.WeaponClass, d Defences) bool {
	switch c {
	case xxvc.Rocket:
		return d.SawChaff || (d.ECM && PercentileRoll(r, 50)) || (d.Fields && PercentileRoll(r, 75))
	case xxvc.Laser:
		return d.SawMist
	case xxvc.Microwave:
		return d.Body
	}
	return false
}

// dexMissileAdjustments is gdconsts.pas's table by dexterity 1..22 (BR1 = BR2).
var dexMissileAdjustments = [23]int8{0, -6, -4, -3, -2, -1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 2, 3, 3, 4, 4}

// DexMissileAdjust is charhlp.pas DexMissileAdjust.
func DexMissileAdjust(dex int) int { return int(dexMissileAdjustments[dex]) }

// Initiative is cbtacts.pas BeginRound's roll for an enabled creature: d10 +
// dexterity's missile adjustment + 1, at least 2, 6 less for a surprised side,
// 0 outside 0..20; then the tie-breaker d127. Disabled: 0, no roll.
func Initiative(r *rules.RNG, enabled bool, dex int, surprised bool) (initiative, tieBreaker int) {
	if !enabled {
		return 0, 0
	}
	i := int8(rules.RNGDice(r, 1, 10) + DexMissileAdjust(dex) + 1)
	if i < 2 {
		i = 2
	}
	if surprised {
		i -= 6
	}
	if i < 0 || i > 20 {
		i = 0
	}
	return int(i), rules.RNGDice(r, 1, 127)
}
