package rules1e

import "gobox/rules"

// Spell effect routines of ovr23 (0x100F..0x2540). CastSpell (ovr23:0C20)
// calls them through the table DS:6B84 filled by ovr23:3413, indexed by the
// spell number DS:681B; each is called without arguments and works on the
// spell globals: spellID (DS:681B), spellN/spellList (DS:6C94/6C91, list
// entry 1 is DS:6C95, the first target).

// spellBless = ovr23:100F: Bless, "is Blessed" on the selected character's side.
func (e *RuleState) spellBless() {
	e.applySpellToSide(int(e.Acting().Side), "is Blessed")
}

// spellBane = ovr23:1040: Bane, "is Cursed".
func (e *RuleState) spellBane() {
	side := 0
	if IsParty(e.Acting()) {
		side = 1
	}
	e.applySpellToSide(side, "is Cursed")
}

// spellCureLightWounds = ovr23:106B: Cure Light Wounds on the first target, 1d8.
func (e *RuleState) spellCureLightWounds() {
	if e.NSpellTargets != 0 {
		if e.Heal(e.tgt(1), e.Dice(1, 8), false) {
			e.healedMessage(e.tgt(1))
		}
	}
}

// spellDamage1d8 = ovr23:10A7: 1d8 magic damage.
func (e *RuleState) spellDamage1d8() {
	e.applySpell(int(e.CurSpell), 0, 0, e.RollDice(1, 8), 8, "")
}

// spellDamageLevel = ovr23:1190: damage equal to the caster level, type 9.
func (e *RuleState) spellDamageLevel() {
	e.applySpell(int(e.CurSpell), 0, 0, e.casterLevel(int(e.CurSpell)), 9, "")
}

// spellCharmPerson = ovr23:11DD: Charm Person; large or strong targets are unaffected.
func (e *RuleState) spellCharmPerson() {
	t := e.tgt(1)
	if t.Kind > 1 || t.Size > 1 {
		e.CombatSay(t, "is unaffected", true)
		return
	}
	lvl := e.casterLevel(int(e.CurSpell))
	e.applySpell(int(e.CurSpell), (int(e.Acting().Side)<<7+lvl)&0xFF, 1, 0, 0, "is charmed")
	if a := e.FindAffect(t, AffCharmed); a != nil {
		e.callAffect(AffCharmed, e.tgt(1), a, 0)
	}
}

// spellStrengthPercent = ovr23:12A5: Strength; the percentile comes from the caster
// level (levels above 6 leave DS:4845 as it was).
func (e *RuleState) spellStrengthPercent() {
	e.spellValue = 0x12
	switch e.casterLevel(int(e.CurSpell)) {
	case 1:
		e.spellValue2 = 0
	case 2:
		e.spellValue2 = 1
	case 3:
		e.spellValue2 = 0x33
	case 4:
		e.spellValue2 = 0x4C
	case 5:
		e.spellValue2 = 0x5B
	case 6:
		e.spellValue2 = 0x64
	}
	t := e.tgt(1)
	out, ok := e.SetStrength(t, Score(e.spellValue), e.spellValue2)
	if ok {
		e.CombatSay(t, "is stronger", true)
	}
	e.AddAffect(t, affEnlarged, e.spellAmount(int(e.CurSpell)), out, 1)
}

// spellReduce = ovr23:136E: Reduce (undoes the strength affect 0x0C) unless the target saves.
func (e *RuleState) spellReduce() {
	if e.NSpellTargets > 0 {
		t := e.tgt(1)
		if !e.SavingThrow(t, SaveVsSpell, 0) && e.CureAffect(t, affEnlarged) {
			e.CombatSay(t, "has been reduced", true)
		}
	}
}

// spellFriends = ovr23:13D8: Friends, charisma +2d4 (at most 25).
func (e *RuleState) spellFriends() {
	t := e.tgt(1)
	old := t.Stats[StatCha]
	t.Stats[StatCha] += Score(e.Dice(2, 4))
	if t.Stats[StatCha] > 0x19 {
		t.Stats[StatCha] = 0x19
	}
	e.applySpell(int(e.CurSpell), int(old), 1, 0, 0, "is friendly")
}

// spellMagicMissile = ovr23:1439: magic damage of (level/2)d4 + level/2.
func (e *RuleState) spellMagicMissile() {
	lvl := e.casterLevel(int(e.CurSpell))
	e.applySpell(int(e.CurSpell), 0, 0, (e.RollDice(lvl>>1, 4)+lvl>>1)&0xFF, 8, "")
}

// spellBurningHands = ovr23:14C7: caster level + 1d8 damage, type 0xC.
func (e *RuleState) spellBurningHands() {
	lvl := e.casterLevel(int(e.CurSpell))
	e.applySpell(int(e.CurSpell), 0, 0, (lvl+e.RollDice(1, 8))&0xFF, 0xC, "")
}

// spellSleep = ovr23:151B: Sleep. 4d4 hit dice of targets fall asleep, weakest
// first in list order; the others (and the already asleep) leave the list.
func (e *RuleState) spellSleep() {
	e.areaSpell = 1
	e.spellValue = byte(e.Dice(4, 4))
	for i := 1; i <= e.NSpellTargets; i++ {
		t := e.tgt(i)
		switch lvl := int8(t.MaxLevel); {
		case lvl >= 0 && lvl <= 1:
			e.spellValue2 = 1
		case lvl == 2:
			e.spellValue2 = 2
		case lvl == 3:
			e.spellValue2 = 4
		case lvl == 4:
			e.spellValue2 = 6
		case lvl == 5:
			if t.Race == raceMonster {
				e.spellValue2 = 0xA
			} else {
				e.spellValue2 = 0x14
			}
		default:
			e.spellValue2 = 0x14
		}
		if t.Status != StatusAnimated && e.FindAffect(t, affAsleep) == nil && e.spellValue >= e.spellValue2 {
			e.spellValue -= e.spellValue2
		} else {
			e.SetTgt(i, nil)
		}
	}
	e.applySpell(int(e.CurSpell), 0, 0, 0, 0, "falls asleep")
}

// SpellHold = ovr23:1658: Hold Person/Monster. The save bonus depends on the
// number of targets; each target gets a missile and the affect from the spell record.
func (e *RuleState) SpellHold() {
	var bonus int8 // other counts leave the local unset (0)
	switch e.NSpellTargets {
	case 1:
		if e.CurSpell == SpHoldPersonCl {
			bonus = -2
		} else {
			bonus = -3
		}
	case 2:
		bonus = -1
	}
	e.holdTargets(bonus, "is held")
}

// holdTargets = ovr23:1658's loop, cotab O0C2:1B24 (bonus, msg): each
// target, last first, gets a missile (but the first one shot) and the spell
// record's affect unless it saves. Large or strong targets save, but against
// Hold Monsters; Faerie Fire and Charm Monsters never miss their first.
func (e *RuleState) holdTargets(bonus int8, msg string) {
	spell := e.CurSpell
	for i := e.NSpellTargets; i >= 1; i-- {
		t := e.tgt(i)
		if t == nil {
			continue
		}
		if i < e.NSpellTargets {
			e.Sound(string(rules.CueSpell))
			sel := e.Acting()
			e.shoot(e.CombatX(sel), e.CombatY(sel), e.CombatX(t), e.CombatY(t), 0, SpellIDs[spell])
		}
		var saved bool
		saveMode := e.SpellRec(spell, spellFieldSaveMode)
		if (spell == spFaerieFire || spell == spCharmMonsters) && i == e.NSpellTargets {
			saved, saveMode = true, saveNegate
		} else {
			saved = e.SavingThrow(t, SaveVsSpell, int(bonus)) // vs spell, the record's byte 9, as the original (ruling 16)
		}
		if (t.Kind > 1 || t.Size > 1) && spell != spHoldMonsters {
			saved = true
		}
		e.AddAffectMsg(t, e.spellAffect(spell), e.spellAmount(int(spell)), byte(e.casterLevel(int(spell))), 0,
			saveMode, saved, msg)
	}
}

// spellRaiseDead = ovr23:185B: Raise Dead style: a dead target is dropped from
// the list; one with affect 0x37 is brought to 1 hit point and stunned (0x0F).
func (e *RuleState) spellRaiseDead() {
	t := e.tgt(1)
	if t.Status == StatusAnimated {
		e.SetTgt(1, nil)
		return
	}
	if e.FindAffect(t, AffPoisoned) == nil {
		return
	}
	if t.HP == 0 {
		t.HP = 1
	}
	e.applySpell(int(e.CurSpell), 0xFF, 1, 0, 0, "is affected")
	e.callAffect(affGetUp, t, nil, 1)
	e.AddAffect(t, affPoisonDamage, 10, 0xFF, 1)
}

// spellUndeadTargets = ovr23:190F: the party members of kind 0xE whose hit points
// fit in the caster's hit points become the targets.
func (e *RuleState) spellUndeadTargets() {
	e.spellValue = e.Acting().HP
	e.NSpellTargets = 0
	for _, p := range e.Party {
		if p.Kind == 0xE && e.spellValue >= p.HP {
			e.spellValue -= p.HP
			e.NSpellTargets++
			e.SetTgt(e.NSpellTargets, p)
		}
	}
	e.applySpell(int(e.CurSpell), 0, 0, 0, 0, "is charmed")
}

// spellApplyAnd17 = ovr23:19C3: applies the spell, then runs the affect 0x17 handler on the first target.
func (e *RuleState) spellApplyAnd17() {
	e.applySpell(int(e.CurSpell), 0, 1, 0, 0, "")
	e.callAffect(affSpiritualHammer, e.tgt(1), nil, 0)
}

// spellMirrorImage = ovr23:1A8A: Mirror Image, 1d4 images (in the high
// nibble, mirrorImage's rule).
func (e *RuleState) spellMirrorImage() {
	e.applySpell(int(e.CurSpell), e.Dice(1, 4)*imageStep, 0, 0, 0, "is duplicated")
}

// spellStinkingCloud = ovr23:1B11: Stinking Cloud (clouds.go).
func (e *RuleState) spellStinkingCloud() { e.createCloud(CloudStinking) }

// spellEnlarge = ovr23:1F41: Enlarge/Strength of the first target: 1d4/1d6/1d8
// by its classes, added to its strength, exceptional strength above 18.
func (e *RuleState) spellEnlarge() {
	var ext, add byte // add is an unset local when no class qualifies
	t := e.tgt(1)
	if int8(t.Levels[ClsMage]) > 0 {
		add = byte(e.Dice(1, 4))
	}
	if int8(t.Levels[ClsCleric]) > 0 || int8(t.Levels[ClsThief]) > 0 {
		add = byte(e.Dice(1, 6))
	}
	if FighterKind(t) {
		add = byte(e.Dice(1, 8))
	}
	str := t.Stats[StatStr] + Score(add)
	if str > 0x12 {
		if FighterKind(t) {
			ext = byte(int(str-0x12)*10 + int(t.Str00))
			if ext > 0x64 {
				ext = 0x64
			}
			str = 0x12
		} else {
			str = 0x12
		}
	}
	out, ok := e.SetStrength(t, str, ext)
	if ok {
		e.CombatSay(t, "is stronger", true)
	}
	e.AddAffect(t, affStrength, e.spellAmount(int(e.CurSpell)), out, 1)
}

// spellAnimateDead = ovr23:206F: Animate Dead. Corpses of the party list (status 6,
// kind 0) are put back on their feet on the caster's side, one per caster level.
func (e *RuleState) spellAnimateDead() {
	e.areaSpell = 1
	lvl := e.casterLevel(int(e.CurSpell))
	e.NSpellTargets = 0
	sel := e.Acting()
	for _, p := range e.Party {
		if lvl <= 0 {
			break
		}
		if p.Status != StatusDead || p.Kind != 0 || !e.combatSquareFree(p, e.CombatX(p), e.CombatY(p), true) {
			continue
		}
		data := byte(int(p.Side)<<4 + e.casterLevel(int(e.CurSpell)))
		p.Side = sel.Side
		p.Auto = 1
		p.Undead = 2
		p.FightLevel = 0
		p.BaseMove = 6
		for i := 0; i <= 0x14; i++ {
			p.Memorized[i] = 0
		}
		if p.Control > 0x7F {
			p.Control = 0xB2
		} else {
			p.Control = 0xB3
		}
		p.Kind = 4
		if e.BeginFlag == 5 {
			p.Combat.Target = nil
		}
		lvl--
		if e.Revive(p, p.MaxHP) {
			e.AddAffectMsg(p, affAnimated, 0, data, 1, 0, false, "is animated")
			p.Status = StatusAnimated
		}
	}
}

// spellCureBlind = ovr23:2217: cures blindness (affect 0x21): "can see".
func (e *RuleState) spellCureBlind() {
	if e.CureAffect(e.tgt(1), 0x21) {
		e.hitMessage(e.tgt(1), true, "can see")
	}
}

// cureDiseaseAffects = ovr23:228A: with the cure flag (DS:681F) set, cures
// affects 0x22, 0x2B (and 0x2C, 0x1F) and 0x32 (and 0x39) of the first target.
func (e *RuleState) cureDiseaseAffects() bool {
	cured := false
	e.Rv.Cure = 1
	t := e.tgt(1)
	if e.CureAffect(t, affCauseDisease) {
		cured = true
	}
	if e.CureAffect(e.tgt(1), 0x2B) {
		cured = true
		e.RemoveAffect(e.tgt(1), 0x2C, nil)
		e.RemoveAffect(e.tgt(1), 0x1F, nil)
	}
	if e.CureAffect(e.tgt(1), 0x32) {
		cured = true
		e.RemoveAffect(e.tgt(1), 0x39, nil)
	}
	e.Rv.Cure = 0
	return cured
}

// spellCureDisease = ovr23:232F: Cure Blindness spell body (the result is unused).
func (e *RuleState) spellCureDisease() { e.cureDiseaseAffects() }

// spellDispel = ovr23:2385: Dispel Magic. Each pass walks the affects of the
// first target (the loop variable only counts passes); every temporary
// affect goes with a chance of 50% +5% per level above, -10% per level
// below, the level of the affect being its data low nibble.
func (e *RuleState) spellDispel() {
	e.areaSpell = 1
	lvl := e.casterLevel(int(e.CurSpell))
	for i := 1; i <= e.NSpellTargets; i++ {
		hit := false
		t := e.tgt(i) // fixed: the original read the first target on every pass
		if t == nil {
			continue
		}
		for _, a := range append([]*Affect(nil), t.Affects...) {
			if a.Data == 0xFF {
				continue
			}
			e.spellValue2 = a.Data & 0xF
			switch {
			case byte(lvl) > e.spellValue2:
				e.spellValue = byte((lvl-int(e.spellValue2))*5 + 0x32)
			case byte(lvl) < e.spellValue2:
				e.spellValue = byte(0x32 - (int(e.spellValue2)-lvl)*2)
			default:
				e.spellValue = 0x32
			}
			if byte(e.Dice(1, 100)) <= e.spellValue {
				e.RemoveAffect(t, a.Type, a)
				hit = true
			}
		}
		if hit {
			e.hitMessage(t, true, "is affected")
		}
	}
}

// spellPrayer = ovr23:24D6: "is praying", cast at caster level with the side in the high nibble.
func (e *RuleState) spellPrayer() {
	e.applySpell(int(e.CurSpell), int(e.Acting().Side)<<4+e.casterLevel(int(e.CurSpell)), 0, 0, 0, "is praying")
}
