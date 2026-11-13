package rules1e

// spellEffectPodFirst runs the effect of the unified spells 128..145, the ids
// pod has first (spellids_games.go), ported from its spellFx (fold T3f).
// false: the spell is not one of them (or not ported yet).
func (e *RuleState) spellEffectPodFirst(spell SpellID) bool {
	switch spell {
	case spDestruction:
		e.fxDestructionPod()
	case spResurrection:
		e.fxResurrectionPod()
	case spNeutralizePoisonDruid:
		e.fxNeutralizePoisonPod()
	case spFireTouch:
		e.applySpell(int(spell), 0, 0, 0, 0, "is covered in flame") // O0BB:59AD
	case spIronSkin, spMindBlank:
		e.applySpell(int(spell), 0, 0, 0, 0, "is protected") // fxMinorGlobe, fxProtected
	case spMassCharm:
		e.fxMassCharmPod()
	case spOttosIrresistibleDance:
		e.fxOttosDancePod()
	case spPowerWordBlind:
		e.fxPowerWordBlindPod()
	case spPowerWordKill:
		e.fxPowerWordKillPod()
	default:
		return false
	}
	return true
}

// PoD's spell numbers beyond POR's: the level cap, the spell record's
// fields, the marks and the bits its handlers read.
const (
	maxSpellLevelPod     = 31   // cappedCasterLevel's cap
	spellFieldRangePod   = 2    // spellRangeTouchPod: a touch spell, an attack roll first
	spellFieldSaveMode   = 8    // saveNegate, saveHalf
	spellFieldSaveType   = 9    // the save
	spellFieldAffectPod  = 0xA  // the affect the spell lays, as PoD's number
	spellRangeTouchPod   = 0xFF // spellFieldRangePod of a touch spell
	destroyMarkPod       = 255  // Rv.Damage: kill outright, unless a handler clears it
	dmgDestroyPod        = dmgType64Cotab | DmgMagic
	sizeManPod           = 1 // the largest Size Mass Charm takes
	mindBlankRoundsPod   = 1440
	ottoDanceDiceSides   = 4 // Otto's: 1d4+1 rounds
	ottoDanceBonusRounds = 1
	powerWordBlindHPPod  = 100 // the hit points Power Word Blind blinds in all
	powerWordBlindSides  = 4   // PoD's kept 1d4 duration roll
	pwKillOneMaxHPPod    = 60
	pwKillManyMaxHPPod   = 10
	massCharmSaveBonus   = -2
	sideBitPod           = 7 // an affect's data: the caster's side in the high bit
)

// PoD's MonsterFlags bits (byte 350).
const (
	monsterPersonLikePod MonsterFlags = 2 // Charm Person works on it
	monsterNoBreathPod   MonsterFlags = 8 // not alive to destroy; immune to dmgType64
)

// fxDestructionPod = O0BB:5142: no save; only what lives can be destroyed.
func (e *RuleState) fxDestructionPod() {
	t := e.tgt(1)
	e.Rv.DmgType = dmgDestroyPod
	e.Rv.Damage = destroyMarkPod
	e.AffectGroup(t, grpAddAffect)
	if t.MonsterFlags&monsterNoBreathPod == 0 && e.Rv.Damage != 0 {
		e.killWithMessage(t, StatusGone, "is destroyed")
	} else {
		e.CombatSay(t, "is unaffected", true)
	}
}

// fxResurrectionPod = O0BB:51D1: like Raise Dead, but whole and for any
// race but none, and no constitution lost.
func (e *RuleState) fxResurrectionPod() {
	t := e.tgt(1)
	if t.Status == StatusDead && t.Stats[StatCon] > 0 && t.raceNumber() != 0 {
		e.Rv.Cure = 1
		e.RemoveAffect(t, AffPoisoned, nil)
		e.Rv.Cure = 0
		t.Status = StatusOkay
		t.Okay = 1
		t.HP = t.MaxHP
		e.CombatSay(t, "is ressurected", true) // sic
	}
}

// fxNeutralizePoisonPod = O0BB:3E7D: an animated target drops out; a
// poisoned one is back on its feet with at least a hit point, cured of
// its poison affects.
func (e *RuleState) fxNeutralizePoisonPod() {
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

// fxMassCharmPod = O0BB:5A2D: twice the caster level in hit dice of
// persons not of the caster's side and at most man-sized are charmed, in
// list order, while they fit; the charm handler runs at once.
func (e *RuleState) fxMassCharmPod() {
	e.areaSpell = 1
	sel := e.Acting()
	budget := e.cappedCasterLevel(int(spMassCharm)) * 2
	used := 0
	for i := 1; i <= e.NSpellTargets && used != budget; i++ {
		t := e.tgt(i)
		if t == nil {
			continue
		}
		if t.MonsterFlags&monsterPersonLikePod == 0 || t.Size > sizeManPod || t.Side == sel.Side {
			e.CombatSay(t, "is unaffected", true)
			continue
		}
		hd := int(t.MaxLevel)
		if used+hd > budget {
			continue
		}
		used += hd
		data := byte(int(sel.Side)<<sideBitPod + int(byte(e.casterLevel(int(spMassCharm)))))
		e.AddAffectMsg(t, AffCharmed, e.spellAmount(int(spMassCharm)), data, 1, saveNegate,
			e.SavingThrow(t, SaveVsSpell, massCharmSaveBonus), "is charmed")
		if a := e.FindAffect(t, AffCharmed); a != nil {
			e.callAffect(AffCharmed, t, a, 0)
		}
	}
}

// fxOttosDancePod = O0BB:5BC7: a touch attack against the target's armour.
func (e *RuleState) fxOttosDancePod() {
	t := e.tgt(1)
	e.ApplyItems(t)
	e.AffectGroup(t, grpArmour)
	if e.AttackHits(e.Acting(), t, int(t.AC[0])) {
		e.AddAffectMsg(t, affDancingPod, e.spellAmount(int(spOttosIrresistibleDance)),
			byte(e.cappedCasterLevel(int(spOttosIrresistibleDance))), 0, 0, false, "begins to dance wildly")
	}
}

// fxPowerWordBlindPod = O0BB:5C8B: targets are blinded in list order while
// their maximum hit points fit in 100. PoD rolls a 1d4 duration it then
// ignores; the roll is kept.
func (e *RuleState) fxPowerWordBlindPod() {
	e.areaSpell = 1
	_ = e.Dice(1, powerWordBlindSides)
	sum := 0
	for i := 1; i <= e.NSpellTargets && sum <= powerWordBlindHPPod; i++ {
		t := e.tgt(i)
		if hp := int(t.MaxHP); hp+sum <= powerWordBlindHPPod {
			sum += hp
			e.AddAffectMsg(t, AffBlind, e.spellAmount(int(e.CurSpell)), 0, 0, 0, false, "is blinded")
		}
	}
}

// killUnlessWardedPod: the add-affect group may clear the destroy mark;
// else the target dies with status and msg.
func (e *RuleState) killUnlessWardedPod(t *Character, status byte, msg string) {
	e.Rv.Damage = destroyMarkPod
	e.AffectGroup(t, grpAddAffect)
	if e.Rv.Damage != 0 {
		e.killWithMessage(t, status, msg)
	}
}

// fxPowerWordKillPod = O0BB:617E: the first target dies if it has 11..60
// hit points; otherwise every target with 10 or fewer does.
func (e *RuleState) fxPowerWordKillPod() {
	e.areaSpell = 1
	if e.NSpellTargets <= 0 {
		return
	}
	if t := e.tgt(1); t != nil && t.HP <= pwKillOneMaxHPPod && t.HP > pwKillManyMaxHPPod {
		e.killUnlessWardedPod(t, StatusDead, "is slain")
		return
	}
	for i := 1; i <= e.NSpellTargets; i++ {
		if t := e.tgt(i); t != nil && t.HP <= pwKillManyMaxHPPod {
			e.killUnlessWardedPod(t, StatusDead, "is slain")
		}
	}
}
