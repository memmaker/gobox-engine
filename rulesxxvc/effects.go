package rulesxxvc

import (
	"slices"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// EffectID is C_Effect's effect number (TEffect.effectId).
type EffectID uint8

// The C_Effect numbers the rules read so far.
const (
	EffHelpless              EffectID = 1
	EffScared                EffectID = 2
	EffUnkillable            EffectID = 3
	EffBlind                 EffectID = 13
	EffStunned               EffectID = 14
	EffImmuneToMicrowave     EffectID = 18
	EffImmuneToRocket        EffectID = 20
	EffImmuneToLaser         EffectID = 24
	EffBlockScared           EffectID = 16
	EffBlockHelpless         EffectID = 19
	EffBlockDazzleGrenade    EffectID = 21
	EffBlockStunWeapons      EffectID = 22
	EffImmuneToHeatAndPlasma EffectID = 23
	EffDodgeSlugHits         EffectID = 25
	EffResistRangedHits      EffectID = 26
	EffResistMeleeHits       EffectID = 27
	EffDizzyingPoison        EffectID = 28
	EffDizzy                 EffectID = 29
	EffLethalPoison          EffectID = 30
	EffStrikeAdjacent        EffectID = 31
	EffSpitAcid              EffectID = 32
)

// blindPenalty is EffBlind's 4 off the attack roll, armour class and save.
const blindPenalty = 4

// rollSave is FxEngine.RollSavingThrow of c's save kind with the modifier;
// a blind c (COND_SAVING_THROW) saves at -4 on the Byte roll.
func rollSave(c *Character, kind int, modifier int8, rng *rules.RNG) bool {
	roll, made, fixed := SaveDie(rng, modifier)
	if fixed {
		return made
	}
	if hasEffect(c, EffBlind) {
		roll = int(uint8(roll - blindPenalty))
	}
	return SaveMade(c.Saves[kind], roll)
}

// defenderDamage is CalcWeaponDamage's WEAPON_DAMAGE_DEFENDER effects of d
// against a weapon of class wc (armed false: no readied weapon) at dist:
// IMMUNE_TO_ROCKET, DODGE_SLUG_HITS, IMMUNE_TO_MICROWAVE,
// IMMUNE_TO_HEAT_AND_PLASMA, IMMUNE_TO_LASER, RESIST_RANGED_HITS,
// RESIST_MELEE_HITS, UNKILLABLE, in that order. dodged reports a dodge
// (the hit no longer counts).
func defenderDamage(d *Character, dmg int, armed bool, wc xxvc.WeaponClass, dist int, rng *rules.RNG) (out int, dodged bool) {
	immune := func(id EffectID, classes ...xxvc.WeaponClass) {
		if armed && hasEffect(d, id) && slices.Contains(classes, wc) {
			dmg = 0
		}
	}
	immune(EffImmuneToRocket, xxvc.Rocket)
	if armed && hasEffect(d, EffDodgeSlugHits) && wc == xxvc.Slug && dist > 1 && PercentileRoll(rng, 50) {
		dmg, dodged = 0, true
	}
	immune(EffImmuneToMicrowave, xxvc.Microwave)
	immune(EffImmuneToHeatAndPlasma, xxvc.Heat, xxvc.Plasma)
	immune(EffImmuneToLaser, xxvc.Laser)
	if armed && hasEffect(d, EffResistRangedHits) && wc >= xxvc.Slug && wc <= xxvc.Heat && dist > 1 && PercentileRoll(rng, 50) {
		dmg = 0
	}
	if armed && hasEffect(d, EffResistMeleeHits) && wc == xxvc.Melee && PercentileRoll(rng, 50) {
		dmg = 0
	}
	if hasEffect(d, EffUnkillable) && dmg > d.HP {
		dmg = int(uint8(d.HP)) // Lo(hitPoints)
	}
	return dmg, dodged
}

// incapacitating is gdconsts.pas incapacitatingEffects.
var incapacitating = [...]EffectID{EffScared, EffHelpless, EffStunned, EffDizzy}

// Effect is types.pas TEffect: Minutes left, 0 for one that does not run out.
type Effect struct {
	ID            EffectID
	Minutes       int
	Strength      int // the save modifier its handler passes
	CallsOnRemove bool
}

// maxEffectStep is gametime.pas MAX_EFFECT_STEP: minutes off a duration at once.
const maxEffectStep = 10

// timeOutEffects is gametime.pas TimeOutEffects for c: minutes run the
// durations down maxEffectStep at a time; an effect that runs out goes
// (RemoveEffect).
// An effect that calls on remove runs its handler (removed).
// ponytail: the original's walk quirk after a removal (passed effects losing another
// step) is not kept; the camp's effectFlags skip only saves work.
func timeOutEffects(c *Character, s *combatState, minutes int) {
	for minutes > 0 {
		step := min(minutes, maxEffectStep)
		minutes -= step
		c.Effects = slices.DeleteFunc(c.Effects, func(e Effect) bool {
			out := e.Minutes != 0 && step >= e.Minutes
			if out {
				removed(e, s)
			}
			return out
		})
		for i := range c.Effects {
			if c.Effects[i].Minutes != 0 {
				c.Effects[i].Minutes -= step
			}
		}
	}
}

// hasEffect is CharHlp.HasEffect.
func hasEffect(c *Character, id EffectID) bool {
	return slices.ContainsFunc(c.Effects, func(e Effect) bool { return e.ID == id })
}

// removeEffect is FxEngine.RemoveEffect of the first effect id; s is the
// character's combat state, nil outside a fight.
func removeEffect(c *Character, s *combatState, id EffectID) {
	if i := slices.IndexFunc(c.Effects, func(e Effect) bool { return e.ID == id }); i >= 0 {
		removed(c.Effects[i], s)
		c.Effects = slices.Delete(c.Effects, i, i+1)
	}
}

// removed is RemoveEffect's CallEffectHandler(removed = True) for an effect
// that calls on remove: EffHelpless (1, 5..12) and EffStunned run
// EffEndActions; every other BR1 handler does nothing on removal. No AddEffect
// call in BR1 sets callsOnRemove, so only an imported effect has it.
func removed(e Effect, s *combatState) {
	if e.CallsOnRemove && s != nil && (e.ID == EffHelpless || e.ID >= 5 && e.ID <= 12 || e.ID == EffStunned) {
		endActions(s)
	}
}

// removeIncapacitating is FxEngine.RemoveIncapacitatingEffects: the first of
// each incapacitating effect goes.
func removeIncapacitating(c *Character, s *combatState) {
	for _, id := range incapacitating {
		removeEffect(c, s, id)
	}
}

// addEffect is FxEngine.AddEffectWithMessage with a save that cancels: the
// COND_EFFECT_IMMUNITY effects of c against the source's readied weapon
// (BLOCK_STUN_WEAPONS: sonic or a stun grenade; BLOCK_DAZZLE_GRENADE;
// BLOCK_HELPLESS; BLOCK_SCARED; UNKILLABLE blocks every effect), then a held
// effect only lengthens, else e is added. It answers whether c is affected.
func (r *Rules) addEffect(c, source *Character, e Effect, saved bool) bool {
	_, wt, wty, armed := r.readied(source, xxvc.SlotWeapon)
	blocked := hasEffect(c, EffUnkillable) ||
		armed && hasEffect(c, EffBlockStunWeapons) && (wty.WeaponClass == xxvc.Sonic || wt.Type == stunGrenadeType) ||
		armed && hasEffect(c, EffBlockDazzleGrenade) && wt.Type == dazzleGrenadeType ||
		e.ID == EffHelpless && hasEffect(c, EffBlockHelpless) ||
		e.ID == EffScared && hasEffect(c, EffBlockScared)
	if blocked || saved {
		return false
	}
	if i := slices.IndexFunc(c.Effects, func(h Effect) bool { return h.ID == e.ID }); i >= 0 {
		if h := &c.Effects[i]; h.Minutes != 0 && e.Minutes > h.Minutes {
			h.Minutes = e.Minutes
		}
		return true
	}
	c.Effects = append(c.Effects, e)
	return true
}

// poison is the attacker's COND_HIT_PHASE_2 effects on its target d:
// DIZZYING_POISON (save 1 at its strength or DIZZY) and LETHAL_POISON (save
// 1 at its strength or KillCreature poisoned).
func (r *Rules) poison(a, d *Character, did rules.CharID, ds *combatState, rng *rules.RNG) []rules.Effect {
	var out []rules.Effect
	for _, id := range [...]EffectID{EffDizzyingPoison, EffLethalPoison} { // TriggerEffect: the first of each
		i := slices.IndexFunc(a.Effects, func(e Effect) bool { return e.ID == id })
		if i < 0 {
			continue
		}
		switch e := a.Effects[i]; id {
		case EffDizzyingPoison:
			if r.addEffect(d, a, Effect{ID: EffDizzy}, rollSave(d, 1, int8(e.Strength), rng)) {
				out = append(out, rules.Narrated{Text: d.Name + " is poisoned"}, rules.Narrated{Text: d.Name + " is dizzy"})
			}
		case EffLethalPoison:
			if !rollSave(d, 1, int8(e.Strength), rng) {
				out = append(out, rules.Narrated{Text: d.Name + " is poisoned"})
				if d.Status != StDead && d.Status != StGone {
					d.Status, d.Enabled, d.HP = StPoisoned, false, 0
					removeIncapacitating(d, ds)
					if ds != nil {
						ds.initiative = 0
					}
					out = append(out, rules.Fell{Unit: unitOf(did)})
				}
			}
		}
	}
	return out
}

// monsterSpecial is the Maneuver Decide returns for C_Cond.MONSTER_SPECIAL:
// the ruleset's own, never offered to a player.
const monsterSpecial rules.ManeuverID = 1

// acidReach is EffSpitAcid's range.
const acidReach = 12

// monsterSpecial is TriggerEffectsWithCondition(MONSTER_SPECIAL): SPIT_ACID
// (an enemy target within 12: face it, 30% for 2d8, inflicted twice as the
// original's InflictDamage and ProcessAttackResult both do) then
// STRIKE_ADJACENT (an adjacent enemy target: 4, or 8 on a failed save 6).
func (r *Rules) monsterSpecial(f rules.Fight, id rules.CharID, c *Character, s *combatState, rng *rules.RNG) rules.Outcome {
	var o rules.Outcome
	hurt := func(did rules.CharID, d *Character, dmg int) {
		o.Effects = append(o.Effects, rules.Attacked{From: unitOf(id), To: unitOf(did), Hit: dmg > 0, Damage: int32(dmg)})
		if dmg > 0 && InflictDamage(d, dmg) {
			ds := r.fight[did]
			if ds != nil {
				ds.initiative = 0
			}
			removeIncapacitating(d, ds)
			o.Effects = append(o.Effects, rules.Fell{Unit: unitOf(did)})
		}
	}
	at, _ := unitAt(f, id)
	near := func(reach int) (*Character, rules.BoardPos, bool) {
		d := r.chars[s.target]
		dat, ok := unitAt(f, s.target)
		return d, dat, ok && d != nil && d.Enabled && d.team() != c.team() && distance(at, dat) <= reach
	}
	if d, dat, ok := near(acidReach); ok && hasEffect(c, EffSpitAcid) {
		facing := rules.Facing(rules.DirectionTo(at.X, at.Y, dat.X, dat.Y))
		o.Effects = append(o.Effects, rules.Turned{Unit: unitOf(id), Facing: facing}, rules.Narrated{Text: c.Name + " spits acid"})
		dmg := 0
		if PercentileRoll(rng, 30) {
			dmg = rules.RNGDice(rng, 2, 8)
			InflictDamage(d, dmg)
		}
		hurt(s.target, d, dmg)
	}
	if d, _, ok := near(1); ok && hasEffect(c, EffStrikeAdjacent) {
		dmg := 8
		if rollSave(d, 6, 0, rng) {
			dmg = 4
		}
		hurt(s.target, d, dmg)
	}
	return o
}
