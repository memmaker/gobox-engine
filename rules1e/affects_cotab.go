package rules1e

// cotabHandlers are CotAB's affect handlers (fold T4c): its variants of
// POR's affects and its own, by the names CotabAffectVariants gives them,
// each ported exactly from cotab-go/rules (affects.go, O064 and O08D) and
// checked against it there (cotab-go/rules/affects1e_test.go). Where the
// game's handler names an affect, the handler names the unified id
// CotabAffectIDs gives that byte.
var cotabHandlers = map[HandlerName]AffectHandler{
	"constitutionItemCotab":    (*RuleState).constitutionItemCotab,
	"scoreUpItemCotab":         (*RuleState).scoreUpItemCotab,
	"freeActionDexCotab":       (*RuleState).freeActionDexCotab,
	"dexterityDownItemCotab":   (*RuleState).dexterityDownItemCotab,
	"lowIntelligenceItemCotab": (*RuleState).lowIntelligenceItemCotab,
	"lowStrengthMindItemCotab": (*RuleState).lowStrengthMindItemCotab,
	"itemGrantsAffectCotab":    (*RuleState).itemGrantsAffectCotab,
	"sticksToSnakes":           (*RuleState).sticksToSnakes,
	"wardEnchantedCotab":       (*RuleState).wardEnchantedCotab,
	"armourBonusCotab":         (*RuleState).armourBonusCotab,
	"suffocatingCotab":         (*RuleState).suffocatingCotab,
	"charismaCotab":            (*RuleState).noEffectCotab,
	"gnomeVsManSizedCotab":     (*RuleState).gnomeVsManSizedCotab,
	"invisibleCotab":           (*RuleState).invisibleCotab,
	"dwarfVsOrcCotab":          (*RuleState).dwarfVsOrcCotab,
	"stunnedCotab":             (*RuleState).stunnedCotab,
	"confusedCotab":            (*RuleState).confusedCotab,
	"stumblingCotab":           (*RuleState).stumblingCotab,
	"vsGiantsCotab":            (*RuleState).vsGiantsCotab,
	"gnomeExtraCotab":          (*RuleState).gnomeExtraCotab,
	"fireBodyCotab":            (*RuleState).fireBodyCotab,
	"iceBodyCotab":             (*RuleState).iceBodyCotab,
	"engulfsCotab":             (*RuleState).engulfsCotab,
	"magicWeaponsOnlyCotab":    (*RuleState).magicWeaponsOnlyCotab,
	"minorGlobeCotab":          (*RuleState).minorGlobeCotab,
	"feeblemindedCotab":        (*RuleState).feeblemindedCotab,
	"hiddenFromKind19Cotab":    (*RuleState).hiddenFromKind19Cotab,
	"unaffectedByType32Cotab":  (*RuleState).unaffectedByType32Cotab,
	"miscastingCotab":          (*RuleState).noEffectCotab,
	"vsKind3Cotab":             (*RuleState).vsKind3Cotab,
	"vsKind8Cotab":             (*RuleState).vsKind8Cotab,
	"berserkCotab":             (*RuleState).berserkCotab,
	"halfDamageCotab":          (*RuleState).halfDamageCotab,
	"saveVsElementsCotab":      (*RuleState).saveVsElementsCotab,
	"electricityHealsCotab":    (*RuleState).electricityHealsCotab,
	"arrowsScratchCotab":       (*RuleState).arrowsScratchCotab,
	"spitAcidCotab":            (*RuleState).noEffectCotab, // CotAB's table has no handler for 86
	"deathRays":                (*RuleState).deathRays,
	"breathAcidCotab":          (*RuleState).noEffectCotab, // none for 90
	"hugsCotab":                (*RuleState).hugsCotab,
	"burnsCotab":               (*RuleState).burnsCotab,
	"magicResist50Cotab":       (*RuleState).magicResist50Cotab,
	"magicResist15Cotab":       (*RuleState).magicResist15Cotab,
	"slingStonesCotab":         (*RuleState).slingStonesCotab,
	"avoidsBowsCotab":          (*RuleState).avoidsBowsCotab,
	"paralysingGazeCotab":      (*RuleState).paralysingGazeCotab,
	"fireTouchCotab":           (*RuleState).fireTouchCotab,
	"gazeParalyseCotab":        (*RuleState).noEffectCotab, // none for 126
	"breathFireCotab":          (*RuleState).noEffectCotab, // none for 128
	"magicImmuneCotab":         (*RuleState).magicImmuneCotab,
	"slainByArrowCotab":        (*RuleState).slainByArrowCotab,
	"breathFireCloseCotab":     (*RuleState).noEffectCotab, // none for 131
	"throwLightningCotab":      (*RuleState).noEffectCotab, // none for 132
	"noMindAffectsCotab":       (*RuleState).noMindAffectsCotab,
	"electricityImmuneCotab":   (*RuleState).electricityImmuneCotab,
	"rootedCotab":              (*RuleState).rootedCotab,
	"berserkFoeCotab":          (*RuleState).berserkFoeCotab,
	"alwaysInvisibleCotab":     (*RuleState).alwaysInvisibleCotab,
	"engulfingCotab":           (*RuleState).engulfingCotab,
	"paladinCures":             (*RuleState).paladinCures,
	"zapsAttackersCotab":       (*RuleState).zapsAttackersCotab,
	"huggingCotab":             (*RuleState).huggingCotab,
	"dispelEvilCotab":          (*RuleState).dispelEvilCotab,
	"giantStrengthCotab":       (*RuleState).noEffectCotab, // the Stronger spell's score is CotAB's stats.go's
	"thieveryItemCotab":        (*RuleState).thieveryItemCotab,
}

// CotAB's values the handlers test.
const (
	beginFlagCombat = 5 // BeginFlag in a fight (CotAB's gameState stateCombat)

	controlHiredCotab   = 0x7F // above: a hired NPC
	controlCharmedCotab = 0xB3
	controlBerserkCotab = 0xB2

	// Creature flags (Character.CreatureFlags, CotAB +331).
	creatureEnchantedCotab = 1
	creatureGoblinoidCotab = 2
	creatureOrcCotab       = 4
	creatureGiantCotab     = 8

	dmgType32Cotab = 0x20
	dmgType64Cotab = 0x40

	kindManSizedCotab = 1
	kindGiantACotab   = 2
	kindKind3Cotab    = 3
	kindKind8Cotab    = 8
	kindGiantBCotab   = 10
	kindKind19Cotab   = 19
	sizeMaskCotab     = 0x7F
	sizeLargeCotab    = 2

	acCeilingCotab     = 60
	acArmourBonusCotab = 2
	imageShiftCotab    = 4
	runAwayTimeCotab   = 10
	berserkTimeCotab   = 1
	magicWeaponFull    = 3
	globeMaxLevelCotab = 4
	feebleScoreCotab   = 7
	spellLevelField    = 1
	spellSaveField     = 8 // what a saving throw does (0: nothing)
	paladinCureLevels  = 5
	resistTopLevel     = 11 // magic resistance changes 5% per caster level under it
	resistPerLevel     = 5

	itemUsesArrowsCotab   = 1
	itemTypeSlingCotab    = ItemType(85)
	itemTypeLongBowCotab  = ItemType(87)
	itemTypeShortBowCotab = ItemType(88)
	itemTypeArrowCotab    = ItemType(28)
	slayingArrowWordCotab = WordID(135)

	engulfAttacksCotab = 2
	hugAttacksCotab    = 1
	grabDiceCotab      = 2
	grabSidesCotab     = 8
	hugMinRollCotab    = 18

	rayTriesCotab          = 4
	disintegrateRangeCotab = 2
	stoneRangeCotab        = 3
	deathRayRangeCotab     = 4
	woundRangeCotab        = 5
)

// inCombat: CotAB's gameState == stateCombat.
func (e *RuleState) inCombat() bool { return e.BeginFlag == beginFlagCombat }

func (e *RuleState) noEffectCotab(c *Character, a *Affect, flag byte) {}

// itemGrantsAffectCotab = O064:26D3 itemHandler: the readied item grants
// the affect its +0x3D names (data 255, with its removal call), run at once
// outside a fight.
func (e *RuleState) itemGrantsAffectCotab(c *Character, a *Affect, flag byte) {
	typ := AffectType(affectItem(a).Spells[1])
	if flag != 0 {
		e.RemoveAffect(c, typ, nil)
		return
	}
	e.AddAffect(c, typ, 0, 0xFF, 1)
	if !e.inCombat() {
		e.callAffect(typ, c, nil, 0)
	}
}

// sticksToSnakes = CotAB O064:00DC, SotSB, PoD O07C snakesHandler (one rule): the snakes eat the character's attacks.
func (e *RuleState) sticksToSnakes(c *Character, a *Affect, flag byte) {
	n := byte(c.Attacks[0] + c.Attacks[1])
	if a.Data > n {
		a.Data -= n
	} else {
		e.RemoveAffect(c, affSticksToSnakes, nil)
	}
	e.hitMessage(c, true, "is fighting with snakes")
	e.ClearCombatInfo(c)
}

// wardEnchantedCotab = O064:015F.
func (e *RuleState) wardEnchantedCotab(c *Character, a *Affect, flag byte) {
	if e.Acting().CreatureFlags&creatureEnchantedCotab != 0 {
		e.Rv.HitRoll -= 7
	}
}

// armourBonusCotab = O064:01DF: two better, up to armour class 0.
func (e *RuleState) armourBonusCotab(c *Character, a *Affect, flag byte) {
	for i := range c.AC {
		if c.AC[i] < acCeilingCotab-acArmourBonusCotab {
			c.AC[i] += acArmourBonusCotab
		} else {
			c.AC[i] = acCeilingCotab
		}
	}
}

// suffocatingCotab = O064:0379.
func (e *RuleState) suffocatingCotab(c *Character, a *Affect, flag byte) {
	if a.Data == 0 {
		e.killWithMessage(c, StatusDead, "Suffocates")
		return
	}
	a.Data--
}

// gnomeVsManSizedCotab = O064:044A: +1 to hit a goblinoid.
func (e *RuleState) gnomeVsManSizedCotab(c *Character, a *Affect, flag byte) {
	if targetOf(c).CreatureFlags&creatureGoblinoidCotab != 0 {
		e.Rv.HitRoll++
	}
}

// invisibleCotab = O064:06C6: attackers who cannot see it are hampered.
func (e *RuleState) invisibleCotab(c *Character, a *Affect, flag byte) {
	if flag == 0 && e.FindAffect(e.Acting(), affDetectInvisible) == nil {
		e.AffectVeto = 1
		e.Rv.HitRoll -= 4
	}
}

// dwarfVsOrcCotab = O064:06FB: +1 to hit an orc.
func (e *RuleState) dwarfVsOrcCotab(c *Character, a *Affect, flag byte) {
	if targetOf(c).CreatureFlags&creatureOrcCotab != 0 {
		e.Rv.HitRoll++
	}
}

// stunnedCotab = O064:0071: the turn is over.
func (e *RuleState) stunnedCotab(c *Character, a *Affect, flag byte) { e.ClearCombatInfo(c) }

// confusedCotab = O064:09D9: each round the confused run, stand dazed, go
// berserk or rage; a saving throw may end it.
func (e *RuleState) confusedCotab(c *Character, a *Affect, flag byte) {
	switch r := e.Dice(1, 100); {
	case r <= 10:
		e.RemoveAffect(c, affConfusedCotab, nil)
		ci := combatInfo(c)
		ci.Turned = 1
		c.Auto = 1
		if c.Control <= controlHiredCotab {
			c.Control = controlCharmedCotab
		}
		ci.Target = nil
		e.AddAffectMsg(c, affFrightened, runAwayTimeCotab, 0, 1, 1, false, "runs away")
	case r <= 60:
		e.hitMessage(c, true, "is confused")
		e.ClearCombatInfo(c)
	case r <= 80:
		e.AddAffectMsg(c, affBerserkFoeCotab, berserkTimeCotab, c.Side, 1, 1, false, "goes berserk")
		e.callAffect(affBerserkFoeCotab, c, nil, 0)
	default:
		e.hitMessage(c, true, "is enraged")
	}
	if e.SavingThrow(c, SaveVsSpell, -2) {
		e.RemoveAffect(c, affConfusedCotab, nil)
	}
}

// stumblingCotab = O064:0B80: with no initiative left the attack misses.
func (e *RuleState) stumblingCotab(c *Character, a *Affect, flag byte) {
	if combatInfo(c).Delay == 0 {
		e.AffectVeto = 1
		e.Rv.HitRoll = -1
	}
}

// vsGiantsCotab = O064:1153: large giants hit a dwarf or gnome less.
func (e *RuleState) vsGiantsCotab(c *Character, a *Affect, flag byte) {
	sel := e.Acting()
	if (sel.Kind == kindGiantACotab || sel.Kind == kindGiantBCotab) && sel.Size&sizeMaskCotab == sizeLargeCotab {
		e.Rv.HitRoll -= 4
	}
}

// gnomeExtraCotab = O064:11A2.
func (e *RuleState) gnomeExtraCotab(c *Character, a *Affect, flag byte) {
	sel := e.Acting()
	if sel.Kind == kindManSizedCotab && sel.Size&sizeMaskCotab == sizeLargeCotab {
		e.Rv.HitRoll -= 4
	}
}

// fireBodyCotab = O064:1212.
func (e *RuleState) fireBodyCotab(c *Character, a *Affect, flag byte) {
	if e.Rv.DmgType&dmgCold != 0 {
		e.Rv.Save += 2
		return
	}
	if e.Rv.DmgType&DmgFire != 0 && !e.Rv.SaveMade {
		e.Rv.Damage *= 2
	}
}

// iceBodyCotab = O064:1243 (the saving throw, as cotab-go tests it).
func (e *RuleState) iceBodyCotab(c *Character, a *Affect, flag byte) {
	if e.Rv.DmgType&DmgFire != 0 {
		e.Rv.Save += 2
		return
	}
	if e.Rv.DmgType&dmgCold != 0 && !e.Rv.SaveMade {
		e.Rv.Damage *= 2
	}
}

// magicWeaponsOnlyCotab = O064:12D8: plain weapons do nothing, weak magic
// ones half.
func (e *RuleState) magicWeaponsOnlyCotab(c *Character, a *Affect, flag byte) {
	w := e.readiedWeapon(e.Acting())
	if w == nil || w.Plus == 0 {
		e.Rv.Damage = 0
		return
	}
	if int8(w.Plus) < magicWeaponFull {
		e.Rv.Damage >>= 1
	}
}

// minorGlobeCotab = O064:13CA: spells under level 4 are stopped.
func (e *RuleState) minorGlobeCotab(c *Character, a *Affect, flag byte) {
	if e.CurSpell > 0 && e.T().Spells[e.CurSpell][spellLevelField] < globeMaxLevelCotab {
		e.blockAffect(affAny)
	}
}

// feeblemindedCotab = O064:1595 spell_stupid.
func (e *RuleState) feeblemindedCotab(c *Character, a *Affect, flag byte) {
	c.Stats[StatInt] = feebleScoreCotab
	c.Stats[StatWis] = feebleScoreCotab
	e.CombatSay(c, "is stupid", true)
	if !e.inCombat() {
		return
	}
	ci := combatInfo(c)
	ci.CanCast = 0
	if ci.Spell > 0 {
		e.CombatSay(c, "lost a spell", true)
		e.forgetSpell(c, SpellID(ci.Spell))
		ci.Spell = 0
	}
}

// hiddenFromKind19Cotab = O064:1636.
func (e *RuleState) hiddenFromKind19Cotab(c *Character, a *Affect, flag byte) {
	if e.Acting().Kind != kindKind19Cotab {
		return
	}
	if e.FindAffect(e.Acting(), affDetectInvisible) == nil {
		e.AffectVeto = 1
	}
	e.Rv.HitRoll -= 4
}

// unaffectedByType32Cotab = O064:16D2.
func (e *RuleState) unaffectedByType32Cotab(c *Character, a *Affect, flag byte) {
	if e.Rv.DmgType&dmgType32Cotab > 0 {
		e.blockAffect(affAny)
		e.CombatSay(c, "is unaffected", true)
	}
}

// vsKind3Cotab = O064:171A: a heavy blow against kind 3.
func (e *RuleState) vsKind3Cotab(c *Character, a *Affect, flag byte) {
	if targetOf(c).Kind == kindKind3Cotab {
		e.Rv.Damage = int16(e.Dice(1, 12)*3 + 4 + int(strDamage(c)))
		e.Rv.HitRoll += 2
	}
}

// vsKind8Cotab = O064:1772.
func (e *RuleState) vsKind8Cotab(c *Character, a *Affect, flag byte) {
	if targetOf(c).Kind == kindKind8Cotab {
		e.Rv.HitRoll += 3
		e.Rv.Damage += 3
	}
}

// berserkCotab = O064:17B3 spl_berzerk: attacks the nearest one, of
// either side.
func (e *RuleState) berserkCotab(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		if c.Control == controlCharmedCotab {
			c.Control = 0
		}
		c.Side = 0
		return
	}
	e.goBerserkCotab(c)
	if e.inCombat() {
		combatInfo(c).CanCast = 0
		c.Side = targetOf(c).Side ^ 1
		e.CombatSay(c, "goes berzerk", true)
	}
}

// goBerserkCotab is the start the two berserk handlers share: computer
// controlled, against the nearest combatant.
func (e *RuleState) goBerserkCotab(c *Character) {
	c.Auto = 1
	if c.Control <= controlHiredCotab || c.Control == controlCharmedCotab {
		c.Control = controlCharmedCotab
	} else {
		c.Control = controlBerserkCotab
	}
	if !e.inCombat() {
		return
	}
	ci := combatInfo(c)
	ci.Target = nil
	e.BuildNear(e.CombatX(c), e.CombatY(c), 0xFF, 0xFF, e.CombatSize(c))
	if len(e.Near) > 1 {
		ci.Target = e.Combatant(e.Near[1].Index)
	}
}

// halfDamageCotab = O064:197F.
func (e *RuleState) halfDamageCotab(c *Character, a *Affect, flag byte) { e.Rv.Damage >>= 1 }

// saveVsElementsCotab = O064:1990: fire and cold are halved, or with a
// saving throw against a spell that allows one, stopped.
func (e *RuleState) saveVsElementsCotab(c *Character, a *Affect, flag byte) {
	if e.Rv.DmgType&(DmgFire|dmgCold) == 0 {
		return
	}
	if e.SavingThrow(c, SaveVsSpell, 0) && e.T().Spells[e.CurSpell][spellSaveField] != 0 {
		e.Rv.Damage = 0
		return
	}
	e.Rv.Damage >>= 1
}

// electricityHealsCotab = O064:19E1: the original rolls 1d8 into a local
// and heals a flat 8.
func (e *RuleState) electricityHealsCotab(c *Character, a *Affect, flag byte) {
	if e.Rv.DmgType&dmgElectric != 0 {
		e.blockAffect(affAny)
		e.Dice(1, 8)
		c.HP += 8
	}
}

// arrowsScratchCotab = O064:1A14.
func (e *RuleState) arrowsScratchCotab(c *Character, a *Affect, flag byte) {
	if w := e.readiedWeapon(e.Acting()); w != nil && e.Base(w.Type).Usage == itemUsesArrowsCotab {
		e.Rv.Damage = 1
	}
}

// burnsCotab = O064:205D: d6 more unless the target resists fire.
func (e *RuleState) burnsCotab(c *Character, a *Affect, flag byte) {
	t := targetOf(c)
	if e.FindAffect(t, affResistFire) == nil && e.FindAffect(t, affFireShieldCold) == nil &&
		e.FindAffect(t, AffFireResistance) == nil {
		e.Rv.Damage += int16(e.Dice(1, 6))
	}
}

// resistRollCotab = O064:20EE (base): base% plus (POR: minus) 5% for each
// caster level under 11, against the affect or magic damage.
func (e *RuleState) resistRollCotab(base int) {
	lvl := e.casterLevel(int(e.CurSpell))
	chance := byte(base + (resistTopLevel-lvl)*resistPerLevel)
	if e.Rv.AffVal == 0 && e.Rv.DmgType&DmgMagic == 0 {
		return
	}
	if byte(e.Dice(1, 100)) <= chance {
		e.blockAffect(affAny)
	}
}

// magicResist50Cotab = O064:214F.
func (e *RuleState) magicResist50Cotab(c *Character, a *Affect, flag byte) { e.resistRollCotab(50) }

// magicResist15Cotab = O064:215D.
func (e *RuleState) magicResist15Cotab(c *Character, a *Affect, flag byte) { e.resistRollCotab(15) }

// slingStonesCotab = O064:22F9: a sling does d6+1.
func (e *RuleState) slingStonesCotab(c *Character, a *Affect, flag byte) {
	if w := e.WieldOf(e.Acting()).Weapon; w != nil && w.Type == itemTypeSlingCotab {
		e.Rv.Damage = int16(e.RollDice(1, 6) + 1)
	}
}

// avoidsBowsCotab = O064:23A2.
func (e *RuleState) avoidsBowsCotab(c *Character, a *Affect, flag byte) {
	if w := e.WieldOf(e.Acting()).Weapon; w != nil && (w.Type == itemTypeLongBowCotab || w.Type == itemTypeShortBowCotab) {
		e.avoidDistantAttack(c, 50)
	}
}

// paralysingGazeCotab = O064:2523 spl_paralyze.
func (e *RuleState) paralysingGazeCotab(c *Character, a *Affect, flag byte) {
	t := targetOf(c)
	if !e.SavingThrow(t, SaveVsParalysis, 0) {
		e.AddAffect(t, AffHeld, 0, 0xFF, 0)
		e.CombatSay(t, "is paralyzed", true)
	}
}

// fireTouchCotab = O064:259D (cold, as the original).
func (e *RuleState) fireTouchCotab(c *Character, a *Affect, flag byte) {
	e.Rv.DmgType = dmgCold | DmgMagic
	e.Damage(targetOf(c), e.RollDice(2, 8), 0, false)
}

// magicImmuneCotab = O064:2623.
func (e *RuleState) magicImmuneCotab(c *Character, a *Affect, flag byte) {
	if e.Rv.AffVal == 0 && e.Rv.DmgType&DmgMagic == 0 {
		return
	}
	e.blockAffect(affAny)
}

// slainByArrowCotab = O064:2643: an arrow of slaying kills outright.
func (e *RuleState) slainByArrowCotab(c *Character, a *Affect, flag byte) {
	var it *Item
	if e.AmmoFor(e.Acting(), e.WieldOf(e.Acting()), &it) && it != nil && it.Type == itemTypeArrowCotab && it.Words[2] == slayingArrowWordCotab {
		c.Status = StatusGone
		c.Okay = 0
		c.HP = 0
		e.removeCombatAffects(c)
		e.AffectGroup(c, grpDown)
		if c.Okay == 0 {
			e.combatRemoveFromMap(c)
		}
	}
}

// noMindAffectsCotab = O064:2750.
func (e *RuleState) noMindAffectsCotab(c *Character, a *Affect, flag byte) {
	e.blockAffect(affFrightened)
	e.blockAffect(affEnfeebled)
	e.blockAffect(affFeebleminded)
	if e.Rv.DmgType&dmgType64Cotab != 0 {
		e.blockAffect(affAny)
	}
}

// electricityImmuneCotab = O064:27B5.
func (e *RuleState) electricityImmuneCotab(c *Character, a *Affect, flag byte) {
	if e.Rv.DmgType&dmgElectric != 0 {
		e.blockAffect(affAny)
	}
}

// rootedCotab = O064:27CC.
func (e *RuleState) rootedCotab(c *Character, a *Affect, flag byte) { combatInfo(c).Moves = 0 }

// berserkFoeCotab = O064:27E0: attacks the nearest; afterwards back to the
// side its data keeps.
func (e *RuleState) berserkFoeCotab(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		if c.Control == controlCharmedCotab {
			c.Control = 0
		}
		c.Side = a.Data
		return
	}
	e.goBerserkCotab(c)
	if e.inCombat() {
		c.Side = targetOf(c).Side ^ 1
	}
}

// alwaysInvisibleCotab = O064:28D0 add_affect_19.
func (e *RuleState) alwaysInvisibleCotab(c *Character, a *Affect, flag byte) {
	e.AddAffect(c, affInvisible, 0xFF, 0xFF, 0)
}

// paladinCures = CotAB O064:28EF, SotSB O0A4, PoD O07C:2AF7: when the week
// ends, the cures (one per five paladin levels) are back. Felix's ruling,
// all games (PoD's): the higher of the level and the old class's level when
// it is usable.
func (e *RuleState) paladinCures(c *Character, a *Affect, flag byte) {
	if flag == 0 {
		return
	}
	lvl := c.Levels[ClsPaladin]
	if earlierClassActive(c, ClsMonk) {
		lvl = max(lvl, c.OldLevels[ClsPaladin])
	}
	c.PaladinCures = byte(int(lvl)/paladinCureLevels + 1)
}

// zapsAttackersCotab = O064:2975: a close attacker takes double the damage
// it dealt, as magic.
func (e *RuleState) zapsAttackersCotab(c *Character, a *Affect, flag byte) {
	sel := e.Acting()
	if e.RangeTo(sel, c) >= 2 {
		return
	}
	dmg, typ := e.Rv.Damage, e.Rv.DmgType
	e.Rv.Damage *= 2
	e.Rv.DmgType = DmgMagic
	e.CombatSay(sel, "gets zapped", true)
	e.Damage(sel, int(e.Rv.Damage), 0, false)
	e.Rv.Damage, e.Rv.DmgType = dmg, typ
}

// dispelEvilCotab = O064:2A19 sp_dispel_evil.
func (e *RuleState) dispelEvilCotab(c *Character, a *Affect, flag byte) {
	t := targetOf(c)
	if t.CreatureFlags&creatureEnchantedCotab != 0 && !e.SavingThrow(t, SaveVsSpell, 0) {
		e.killWithMessage(t, StatusGone, "is dispelled")
		e.RemoveAffect(e.Acting(), affDispellingEvil, nil)
		e.RemoveAffect(e.Acting(), affDispelEvil, nil)
		return
	}
	e.CombatSay(t, "resists dispel evil", true)
}

// engulfsCotab = O08D:4032: two hits engulf the target, which then
// suffocates.
func (e *RuleState) engulfsCotab(c *Character, a *Affect, flag byte) {
	t := targetOf(c)
	if e.slotHits[1] != 2 || t.Okay == 0 || e.FindAffect(t, affHeldFast) != nil ||
		e.FindAffect(t, affSuffocatingCotab) != nil {
		return
	}
	e.CombatSay(c, "engulfs "+t.Name(), true)
	e.AddAffect(t, affHeldFast, 0, byte(e.CombatIndex(t)), 0)
	e.callAffect(affHeldFast, t, nil, 0)
	e.AddAffect(t, affSuffocatingCotab, 0, byte(e.Dice(2, 4)), 0)
	e.AddAffect(c, affEngulfingCotab, 0, byte(e.CombatIndex(t)), 1)
}

// holdTargetCotab is the shared body of the engulfing and hugging handlers
// (O08D:45C6 and 46FC): the held one is hit every round until either
// falls; flag set or a fall frees it.
func (e *RuleState) holdTargetCotab(c *Character, a *Affect, flag byte, self AffectType, extra []AffectType, attacks Count) {
	t := e.Combatant(a.Data)
	if t == nil {
		t = &Character{}
	}
	free := func() {
		for _, x := range extra {
			e.RemoveAffect(t, x, nil)
		}
	}
	if flag != 0 || c.Okay == 0 || t.Okay == 0 {
		free()
		if flag == 0 {
			a.Call = 0
			e.RemoveAffect(c, self, a)
		}
		return
	}
	c.Attacks[0] = attacks
	c.Attacks[1] = 0
	c.Dice[0] = grabDiceCotab
	c.Sides[0] = grabSidesCotab
	e.FullAttack(c, t, e.Hand(c), 1, nil)
	e.ClearCombatInfo(c)
	if t.Okay == 0 {
		e.RemoveAffect(c, self, nil)
		free()
	}
}

// engulfingCotab = O08D:45C6.
func (e *RuleState) engulfingCotab(c *Character, a *Affect, flag byte) {
	e.holdTargetCotab(c, a, flag, affEngulfingCotab, []AffectType{affHeldFast, affSuffocatingCotab}, engulfAttacksCotab)
}

// huggingCotab = O08D:46FC.
func (e *RuleState) huggingCotab(c *Character, a *Affect, flag byte) {
	e.holdTargetCotab(c, a, flag, affHuggingCotab, []AffectType{affHeldFast}, hugAttacksCotab)
}

// hugsCotab = O08D:480C: a roll of 18 or more grabs the target.
func (e *RuleState) hugsCotab(c *Character, a *Affect, flag byte) {
	if e.Rv.HitRoll < hugMinRollCotab {
		return
	}
	t := targetOf(c)
	e.CombatSay(c, "hugs "+t.Name(), true)
	e.AddAffect(t, affHeldFast, 0, byte(e.CombatIndex(t)), 0)
	e.callAffect(affHeldFast, t, nil, 0)
	e.AddAffect(c, affHuggingCotab, 0, byte(e.CombatIndex(t)), 1)
}

// deathRays = CotAB O08D:42F2 attack_or_kill and PoD O0A1:4C33 (one rule):
// an eye tyrant's up to four tries, each a ray not yet fired whose reach
// covers the target (disintegrate 2 squares, flesh to stone 3, death 4,
// cause serious wounds 5), then its spells Fear, Slow and Sleep (CotAB cast
// its own numbers 84, 55, 21 as ids: the wrong spells). PoD's wording, damage
// types and direct casts for both; the rays' flight is not drawn.
func (e *RuleState) deathRays(c *Character, a *Affect, flag byte) {
	ci := combatInfo(c)
	var used byte
	tries := rayTriesCotab
	ci.Target = nil
	e.pickTarget(c, 0xFF, 0, 1)
	cast := func(bit byte, spell SpellID) {
		var quit bool
		used |= bit
		e.CurSpell = spell
		e.CastSpell(spell, 1, 1, CastReply{Fight: true}, &quit)
	}
	for {
		e.Rv.DmgType = DmgMagic
		t := ci.Target
		r := e.RangeTo(c, t)
		tries--
		if t != nil {
			switch {
			case r <= disintegrateRangeCotab && used&1 == 0:
				used |= 1
				e.CombatSay(c, "fires a disintegrate ray", true)
				if !e.SavingThrow(t, saveVsBreath, 0) {
					e.killWithMessage(t, StatusGone, "is disintegrated")
				}
				e.pickTarget(c, 0xFF, 0, 0)
			case r <= stoneRangeCotab && used&2 == 0:
				used |= 2
				e.CombatSay(c, "fires a flesh to stone ray", true)
				if !e.SavingThrow(t, saveVsPetrify, 0) {
					e.killWithMessage(t, StatusStoned, "is Stoned")
				}
				e.pickTarget(c, 0xFF, 0, 0)
			case r <= deathRayRangeCotab && used&4 == 0:
				used |= 4
				e.Rv.DmgType = dmgDeathRay
				e.CombatSay(c, "fires a death ray", true)
				if !e.SavingThrow(t, SaveVsParalysis, 0) {
					e.killWithMessage(t, StatusDead, "is killed")
				}
				e.pickTarget(c, 0xFF, 0, 0)
			case r <= woundRangeCotab && used&8 == 0:
				used |= 8
				e.CombatSay(c, "Causes Serious Wounds", true)
				e.CurSpell = spCauseSeriousWounds
				e.Damage(t, e.RollDice(2, 8)+1, 0, false)
				e.pickTarget(c, 0xFF, 0, 0)
			case used&16 == 0:
				cast(16, spFear)
			case used&32 == 0:
				cast(32, spSlow)
			case used&64 == 0:
				cast(64, spSleep)
			}
		}
		if tries <= 0 || ci.Target == nil {
			break
		}
	}
}

// dmgDeathRay is the death ray's damage type (PoD 72: magic and 0x40).
const dmgDeathRay = DmgMagic | dmgType64Cotab

// downAffectsCotab = CotAB DS:downAffectData: the affects that end when a
// combatant goes down (POR's combatAffects differ), as unified ids.
var downAffectsCotab = [...]AffectType{affArmourBonusCotab, AffCharmed, affSuffocatingCotab, affSilenced,
	affSpiritualHammer, affCoughing, affHelpless, affAnimated, affSnakeCharm, AffHeld, affAsleep, affHeldFast,
	AffRegenerating, affFallsDead, affHeal3, affRootedCotab, affBerserkFoeCotab, affEngulfingCotab, affHuggingCotab}

// thieveryItemCotab = O0AF:1B04 case 11 (itemReadyEffect): readied or put
// away, an item of thievery reworks the thief skills (O0FE:0AEA
// thiefSkills): the first readied item of thievery or free action counts.
func (e *RuleState) thieveryItemCotab(c *Character, a *Affect, flag byte) {
	var p ItemPowers
	for _, it := range c.Items {
		if it.Readied != 0 && (it.Spells[2] == ItemPower(itemPowerThievery) || it.Spells[2] == ItemPower(itemPowerFreeActionDex)) {
			p.Thievery, p.FreeAction = it.Spells[2] == ItemPower(itemPowerThievery), it.Spells[2] == ItemPower(itemPowerFreeActionDex)
			break
		}
	}
	lvl := int8(c.Levels[ClsThief])
	if earlierClassActive(c, ClsMonk) {
		lvl += int8(c.OldLevels[ClsThief])
	}
	skills := e.T().ThiefSkills(int(lvl), c.Race, int(c.Stats[StatDex]), p)
	for i := range skills {
		c.Thief[i] = Percent(skills[i])
	}
}
