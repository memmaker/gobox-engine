// Package rules holds the rules the Gold Box games share, copied from the
// games' engines and kept behaviour-identical, including the exact order of
// RNG calls. The games pass in their side effects (Hooks) and tables; what an
// original does differently stays in that game. The package never knows which
// game runs.
package rules

import (
	"gobox/char"
	"gobox/core/goldbox"
)

// Hooks are the game's side effects around a roll. Nil funcs do nothing.
type Hooks struct {
	Begin  func()              // AttackHits: before the roll (attacking ends invisibility)
	Groups func()              // AttackHits: after a roll that is not a miss (attacker then defender affects)
	Group  func()              // SavingThrow, TrapHits: the save / defender affect group
	Hit    func(over int) bool // AttackHits: extra rule-set veto (an alternative rule set's hit test); nil accepts
}

// Status values shared by the games' status tables.
const (
	StatusOkay        = 0
	StatusAnimated    = 1
	StatusRunning     = 3
	StatusUnconscious = 4
	StatusDying       = 5
	StatusDead        = 6
	StatusStoned      = 7
)

// Class indexes of the shared class order (cleric druid fighter paladin ranger ...).
const (
	ClsFighter = 2
	ClsPaladin = 3
	ClsRanger  = 4
)

// hitRoll rolls the d20 the attack and trap rolls share: 1 misses, 20 counts 100.
func hitRoll(rv *goldbox.Roll, rng goldbox.RNG) bool {
	rv.HitRoll = int8(rng.Dice(1, 20))
	if rv.HitRoll <= 1 {
		return false
	}
	if rv.HitRoll == 20 {
		rv.HitRoll = 100
	}
	return true
}

// AttackHits: d20 + THAC0 + adj (the side's modifier) must reach need.
func AttackHits(rv *goldbox.Roll, rng goldbox.RNG, att char.Stats, adj int, need int, h Hooks) bool {
	if h.Begin != nil {
		h.Begin()
	}
	if !hitRoll(rv, rng) {
		return false
	}
	if h.Groups != nil {
		h.Groups()
	}
	over := int(rv.HitRoll) + att.THAC0() + adj - int(byte(need))
	return rv.HitRoll >= 0 && over >= 0 && (h.Hit == nil || rv.HitRoll == 100 || h.Hit(over))
}

// TrapHits: a d20 with the defender's affects plus bonus must beat AC 0.
func TrapHits(rv *goldbox.Roll, rng goldbox.RNG, c char.Stats, bonus byte, h Hooks) bool {
	if !hitRoll(rv, rng) {
		return false
	}
	if h.Group != nil {
		h.Group()
	}
	return rv.HitRoll >= 0 && int(rv.HitRoll)+int(bonus) > c.AC(0)
}

// SavingThrow: 1 always fails, 20 always saves, otherwise d20 + bonus + the
// save bonus must reach the target for the type. One path for all games: the
// total is compared signed (por's fix), so a penalty below 0 does not wrap.
func SavingThrow(rv *goldbox.Roll, rng goldbox.RNG, c char.Stats, typ, bonus int, h Hooks) bool {
	rv.SaveMade = true
	rv.Save = byte(rng.Dice(1, 20))
	switch rv.Save {
	case 1:
		rv.SaveMade = false
	case 20:
		rv.SaveMade = true
	default:
		rv.Save = byte(int(rv.Save) + int(int8(c.SaveBonus())) + int(int8(bonus)))
		rv.SaveType = byte(typ)
		if h.Group != nil {
			h.Group()
		}
		rv.SaveMade = c.Save(int(int8(typ))) <= int(int8(rv.Save))
	}
	return rv.SaveMade
}

// Thac0At reads the THAC0 table like the original: flat, indexed
// class*width + level (levels past the row run into the next class's row).
func Thac0At(table [][]byte, class int, level int8) byte {
	w := len(table[0])
	i := class*w + int(level)
	return table[i/w][i%w]
}

// Strength limits.
const (
	StrMax        = 18
	StrEncodeBase = 100
)

// StrengthCode: one byte for a strength: 18/xx as xx+1 (18/00 is 101), others +100.
func StrengthCode(str, str00 byte) byte {
	if str == StrMax {
		return str00 + 1
	}
	return str + StrEncodeBase
}

// DamagePlayer takes dmg off the hit points: down to 0, then unconscious,
// dying (bleeding rounds) or dead at 10 below. over is the overkill; dying
// reports that the status is now Dying (the game stores over in its combat
// block); down that the member is out (the game updates its side count and
// delay when fighting).
func DamagePlayer(c char.Stats, dmg byte) (over byte, dying, down bool) {
	hp := byte(c.HP())
	var left byte
	if hp >= dmg {
		left = hp - dmg
	} else {
		over = dmg - hp
	}
	status := c.Status()
	switch {
	case over > 9 || left == 0 && status == StatusAnimated:
		status = StatusDead
	case over > 0:
		status = StatusDying
		dying = true
	case left == 0:
		status = StatusUnconscious
	}
	c.SetStatus(status)
	if status == StatusOkay || status == StatusAnimated {
		c.SetHP(int(left))
		return over, false, false
	}
	c.SetOkay(0)
	c.SetHP(0)
	return over, dying, true
}
