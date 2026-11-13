package rules1e

import "slices"

// The temple, ported from ovr05. ECL COMBAT with the temple flag (game
// +0x5C4) set runs Temple instead of a fight. The questions (not needed,
// pay?) are the front end's: Cure comes agreed.

// Diseases = DS:0114: the affects Cure Disease removes.
var Diseases = [6]AffectType{affHelpless, affCauseDisease, affWeakened, affSick, affMummyDisease, affRotting}

// CureNeeded: c has what temple service i cures (the tests of ovr05:0249 .. 0928).
func (e *RuleState) CureNeeded(c *Character, i int) bool {
	sv, ok := e.templeService(i)
	if !ok {
		return false
	}
	if sv.Kind.Wound() && (c.Status == StatusDead || c.Status == StatusGone) {
		return false
	}
	switch sv.Kind {
	case TempleLight, TempleSerious, TempleCritical:
		return c.HP < c.MaxHP
	case TempleBlindness:
		return e.FindAffect(c, AffBlind) != nil
	case TempleDisease:
		return e.diseased(c)
	case TempleHeal:
		return c.HP < c.MaxHP || e.FindAffect(c, AffBlind) != nil || e.diseased(c) || e.FindAffect(c, affFeebleminded) != nil
	case TemplePoison:
		return e.FindAffect(c, AffPoisoned) != nil
	case TempleRaiseDead:
		return c.Status == StatusDead || c.Status == StatusAnimated
	case TempleRemoveCurse:
		for _, it := range c.Items {
			if it.Cursed != 0 {
				return true
			}
		}
		return e.FindAffect(c, AffBestowCurse) != nil
	case TempleStone:
		return c.Status == StatusStoned
	}
	return true
}

// diseased: c has one of the Diseases.
func (e *RuleState) diseased(c *Character) bool {
	for _, d := range Diseases {
		if e.FindAffect(c, d) != nil {
			return true
		}
	}
	return false
}

// TempleCure is the work of temple service i on c, paid for (the repair on the whole party).
func (e *RuleState) TempleCure(c *Character, i int) {
	sv, ok := e.templeService(i)
	if !ok {
		return
	}
	if sv.Kind == TempleRepair {
		for _, m := range e.Party {
			e.repair(m)
		}
		return
	}
	e.cure(c, sv.Kind)
}

// repair is every cure the temple has on c, in the order of the services.
func (e *RuleState) repair(c *Character) {
	for _, k := range []TempleKind{TempleBlindness, TempleDisease, TemplePoison, TempleStone, TempleRaiseDead, TempleHeal} {
		e.cure(c, k)
	}
}

// cure is the work of a temple service of kind k on c.
func (e *RuleState) cure(c *Character, k TempleKind) {
	switch k {
	case TempleBlindness: // ovr05:0249
		e.RemoveAffect(c, AffBlind, nil)
	case TempleDisease: // ovr05:02ED
		e.cureDiseases(c)
		if e.restoresStats(k) {
			e.statAffects(c, StatStr)
		}
	case TempleLight: // ovr05:0401: light, serious, critical
		e.Heal(c, e.Dice(1, 8), false)
	case TempleSerious:
		e.Heal(c, e.Dice(2, 8)+1, false)
	case TempleCritical:
		e.Heal(c, e.Dice(3, 8)+3, false)
	case TempleHeal: // all but a few points of the wounds, the blindness, diseases and feeblemind (Felix 2026-11-12)
		if lost := int(c.MaxHP) - int(c.HP) - e.Dice(1, 4); lost > 0 {
			e.Heal(c, lost, false)
		}
		e.cure(c, TempleBlindness)
		e.cureDiseases(c)                                    // never RestoresStats: pod's Heal (O04C:0199) reruns no strength
		if a := e.FindAffect(c, affFeebleminded); a != nil { // pod O04C:0199: Int and Wis again
			e.RemoveAffect(c, affFeebleminded, a)
			e.statAffects(c, StatInt)
			e.statAffects(c, StatWis)
		}
	case TemplePoison: // ovr05:0751
		e.Rv.Cure = 1
		e.RemoveAffect(c, AffPoisoned, nil)
		e.RemoveAffect(c, affSlowPoison, nil)
		e.RemoveAffect(c, affPoisonDamage, nil)
		e.Rv.Cure = 0
	case TempleRaiseDead:
		e.raiseDead(c)
	case TempleRemoveCurse: // ovr05:083A: the Remove Curse spell (ovr23:2540)
		e.spellTarget = c
		e.spellRemoveCurse()
	case TempleStone: // ovr05:0928
		if c.Status == StatusStoned {
			c.Status = StatusOkay
			c.Okay = 1
			c.HP = 1
		}
	}
}

// raiseDead = ovr05:0510: back with 1 hp and a point of constitution less; while the constitution is still 14 or
// more, the hit points above the dice divided by the constitution bonus summed over the class levels come off the
// maximum when the constitution is below 17 or c has fighter levels (por-re src/pure Temple).
func (e *RuleState) raiseDead(c *Character) {
	if c.Status != StatusDead && c.Status != StatusAnimated {
		return
	}
	e.Rv.Cure = 1
	e.RemoveAffect(c, affAnimated, nil)
	e.RemoveAffect(c, AffPoisoned, nil)
	e.Rv.Cure = 0
	c.HP = 1
	c.Status = StatusOkay
	c.Okay = 1
	if c.Stats[StatCon] > 0 {
		c.Stats[StatCon]--
	}
	con := c.Stats[StatCon]
	if con < 14 {
		return
	}
	var loss, bonus byte
	if c.MaxHP > c.HPRoll {
		loss = c.MaxHP - c.HPRoll
	}
	for i, l := range c.Levels {
		if int8(l) <= 0 {
			continue
		}
		switch {
		case i == ClsFighter:
			bonus += byte(con-14) * byte(l)
		case con > 15:
			bonus += byte(l) * 2
		default:
			bonus += byte(l)
		}
	}
	if bonus > 0 {
		loss /= bonus
	}
	if con < 17 || int8(c.Levels[ClsFighter]) > 0 {
		c.MaxHP -= loss
	}
}

// restoresStats: the temple's service of kind k works the scores out again after it cures (TempleService.RestoresStats).
func (e *RuleState) restoresStats(k TempleKind) bool {
	return slices.ContainsFunc(e.T().Town.Temple, func(sv TempleService) bool { return sv.Kind == k && sv.RestoresStats })
}

// cureDiseases removes the Diseases from c (pod: nausea 31, disease 34, weakness 43, wasting 44) with the cure flag set.
func (e *RuleState) cureDiseases(c *Character) {
	e.Rv.Cure = 1
	for _, d := range Diseases {
		e.RemoveAffect(c, d, nil)
	}
	e.Rv.Cure = 0
}
