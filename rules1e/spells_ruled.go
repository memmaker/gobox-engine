package rules1e

// Felix's spell rulings (fold T3h, FOLD-PLAN "grill Q1" and "spell-rulings
// page"): one effect for every game, ported from the game named on each.

// spellEffectRuled runs the ruled spells; false: not one of them.
func (e *RuleState) spellEffectRuled(spell SpellID) bool {
	switch spell {
	case spCureSeriousWounds:
		e.healTarget(e.Dice(2, 8) + 1)
	case spCureCriticalWounds:
		e.healTarget(e.Dice(3, 8) + 3)
	case spPotionOfExtraHealing:
		e.fxHeal2d4()
	case spNeutralizePoisonCleric:
		e.fxNeutralizePoison()
	case spProtectionEvil10Radius, spMinorGlobeOfInvulnerability:
		e.applySpell(int(e.CurSpell), 0, 0, 0, 0, "is protected")
	case spDetectMagicDruid:
		e.applySpell(int(e.CurSpell), 0, 0, 0, 0, "is affected")
	case spInvisibilityToAnimals:
		e.applySpell(int(e.CurSpell), 0, 0, 0, 0, "is invisible")
	case spSticksToSnakes:
		e.fxSticksToSnakes()
	case spDispelEvil:
		e.AddAffectMsg(e.Acting(), affDispellingEvil, e.spellAmount(int(spDispelEvil)), 0, 0, 0, false, "")
		e.applySpell(int(e.CurSpell), 0, 0, 0, 0, "is affected")
	case spFlameStrike:
		e.areaSpell = 1
		e.applySpell(int(e.CurSpell), 0, 0, e.RollDice(6, 8), DmgMagic|DmgFire, "")
	case spIceStorm:
		e.areaSpell = 1
		e.applySpell(int(e.CurSpell), 0, 0, e.RollDice(3, 10), dmgCold|DmgMagic, "")
	case spRaiseDead:
		e.fxRaiseDead()
	case spSlayLiving:
		e.fxSlayLiving()
	case spFaerieFire:
		e.holdTargets(0, "is highlighted")
	case spCharmMonsters:
		e.fxCharmMonsters()
	case spHoldMonsters:
		e.SpellHold()
	case spConfusion:
		e.fxConfusion()
	case spDimensionDoor:
		e.fxDimensionDoor()
	case spFear:
		e.fxFear()
	case spFireShield:
		e.fxFireShield()
	case spFumble:
		e.fxFumble()
	case spRemoveCurseMage:
		e.spellRemoveCurse() // Felix's ruling: lifts the curse as the cleric's
	case spConeOfCold:
		e.fxConeOfCold()
	case spFeeblemind:
		e.fxFeeblemind()
	case spHeal:
		e.fxHeal()
	case spCharmPersonOrMammal:
		e.fxCharmPersonOrMammal()
	case spStoneToFlesh:
		e.fxStoneToFlesh()
	default:
		return false
	}
	return true
}

// healTarget = cotab fxCureSerious/fxCureCritical: Heal (it wakes the
// unconscious and refuses the mummy-diseased) on the first target.
func (e *RuleState) healTarget(amount int) {
	if e.NSpellTargets == 0 {
		return
	}
	if t := e.tgt(1); e.Heal(t, amount, false) {
		e.healedMessage(t)
	}
}

// fxNeutralizePoison = cotab O0C2:41DA: an animated target drops out; a
// poisoned one is cured of its poisons and back on its feet.
func (e *RuleState) fxNeutralizePoison() {
	t := e.tgt(1)
	if t.Status == StatusAnimated {
		e.SetTgt(1, nil)
		return
	}
	if e.FindAffect(t, AffPoisoned) == nil {
		e.CombatSay(t, "is unaffected", true)
		return
	}
	if t.HP == 0 {
		t.HP = 1
	}
	e.Rv.Cure = 1
	e.RemoveAffect(t, AffPoisoned, nil)
	e.RemoveAffect(t, affSlowPoison, nil)
	e.RemoveAffect(t, affPoisonDamage, nil)
	e.Rv.Cure = 0
	e.CombatSay(t, "is unpoisoned", true)
	t.Okay = 1
	t.Status = StatusOkay
}

// sticksToSnakesTooBig: the hit dice from which a target is too big for the snakes.
const sticksToSnakesTooBig = 6

// fxSticksToSnakes = pod O0BB:3FFA with sotsb's caster message: a small
// target gets the snakes (their handler runs at once); a big one smashes
// them flat unless it is helpless.
func (e *RuleState) fxSticksToSnakes() {
	t := e.tgt(1)
	if t.MaxLevel < sticksToSnakesTooBig {
		e.CombatSay(e.Acting(), "turns sticks into snakes", true)
		e.applySpell(int(e.CurSpell), e.cappedCasterLevel(int(e.CurSpell)), 0, 0, DmgMagic, "")
		typ := e.spellAffect(e.CurSpell)
		if a := e.FindAffect(t, typ); a != nil {
			e.callAffect(typ, t, a, 0)
		}
		return
	}
	if !e.IsHelpless(t) {
		e.CombatSay(t, "smashes them flat", true)
	}
}

// fxRaiseDead = cotab O0C2:451A: a dead (or animated) target with
// constitution left and not an elf comes back with a hit point, a point of
// constitution less.
func (e *RuleState) fxRaiseDead() {
	t := e.tgt(1)
	if (t.Status != StatusDead && t.Status != StatusAnimated) || t.Stats[StatCon] == 0 || t.Race == RaceElf {
		return
	}
	e.Rv.Cure = 1
	e.RemoveAffect(t, affAnimated, nil) // the collapse of the animated
	e.RemoveAffect(t, AffPoisoned, nil)
	e.Rv.Cure = 0
	t.Status = StatusOkay
	t.Okay = 1
	t.Stats[StatCon]--
	t.HP = 1
	e.CombatSay(t, "is raised", true)
}

// fxSlayLiving = sotsb fxSlayLiving: the living only (the breathless are
// unaffected); the target's affects may stop it; else it dies unless it
// saves, then taking 2d8+1.
func (e *RuleState) fxSlayLiving() {
	t := e.tgt(1)
	e.Rv.DmgType = dmgSlaying
	e.Rv.Damage = slayMark
	e.AffectGroup(t, grpAddAffect)
	if t.MonsterFlags&monsterNoBreath != 0 {
		e.Rv.Damage = 0
	}
	if e.Rv.Damage == 0 {
		e.CombatSay(t, "is unaffected", true)
		return
	}
	if !e.SavingThrow(t, SaveVsSpell, 0) {
		e.killWithMessage(t, StatusDead, "is slain")
		return
	}
	e.Rv.DmgType = DmgMagic
	e.Damage(t, e.RollDice(2, 8)+1, 0, false)
}

// fxCharmMonsters = cotab O0C2:47DC: held like Hold Monsters; the charm
// handler runs at once for each one charmed.
func (e *RuleState) fxCharmMonsters() {
	e.holdTargets(0, "is charmed")
	for i := 1; i <= e.NSpellTargets; i++ {
		if t := e.tgt(i); t != nil {
			if a := e.FindAffect(t, AffCharmed); a != nil {
				e.callAffect(AffCharmed, t, a, 0)
			}
		}
	}
}

// fxConfusion = cotab O0C2:486D with pod's area flag: up to 2d8 targets
// are confused unless they save.
func (e *RuleState) fxConfusion() {
	e.areaSpell = 1
	if n := e.Dice(2, 8); e.NSpellTargets > n {
		e.NSpellTargets = n
	}
	for i := 1; i <= e.NSpellTargets; i++ {
		t := e.tgt(i)
		if t == nil {
			continue
		}
		saved := e.SavingThrow(t, SaveVsSpell, 0)
		e.AddAffectMsg(t, e.spellAffect(spConfusion), e.spellAmount(int(spConfusion)), 0, 0, saveNegate, saved, "is confused")
	}
}

// fxDimensionDoor = cotab O0C2:4931: an engulfed or hugged caster is let
// go by whoever holds it; it steps onto the target square.
func (e *RuleState) fxDimensionDoor() {
	sel := e.Acting()
	if e.FindAffect(sel, affHeldFast) != nil {
		e.BuildNear(e.CombatX(sel), e.CombatY(sel), 1, 0xFF, 1)
		for _, n := range e.Near {
			c := e.Combatant(n.Index)
			a := e.FindAffect(c, affHuggingCotab)
			if a == nil {
				a = e.FindAffect(c, affEngulfingCotab)
			}
			if a != nil && e.Combatant(a.Data) == sel {
				e.RemoveAffect(c, affHuggingCotab, nil)
				e.RemoveAffect(c, affEngulfingCotab, nil)
			}
		}
	}
	e.combatSquareFree(sel, e.TgtX, e.TgtY, false)
	e.CombatSay(sel, "teleports", true)
}

// The cone of Fear: 3 wide, 6 long.
const (
	fearConeWidth  = 3
	fearConeLength = 6
)

// fxFear = pod O0BB:4693: everyone in the cone that does not save runs,
// out of the player's hands (unless something blocked the fear).
func (e *RuleState) fxFear() {
	e.areaSpell = 1
	sel := e.Acting()
	e.coneTargets(e.CombatX(sel), e.CombatY(sel), e.TgtX, e.TgtY, fearConeWidth, fearConeLength)
	for i := 1; i <= e.NSpellTargets; i++ {
		t := e.tgt(i)
		if e.SavingThrow(t, SaveVsSpell, 0) {
			e.CombatSay(t, "is unaffected", true)
			continue
		}
		e.AddAffectMsg(t, affFrightened, e.spellAmount(int(spFear)), 0, 1, saveNegate, false, "runs in terror")
		if e.FindAffect(t, affFrightened) == nil {
			continue
		}
		combatInfo(t).Turned = 1
		t.Auto = 1
		if t.Control <= controlHired {
			t.Control = controlCharmed
		}
		combatInfo(t).Target = nil
	}
}

// Fire Shield's flames, the caster's answer to the spell's question
// (Fight.Choice, an index into fireShieldChoices).
const (
	fireShieldHot  int32 = 0
	fireShieldCold int32 = 1
)

// fireShieldChoices are the spell's question's replies.
var fireShieldChoices = []string{"Hot", "Cold"}

// fxFireShield (ruling): the player picks hot or cold before the cast, a
// computer caster 50/50; the shield burns attackers.
func (e *RuleState) fxFireShield() {
	c := e.Acting()
	flame := e.Fight.Choice
	if c.Auto != 0 {
		flame = int32(e.Dice(1, 2) - 1)
	}
	body := affFireShieldHot
	if flame == fireShieldCold {
		body = affFireShieldCold
	}
	rounds := e.spellAmount(int(spFireShield))
	e.AddAffectMsg(c, body, rounds, 0, 0, 0, false, "is protected")
	e.AddAffectMsg(c, affZapsAttackers, rounds, 0, 0, 0, false, "")
}

// fxFumble = cotab O0C2:4DF6: an unsaved target is clumsy, a saved one
// slowed; the handler runs at once. Then the spell's own affect.
func (e *RuleState) fxFumble() {
	t := e.tgt(1)
	e.Rv.DmgType = dmgSlaying
	clumsy := e.spellAffect(spFumble)
	if !e.SavingThrow(t, SaveVsSpell, 0) {
		e.AddAffectMsg(t, clumsy, e.spellAmount(int(spFumble)), 0, 0, 0, false, "is clumsy")
		if e.FindAffect(t, clumsy) != nil {
			e.callAffect(clumsy, t, nil, 0)
		}
	} else {
		e.AddAffectMsg(t, AffSlow, e.spellAmount(int(spFumble)), 0, 0, 0, false, "is slowed")
		if e.FindAffect(t, AffSlow) != nil {
			e.callAffect(AffSlow, t, nil, 0)
		}
	}
	e.applySpell(int(e.CurSpell), 0, 1, 0, 0, "is clumsy")
}

// The cone of Cone of Cold: 2 wide, half the caster level long.
const coneOfColdWidth = 2

// fxConeOfCold = cotab O0C2:5550 with pod's area flag: level + level d4 of
// cold and magic.
func (e *RuleState) fxConeOfCold() {
	e.areaSpell = 1
	sel := e.Acting()
	lvl := byte(e.casterLevel(int(e.CurSpell)))
	length := max((int(lvl)+1)>>1, 1)
	e.coneTargets(e.CombatX(sel), e.CombatY(sel), e.TgtX, e.TgtY, coneOfColdWidth, length)
	e.applySpell(int(e.CurSpell), 0, 0, int(lvl)+e.RollDice(int(lvl), 4), dmgCold|DmgMagic, "")
}

// Feeblemind's save changes by class (sotsb fxFeeblemind).
const (
	feebleRangerLevel  = 8 // a ranger of this level saves 1 worse, above 2
	feeblePaladinLevel = 9 // a paladin from this level saves 4 worse
)

// fxFeeblemind = sotsb fxFeeblemind's class table with pod's memorised
// wipe: the target's spell save is changed for the spell (fighters and
// thieves are unaffected); the handler runs at once and the spells still
// being memorised are lost.
func (e *RuleState) fxFeeblemind() {
	t := e.tgt(1)
	keep := t.Saves[SaveVsSpell]
	applies := true
	switch t.Class {
	case ClsCleric, clsClericFighter, clsClericThief:
		t.Saves[SaveVsSpell]--
	case ClsRanger:
		if t.MaxLevel == feebleRangerLevel {
			t.Saves[SaveVsSpell]++
		}
		if int8(t.MaxLevel) > feebleRangerLevel {
			t.Saves[SaveVsSpell] += 2
		}
	case ClsPaladin:
		if int8(t.MaxLevel) >= feeblePaladinLevel {
			t.Saves[SaveVsSpell] += 4
		}
	case ClsMage, clsFighterMage, clsFighterMageThief, clsMageThief:
		if t.Race == RaceHuman {
			t.Saves[SaveVsSpell] += 4
		} else {
			t.Saves[SaveVsSpell] += 2
		}
	case clsClericFighterMage, ClsClericRanger, clsClericMage:
		t.Saves[SaveVsSpell] += 2
	default:
		applies = false
	}
	if applies {
		e.applySpell(int(e.CurSpell), 0, 0, 0, 0, "")
	}
	typ := e.spellAffect(spFeeblemind)
	if e.FindAffect(t, typ) != nil {
		e.callAffect(typ, t, nil, 0)
		for i, sp := range t.Memorized {
			if sp&SpellPending != 0 {
				t.Memorized[i] = 0
			}
		}
	} else {
		e.CombatSay(t, "is unaffected", true)
	}
	t.Saves[SaveVsSpell] = keep
}

// fxHeal = sotsb fxHeal: cures diseases, blindness and feeblemind, then
// brings the target to its maximum hit points less 1d4.
func (e *RuleState) fxHeal() {
	t := e.tgt(1)
	if e.cureDiseaseAffects() {
		e.CombatSay(t, "is cured", false)
	}
	if e.CureAffect(t, AffBlind) {
		e.CombatSay(t, "is no longer blind", false)
	}
	if e.CureAffect(t, affFeebleminded) {
		e.CombatSay(t, "regains intellect", false) // no score model: the scores stay as Feeblemind set them (unported.go statAffects)
	}
	short := byte(e.Dice(1, 4))
	var amount byte
	if int(t.HP)+int(short) <= int(t.MaxHP) {
		amount = t.MaxHP - t.HP - short
	}
	if e.Heal(t, int(amount), false) {
		e.CombatSay(t, "is healed", false)
	}
}

// fxCharmPersonOrMammal = pod O0BB:4CDE: everyone but monsters that are
// not charmable, at the capped caster level.
func (e *RuleState) fxCharmPersonOrMammal() {
	t := e.tgt(1)
	if t == nil {
		return
	}
	if t.Race != raceMonster || t.CreatureFlags&creatureCharmable != 0 {
		lvl := e.cappedCasterLevel(int(e.CurSpell))
		e.applySpell(int(e.CurSpell), (int(e.Acting().Side)<<7+lvl)&0xFF, 1, 0, 0, "is charmed")
		if a := e.FindAffect(t, AffCharmed); a != nil {
			e.callAffect(AffCharmed, t, a, 0)
		}
		return
	}
	e.CombatSay(t, "is not charmed", true)
}

// creatureCharmable: Charm Monsters and Charm Person or Mammal work on it (sotsb, pod).
const creatureCharmable CreatureFlags = 2

// fxStoneToFlesh = sotsb O113:5AD5: a statue comes back with a hit point;
// in combat its body leaves the map and it stands on its square. Out of
// combat on one not stoned (ruling): no question, the spell is kept.
func (e *RuleState) fxStoneToFlesh() {
	t := e.tgt(1)
	if t.Status != StatusStoned {
		if !e.inCombat() {
			e.CombatSay(t, "is not stoned", true)
			e.spellKept = true
			return
		}
	} else {
		t.Status = StatusOkay
		t.Okay = 1
		t.HP = 1
		e.CombatSay(t, "is no longer a statue", false)
		e.combatSquareFree(t, e.CombatX(t), e.CombatY(t), true)
	}
	e.applySpell(int(e.CurSpell), 0, 0, 0, 0, "")
}
