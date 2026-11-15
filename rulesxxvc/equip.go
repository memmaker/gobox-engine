package rulesxxvc

import (
	"gobox/rules"
	"gobox/rules/xxvc"
)

// The item types and effects EquipWeaponAndArmor tests (c_item.pas,
// combatai.pas, c_effect.pas).
const (
	plasmaThrowerType      = 17
	explosiveGrenadeType   = 33
	stunGrenadeType        = 35
	dazzleGrenadeType      = 36
	rocketLauncherType     = 18
	gasGrenadeType         = 34
	chaffGrenadeType       = 37
	aerosolMistGrenadeType = 38
)

// equipWeaponAndArmor is combatai.pas EquipWeaponAndArmor: the AI readies its
// best armour (class plus plus) and its best weapon by CalcWeaponPriority
// against its target, a grenade launcher beside a grenade, and recalculates.
// firing is mode <> 0. It reports a heavy weapon readied (the turn ends).
// The line to the target (WalkLine, range 255) crossing aerosol mist shifts
// lasers by 3, crossing chaff rockets by 3.
func (r *Rules) equipWeaponAndArmor(f rules.Fight, id rules.CharID, c *Character, s *combatState, firing bool, rng *rules.RNG) bool {
	t := r.chars[s.target]
	laserShift, rocketShift, microwaveShift := 0, 0, 0
	if t != nil {
		var seen lineSeen
		if at, ok := unitAt(f, id); ok {
			if tat, ok := unitAt(f, s.target); ok {
				_, _, _, seen = walkLine(f.Board, at, tat, 255, true, false)
			}
		}
		if seen.mist || hasEffect(t, EffImmuneToLaser) {
			laserShift = 3
		}
		if seen.chaff {
			rocketShift = 3
		} else if t.Armor.ECM || hasEffect(t, EffImmuneToRocket) {
			rocketShift = 2
		}
		if t.Armor.Body || hasEffect(t, EffImmuneToMicrowave) {
			microwaveShift = 3
		}
	}
	u := c.Unarmed[0]
	bestPriority := uint8Wrap(int(u.Sides*u.N + u.Bonus*2))
	bestWeapon, bestArmor, launcher := -1, -1, -1
	launcherReadied := false
	bestArmorValue := 0
	for i, ci := range c.Items {
		tt, ty, ok := r.itemType(ci)
		if !ok {
			continue
		}
		switch ty.Slot {
		case xxvc.SlotWeapon:
			p := r.weaponPriority(f, id, c, s, t, tt, ty, ci, firing, rng)
			switch ty.WeaponClass {
			case xxvc.Rocket:
				p >>= rocketShift
			case xxvc.Laser:
				p >>= laserShift
			case xxvc.Microwave:
				p >>= microwaveShift
			}
			if p > bestPriority {
				bestPriority, bestWeapon = p, i
			}
		case xxvc.SlotArmor:
			// weaponClass holds the class plus 50: the larger value wins
			if v := int(ty.ArmorClass) + plus(ci, tt); bestArmor < 0 || bestArmorValue < v {
				bestArmor, bestArmorValue = i, v
			}
		case xxvc.SlotGrenadeLauncher:
			if ci.Readied {
				launcherReadied = true
			} else {
				launcher = i
			}
		}
	}
	oldArmor, oldWeapon := r.readiedIndex(c, xxvc.SlotArmor), r.readiedIndex(c, xxvc.SlotWeapon)
	if bestArmor != oldArmor {
		if oldArmor >= 0 {
			c.Items[oldArmor].Readied = false
		}
		if bestArmor >= 0 {
			c.Items[bestArmor].Readied = true
		}
	}
	heavy := false
	if bestWeapon != oldWeapon && bestWeapon >= 0 {
		_, ty, _ := r.itemType(c.Items[bestWeapon])
		heavy = ty.Heavy
	}
	if oldWeapon >= 0 {
		c.Items[oldWeapon].Readied = false
	}
	if bestWeapon >= 0 {
		c.Items[bestWeapon].Readied = true
		if _, ty, _ := r.itemType(c.Items[bestWeapon]); !heavy && launcher >= 0 && !launcherReadied && ty.WeaponClass == xxvc.Grenade {
			c.Items[launcher].Readied = true
		}
	}
	if oldArmor != bestArmor || oldWeapon != bestWeapon || heavy {
		r.RecalcStats(c)
		r.recalcAttacksLeft(f, c, s)
	}
	return heavy
}

// weaponPriority is EquipWeaponAndArmor's CalcWeaponPriority (Byte sums).
func (r *Rules) weaponPriority(f rules.Fight, id rules.CharID, c *Character, s *combatState, t *Character, tt xxvc.ItemTemplate, ty xxvc.ItemType, ci Carried, firing bool, rng *rules.RNG) int {
	if ty.WeaponClass > xxvc.Melee && quantity(ci, tt) == 0 {
		return 0
	}
	p := int(ty.Damage.Sides*ty.Damage.N + ty.Damage.Bonus*2)
	switch tt.Type {
	case sonicStunnerType:
		p = 12
		if t != nil && hasEffect(t, EffStunned) {
			p = 0
		}
	case plasmaThrowerType, rocketLauncherType, gasGrenadeType, stunGrenadeType, dazzleGrenadeType, chaffGrenadeType, aerosolMistGrenadeType:
		cloud := tt.Type == chaffGrenadeType || tt.Type == aerosolMistGrenadeType
		p = 0
		if firing && (!c.team() || !cloud) && (!f.Board.ZeroG || !cloud) { // EvaluateItem
			a, _ := r.selectAreaTarget(f, id, c, s, tt, ty, rng)
			p = areaChoice(a, tt.Type, rng)
		}
	}
	p += c.Specialized(tt.Type)*3 + plus(ci, tt)*3
	if ty.WeaponClass == xxvc.Melee {
		p += int(strDamageAdjustments[score(c, rules.Strength)])*2 + int(strToHitAdjustments[score(c, rules.Strength)])*2
	} else {
		p += DexMissileAdjust(score(c, rules.Dexterity)) * 2
	}
	mult := max(int(ty.AttacksDoubled>>1), 1)
	if !firing && ty.WeaponClass >= xxvc.RocketLauncher {
		return 0
	}
	return uint8Wrap(p * mult)
}

// uint8Wrap is a Byte assignment's wrap.
func uint8Wrap(v int) int { return int(uint8(v)) }

// readiedIndex is FindReadiedItem: the last readied item of slot, or -1.
func (r *Rules) readiedIndex(c *Character, slot xxvc.EquipSlot) int {
	at := -1
	for i, ci := range c.Items {
		if _, ty, ok := r.itemType(ci); ok && ci.Readied && ty.Slot == slot {
			at = i
		}
	}
	return at
}

// recalcAttacksLeft is cbtacts.pas RecalcAttacksLeft for phase 1: the readied
// weapon's rate (or baseScale), halved over the round's odd bit, a ranged
// weapon's capped by its shots; never more than were left once it attacked.
func (r *Rules) recalcAttacksLeft(f rules.Fight, c *Character, s *combatState) {
	rate := c.BaseScale
	ci, tt, ty, ok := r.readied(c, xxvc.SlotWeapon)
	if ok {
		rate = int(ty.AttacksDoubled)
	}
	rate = attacksLeft(attackRate(c, rate), f.Round)
	if ok && ty.WeaponClass != xxvc.Melee {
		rate = min(rate, quantity(ci, tt))
	}
	if !s.hasAttacked || rate < s.attacks[0] {
		s.attacks[0] = rate
	}
}

// attackRate is attackRate after TriggerEffectsWithCondition(MOVE_ATTACK_RATE):
// DIZZY's handler (effhndlr.pas EffDizzy) halves it.
func attackRate(c *Character, rate int) int {
	if hasEffect(c, EffDizzy) {
		rate /= 2
	}
	return rate
}

// readiedWeaponRange is viewchar.pas ReadiedWeaponRange: the readied weapon's
// range (0 or 255 read as 1), 12 for a grenade beside a readied launcher; 1
// unarmed.
func (r *Rules) readiedWeaponRange(c *Character) int {
	_, _, ty, ok := r.readied(c, xxvc.SlotWeapon)
	return r.weaponRange(c, ty, ok)
}

// weaponRange is ReadiedWeaponRange with ty as the readied weapon (armed).
func (r *Rules) weaponRange(c *Character, ty xxvc.ItemType, armed bool) int {
	if !armed {
		return 1
	}
	rng := int(ty.Range)
	if rng == 0 || rng == 255 {
		rng = 1
	}
	if ty.WeaponClass == xxvc.Grenade && r.readiedIndex(c, xxvc.SlotGrenadeLauncher) >= 0 {
		rng = 12
	}
	return rng
}
