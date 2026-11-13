package rules1e

import ()

// Affect handlers of GAME.OVR overlay 13 from 0x1C00 to the end, and the
// ovr13:31A2 routine that fills the DS:6828 handler table.

// HandlerName names an affect handler: a game's Tables.Affects points affect
// numbers at handlers by these names (rules.json carries them).
type HandlerName string

// affectHandler is every handler by name.
var affectHandler map[HandlerName]AffectHandler // set by init: the handlers reach callAffect, which reads it

// init joins POR's handlers and the later games' (affects_<game>.go: their
// variants and own affects); a name is one handler.
func init() {
	affectHandler = map[HandlerName]AffectHandler{}
	for _, m := range []map[HandlerName]AffectHandler{porHandlers, cotabHandlers, sotsbHandlers, podHandlers, aimHandlers} {
		for name, h := range m {
			if affectHandler[name] != nil {
				panic("rules1e: two affect handlers named " + string(name))
			}
			affectHandler[name] = h
		}
	}
}

// porHandlers are POR's affect handlers by name.
var porHandlers = map[HandlerName]AffectHandler{
	"CharmGaze":              (*RuleState).CharmGaze,
	"CoughingEnds":           (*RuleState).CoughingEnds,
	"CursedAlignment":        (*RuleState).CursedAlignment,
	"IsHardier":              (*RuleState).IsHardier,
	"IsSilenced":             (*RuleState).IsSilenced,
	"Resist50":               (*RuleState).Resist50,
	"ResistFirePerDie":       (*RuleState).ResistFirePerDie,
	"SaveBonusVsType0":       (*RuleState).SaveBonusVsType0,
	"SaveBonusVsType42":      (*RuleState).SaveBonusVsType42,
	"SpiritualHammerItem":    (*RuleState).SpiritualHammerItem,
	"spiritualHammerPod":     (*RuleState).spiritualHammerPod,
	"SpitAcid":               (*RuleState).SpitAcid,
	"acidDamage1d4":          (*RuleState).acidDamage1d4,
	"addAffect3B":            (*RuleState).addAffect3B,
	"addAffect62":            (*RuleState).addAffect62,
	"agesOnce":               (*RuleState).agesOnce,
	"alwaysMissDelayed":      (*RuleState).alwaysMissDelayed,
	"armourAtLeast":          (*RuleState).armourAtLeast,
	"avoidDistant3C":         (*RuleState).avoidDistant3C,
	"avoidUnmagicWeapon":     (*RuleState).avoidUnmagicWeapon,
	"bestStrength":           (*RuleState).bestStrength,
	"blindedToHit":           (*RuleState).blindedToHit,
	"block30":                (*RuleState).block30,
	"block90":                (*RuleState).block90,
	"blockCold":              (*RuleState).blockCold,
	"blockFire":              (*RuleState).blockFire,
	"blockHeld":              (*RuleState).blockHeld,
	"blockMind":              (*RuleState).blockMind,
	"blockPoisonHold":        (*RuleState).blockPoisonHold,
	"blockSleepCharm":        (*RuleState).blockSleepCharm,
	"bloodDrain":             (*RuleState).bloodDrain,
	"bonusByTargetRace":      (*RuleState).bonusByTargetRace,
	"bonusVsRace4":           (*RuleState).bonusVsRace4,
	"charmSide":              (*RuleState).charmSide,
	"clearCombatBlock":       (*RuleState).clearCombatBlock,
	"clearStinkingCloud":     (*RuleState).clearStinkingCloud,
	"cloudkill":              (*RuleState).clearPoisonCloud,
	"collapses":              (*RuleState).collapses,
	"damageLessQuarter":      (*RuleState).damageLessQuarter,
	"disease":                (*RuleState).disease,
	"dodgeBoulders":          (*RuleState).dodgeBoulders,
	"weakeningTouch":         (*RuleState).weakeningTouch,
	"fallsDead":              (*RuleState).fallsDead,
	"fear":                   (*RuleState).fear,
	"fearLifted":             (*RuleState).fearLifted,
	"firstHitMark":           (*RuleState).firstHitMark,
	"gaseousEscape":          (*RuleState).gaseousEscape,
	"gazeReflected":          (*RuleState).gazeReflected,
	"getUpOrWait":            (*RuleState).getUpOrWait,
	"giantStrengthGirdle":    (*RuleState).giantStrengthGirdle,
	"halveAttacks":           (*RuleState).halveAttacks,
	"halveCold":              (*RuleState).halveCold,
	"halveElectric":          (*RuleState).halveElectric,
	"halveFire":              (*RuleState).halveFire,
	"halveOrCancelUnmagic":   (*RuleState).halveOrCancelUnmagic,
	"halveUnlessPlainWeapon": (*RuleState).halveUnlessPlainWeapon,
	"halveVsMagicWeapon":     (*RuleState).halveVsMagicWeapon,
	"halveVsWeaponFlags":     (*RuleState).halveVsWeaponFlags,
	"heal3":                  (*RuleState).heal3,
	"healOverTime":           (*RuleState).healOverTime,
	"heldNoMoves":            (*RuleState).heldNoMoves,
	"hitBonusDown":           (*RuleState).hitBonusDown,
	"hitBonusUp":             (*RuleState).hitBonusUp,
	"hitBonusVsNames":        (*RuleState).hitBonusVsNames,
	"hitBonusVsTargetNames":  (*RuleState).hitBonusVsTargetNames,
	"holdTarget":             (*RuleState).holdTarget,
	"holdWeakTarget":         (*RuleState).holdWeakTarget,
	"holyWaterDamage":        (*RuleState).holyWaterDamage,
	"isWeakened":             (*RuleState).isWeakened,
	"itemGrantsAffect":       (*RuleState).itemGrantsAffect,
	"magicDamage2d10":        (*RuleState).magicDamage2d10,
	"minus4All":              (*RuleState).minus4All,
	"minus4HitSave":          (*RuleState).minus4HitSave,
	"minus4VsBugbearGnoll":   (*RuleState).minus4VsBugbearGnoll,
	"minus4VsNames":          (*RuleState).minus4VsNames,
	"mirrorImage":            (*RuleState).mirrorImage,
	"missMinus4":             (*RuleState).missMinus4,
	"mostlyInvisible":        (*RuleState).mostlyInvisible,
	"needsGiantStrength":     (*RuleState).needsGiantStrength,
	"noEffect":               (*RuleState).noEffect,
	"oilDamage":              (*RuleState).oilDamage,
	"onlyMagicOrStrong":      (*RuleState).onlyMagicOrStrong,
	"onlyMagicWeapons":       (*RuleState).onlyMagicWeapons,
	"paralyse1d9":            (*RuleState).paralyse1d9,
	"paralyse2d8":            (*RuleState).paralyse2d8,
	"paralyseNotElf":         (*RuleState).paralyseNotElf,
	"plusOneSaveHit":         (*RuleState).plusOneSaveHit,
	"poisonBite":             (*RuleState).poisonBite,
	"poisonBite2":            (*RuleState).poisonBite2,
	"poisonBite4":            (*RuleState).poisonBite4,
	"poisonBiteMinus2":       (*RuleState).poisonBiteMinus2,
	"poisonDamage":           (*RuleState).poisonDamage,
	"poisonEnds":             (*RuleState).poisonEnds,
	"rangerVsGiants":         (*RuleState).rangerVsGiants,
	"prayerSide":             (*RuleState).prayerSide,
	"protFromEvil":           (*RuleState).protFromEvil,
	"protFromGood":           (*RuleState).protFromGood,
	"rake":                   (*RuleState).rake,
	"regenUnlessFireAcid":    (*RuleState).regenUnlessFireAcid,
	"regenerate":             (*RuleState).regenerate,
	"removeEnds17":           (*RuleState).removeEnds17,
	"resist15":               (*RuleState).resist15,
	"resistColdHalf":         (*RuleState).resistColdHalf,
	"resistFireElectric":     (*RuleState).resistFireElectric,
	"resistFireHalf":         (*RuleState).resistFireHalf,
	"resistFireLessPerDie":   (*RuleState).resistFireLessPerDie,
	"reviveFull":             (*RuleState).reviveFull,
	"rots":                   (*RuleState).rots,
	"setCharisma":            (*RuleState).setCharisma,
	"sickDamage":             (*RuleState).sickDamage,
	"standUpBleeding":        (*RuleState).standUpBleeding,
	"startsToTrain":          (*RuleState).startsToTrain,
	"stoneGaze":              (*RuleState).stoneGaze,
	"strikeTied":             (*RuleState).strikeTied,
	"turnInvisible":          (*RuleState).turnInvisible,
	"weakAndSick":            (*RuleState).weakAndSick,
	"breathWeapon":           func(e *RuleState, c *Character, a *Affect, flag byte) { e.BreathWeapon(c, a, int(flag)) },
}

// porAffects is POR's affect number to handler table (ovr13:31A2 and the
// breath slot ovr23:3413 sets). Unlisted numbers keep no handler; noEffect
// are the original's empty handlers (affects that only mark the character),
// itemGrantsAffect the items that grant the affect named in their record (+0x3D).
var porAffects = map[AffectType]HandlerName{
	AffBless:                 "hitBonusUp",
	AffCurse:                 "hitBonusDown",
	affBonusVsRace4:          "bonusVsRace4",
	affStudyingManual:        "startsToTrain",
	AffBonusByTargetRace:     "bonusByTargetRace",
	affTrainingManual:        "IsHardier",
	AffProtEvil:              "protFromEvil",
	affProtGood:              "protFromGood",
	affResistCold:            "resistColdHalf",
	AffCharmed:               "charmSide",
	affEnlarged:              "bestStrength",
	affFriends:               "setCharisma",
	affPoisonDamage:          "poisonDamage",
	affShield:                "armourAtLeast",
	affHitBonusVsNames:       "hitBonusVsNames",
	affResistFire:            "resistFireHalf",
	affSilenced:              "IsSilenced",
	affSlowPoison:            "poisonEnds",
	affSpiritualHammer:       "SpiritualHammerItem",
	affInvisible:             "blindedToHit",
	affHitBonusVsTargetNames: "hitBonusVsTargetNames",
	affMirrorImage:           "mirrorImage",
	AffRangerVsGiants:        "rangerVsGiants",
	affEnfeebled:             "damageLessQuarter",
	affCoughing:              "CoughingEnds",
	affHelpless:              "clearCombatBlock",
	affAnimated:              "collapses",
	AffBlind:                 "minus4All",
	affCauseDisease:          "weakAndSick",
	affPrayerBonus:           "plusOneSaveHit",
	AffBestowCurse:           "minus4HitSave",
	affBlink:                 "alwaysMissDelayed",
	affStrength:              "bestStrength",
	AffHaste:                 "agesOnce",
	affStinkingCloud:         "clearStinkingCloud",
	affBefuddledPod:          "clearCombatBlock", // the one Befuddled (Entangle, fold T3f): loses its turns in every game
	affProtMissiles:          "avoidUnmagicWeapon",
	AffSlow:                  "halveAttacks",
	affWeakened:              "isWeakened",
	affSick:                  "sickDamage",
	affProtEvilRadius:        "protFromEvil",
	affProtGoodRadius:        "protFromGood",
	affMinus4VsNames:         "minus4VsNames",
	affMinus4VsBugbearGnoll:  "minus4VsBugbearGnoll",
	affPrayer:                "prayerSide",
	affMummyDisease:          "healOverTime",
	affSnakeCharm:            "clearCombatBlock",
	AffHeld:                  "clearCombatBlock",
	affAsleep:                "clearCombatBlock",
	affTurnInvisible:         "turnInvisible",
	affRotting:               "rots",
	affHeldFast:              "heldNoMoves",
	AffRegenerating:          "addAffect62",
	AffFireResistance:        "ResistFirePerDie",
	affRegenerate:            "regenerate",
	affPoisonBite:            "poisonBite",
	affPoisonBite4:           "poisonBite4",
	affPoisonBite2:           "poisonBite2",
	affParalyse2d8:           "paralyse2d8",
	affParalyseNotElf:        "paralyseNotElf",
	affParalyse1d9:           "paralyse1d9",
	affPoisonBiteMinus2:      "poisonBiteMinus2",
	AffMonsterInvisible:      "missMinus4",
	affCamouflaged:           "mostlyInvisible",
	affRake:                  "rake",
	affStrikeTied:            "strikeTied",
	affBloodDrain:            "bloodDrain",
	affHoldTarget:            "holdTarget",
	affHoldWeakTarget:        "holdWeakTarget",
	affGetUp:                 "getUpOrWait",
	affMagicDamage2d10:       "magicDamage2d10",
	affAcidDamage:            "acidDamage1d4",
	affFear:                  "fear",
	affFearLifted:            "fearLifted",
	affStoneGaze:             "stoneGaze",
	affCharmGaze:             "CharmGaze",
	affWeakeningTouch:        "weakeningTouch",
	affWeakeningTouch2:       "weakeningTouch",
	affDisease:               "disease",
	affDisplaced:             "firstHitMark",
	affSaveBonusVsType0:      "SaveBonusVsType0",
	affResistFireElectric:    "resistFireElectric",
	affHalveFire:             "halveFire",
	affHalveVsWeaponFlags:    "halveVsWeaponFlags",
	affFallsDead:             "fallsDead",
	affOnlyMagicWeapons:      "onlyMagicWeapons",
	affSaveBonusVsType42:     "SaveBonusVsType42",
	affHeal3:                 "heal3",
	affStandUpBleeding:       "standUpBleeding",
	affRegenUnlessFireAcid:   "regenUnlessFireAcid",
	affStartRegenerating:     "addAffect3B",
	affReviveFull:            "reviveFull",
	affGaseousEscape:         "gaseousEscape",
	affAvoidDistant:          "avoidDistant3C",
	affResist50:              "Resist50",
	affResist15:              "resist15",
	affBlock90:               "block90",
	affBlockSleepCharm:       "blockSleepCharm",
	affBlockHold:             "blockHeld",
	affBlockCold:             "blockCold",
	affBlockPoisonHold:       "blockPoisonHold",
	affBlockFire:             "blockFire",
	affResistFireLessPerDie:  "resistFireLessPerDie",
	affHalveElectric:         "halveElectric",
	affHalveUnlessPlain:      "halveUnlessPlainWeapon",
	affHalveVsMagicWeapon:    "halveVsMagicWeapon",
	affHolyWaterDamage:       "holyWaterDamage",
	affHalveCold:             "halveCold",
	affOnlyMagicOrStrong:     "onlyMagicOrStrong",
	affDodgeBoulders:         "dodgeBoulders",
	affSpitAcid:              "SpitAcid",
	affOilDamage:             "oilDamage",
	affHalveOrCancelUnmagic:  "halveOrCancelUnmagic",
	affBlock30:               "block30",
	affBlockMind:             "blockMind",
	affGazeReflected:         "gazeReflected",
	affGiantStrengthGirdle:   "giantStrengthGirdle",
	affCursedAlignment:       "CursedAlignment",
	affNeedsGiantStrength:    "needsGiantStrength",
	affEndsSpiritualHammer:   "removeEnds17",
	affDetectMagic:           "noEffect",
	affReduced:               "noEffect",
	AffReadMagic:             "noEffect",
	affFindTraps:             "noEffect",
	affDetectInvisible:       "noEffect",
	AffFeatherFall:           "noEffect",
	affRepulsed:              "noEffect",
	AffPoisoned:              "noEffect",
	affResistNormalWeapons:   "noEffect",
	affMinorGlobe:            "noEffect",
	affUnknown5C:             "noEffect",
	affReflectableGaze:       "noEffect",
	AffItem0:                 "itemGrantsAffect",
	affItem1:                 "itemGrantsAffect",
	affItem2:                 "itemGrantsAffect",
	affItem5:                 "itemGrantsAffect",
	affItem6:                 "itemGrantsAffect",
	affItem8:                 "itemGrantsAffect",
	affItemA:                 "itemGrantsAffect",
	affItemB:                 "itemGrantsAffect",
	AffBreathWeapon:          "breathWeapon",
}

// itemBaseFlags is DS:5589, the base record's Material, for the item's type.
func (e *RuleState) itemBaseFlags(it *Item) byte { return e.Base(it.Type).Material }

// affectItem is the item standing in for the affect record (the original
// reads the item's bytes through the affect pointer).
func affectItem(a *Affect) *Item {
	if a.Item != nil {
		return a.Item
	}
	return &Item{}
}

// GiantRule is how a game tells a giant (rules.json giantRule): POR by the
// target's Kind, the later games by a creature flag (CotAB's giant, SotSB's
// and PoD's ranger foe, all 8).
type GiantRule struct {
	Kind byte          `json:",omitempty"` // the giant Kind (POR 2); 0: Kind does not tell
	Flag CreatureFlags `json:",omitempty"` // the giant creature flag; 0: none
}

var (
	porGiants   = GiantRule{Kind: 2}
	LaterGiants = GiantRule{Flag: 8} // CotAB, SotSB, PoD
)

// isGiant: c is a giant by the game's rule.
func (e *RuleState) isGiant(c *Character) bool {
	g := e.T().Giants
	if g == (GiantRule{}) {
		g = porGiants
	}
	return g.Kind != 0 && c.Kind == g.Kind || c.CreatureFlags&g.Flag != 0
}

// rangerVsGiants: POR's ranger (Felix's ruling, all games): + ranger level
// against giants, at any range (AffRangerVsGiants, attacker group 4).
func (e *RuleState) rangerVsGiants(c *Character, a *Affect, flag byte) {
	if r := int8(c.Levels[ClsRanger]); r > 0 && e.isGiant(targetOf(c)) {
		e.Rv.Damage += int16(r)
	}
}

// halveDamage is the shared "damage = damage div 2" of the resistance handlers.
func (e *RuleState) halveDamage() { e.Rv.Damage >>= 1 }

// stoneGaze = ovr13:1CAD (c, a, flag): a gaze that turns the target to stone; an item with
// 0x76 on a readied piece reflects it back. DS:6826/6828 (the target pointer
// the original keeps in globals) is a local here.
func (e *RuleState) stoneGaze(c *Character, a *Affect, flag byte) {
	t := c.Combat.Target
	e.CombatSay(c, "gazes...", true)
	e.shoot(e.CombatX(c), e.CombatY(c), e.CombatX(t), e.CombatY(t), 0, SpellStoneGaze)
	if e.FindAffect(c, affReflectableGaze) != nil {
		reflected := false
		for _, p := range e.Carried(t) {
			if reflected {
				break
			}
			if (p.Readied != 0 || e.DollSlot(t, p) < 0) && (p.Words[0] == 0x76 || p.Words[1] == 0x76 || p.Words[2] == 0x76) {
				e.CombatSay(t, "reflects it!", true)
				e.shoot(e.CombatX(t), e.CombatY(t), e.CombatX(c), e.CombatY(c), 0, SpellStoneGaze)
				reflected = true
				t = c
			}
		}
	}
	if !e.SavingThrow(t, saveVsPetrify, 0) {
		e.killWithMessage(t, StatusStoned, "is Stoned")
	}
}

// CharmGaze = ovr13:1E77 (c, a, flag): charm gaze; needs ovr14:1085 and a line of sight.
func (e *RuleState) CharmGaze(c *Character, a *Affect, flag byte) {
	t := c.Combat.Target
	tx, ty := e.CombatX(t), e.CombatY(t)
	if !e.MayActAgainst(c, t) {
		return
	}
	rng := 12 // fixed, as the Pascal reference: [bp-2] is never set in the original (0: the gaze barely left the vampire); 12 squares
	if !e.LineOfSight(&rng, &tx, &ty, e.CombatX(c), e.CombatY(c)) {
		return
	}
	e.CombatSay(c, "Gazes...", true)
	e.shoot(e.CombatX(c), e.CombatY(c), e.CombatX(t), e.CombatY(t), 0, SpellCharmGaze)
	was := e.CurSpell
	e.CurSpell = spCharmPerson
	data := byte(int(int8(c.Side))<<7 + 0xC)
	saved := e.SavingThrow(t, SaveVsSpell, 0xFE)
	e.AddAffectMsg(t, AffCharmed, 0, data, 1, 1, saved, "is charmed")
	if x := e.FindAffect(t, AffCharmed); x != nil {
		e.callAffect(AffCharmed, t, x, 0)
	}
	e.CurSpell = was // fixed, as the Pascal reference: the original left Charm Person set, so later hits counted as spells
}

// weakeningTouch = ovr13:23A4/23C2 (Felix's ruling: no game drains
// levels): the touch weakens its target instead (affWeakened).
func (e *RuleState) weakeningTouch(c *Character, a *Affect, flag byte) {
	if t := targetOf(c); t != nil {
		e.callAffect(affWeakened, t, a, 0)
	}
}

// disease = ovr13:23EC (c, a, flag): disease; the target stops healing (0x32) and gets
// the wasting affect 0x39.
func (e *RuleState) disease(c *Character, a *Affect, flag byte) {
	t := c.Combat.Target
	if e.FindAffect(t, affMummyDisease) == nil {
		e.AddAffect(t, affMummyDisease, 0x3840, 0xFF, 1)
		e.AddAffect(t, affRotting, 0xA8C0, byte(e.Dice(1, 6)<<4+0xF), 1)
		e.CombatSay(t, "is Diseased", true)
	}
}

// firstHitMark = ovr13:2490 (c, a, flag): marks the affect (bit 4 of its data) the first time
// it is rolled with a hit roll set.
func (e *RuleState) firstHitMark(c *Character, a *Affect, flag byte) {
	if e.CombatRound == 0 && e.Rv.HitRoll == 0 {
		a.Data &= 0x0F
	} else if a.Data&0x10 == 0 {
		e.Rv.HitRoll = -1
		a.Data |= 0x10
	}
}

// levelBonus = EffHndlr.ConSaveBonus: the Constitution save bonus, 1..5 for
// 4-6, 7-10, 11-13, 14-17 and from 18, 0 below 4 (the Pascal's fix of the
// original's unset bonus).
func levelBonus(level Score) byte {
	switch {
	case level >= 4 && level <= 6:
		return 1
	case level >= 7 && level <= 10:
		return 2
	case level >= 11 && level <= 13:
		return 3
	case level >= 14 && level <= 17:
		return 4
	case level >= 18:
		return 5
	}
	return 0
}

// SaveBonusVsType0 = ovr13:24D9 (c, a, flag): saving throw bonus by level for save type 0.
func (e *RuleState) SaveBonusVsType0(c *Character, a *Affect, flag byte) {
	if e.Rv.SaveType == SaveVsParalysis {
		e.Rv.Save += levelBonus(c.Stats[StatCon])
	}
}

// resistFireElectric = ovr13:2556 (c, a, flag): fully resisted by Magic Missile, fire/electric (4)
// otherwise only drops the affect being added.
func (e *RuleState) resistFireElectric(c *Character, a *Affect, flag byte) {
	if e.CurSpell == SpMagicMissile {
		e.blockAffect(affAny)
		e.CombatSay(c, "is unaffected", true)
	} else if e.Rv.DmgType&dmgElectric != 0 {
		e.blockAffect(affAny)
	}
}

// halveFire = ovr13:25A2 (c, a, flag): fire damage (bit 0) is halved.
func (e *RuleState) halveFire(c *Character, a *Affect, flag byte) {
	if e.Rv.DmgType&DmgFire != 0 {
		e.halveDamage()
	}
}

// halveVsWeaponFlags = ovr13:25BC (c, a, flag): halves damage when the selected character's weapon
// has base flags 0x81.
func (e *RuleState) halveVsWeaponFlags(c *Character, a *Affect, flag byte) {
	if w := e.readiedWeapon(e.Acting()); w != nil && e.itemBaseFlags(w)&0x81 != 0 {
		e.halveDamage()
	}
}

// fallsDead = ovr13:2610 (c, a, flag): the affect ends; a living character falls dead.
func (e *RuleState) fallsDead(c *Character, a *Affect, flag byte) {
	a.Call = 0
	if c.Okay != 0 {
		e.killWithMessage(c, StatusDead, "Falls dead")
	}
}

// onlyMagicWeapons = ovr13:264B (c, a, flag): damage is cancelled unless the weapon is magical
// (+0x32 non-zero, or 0xB1 in +0x31).
func (e *RuleState) onlyMagicWeapons(c *Character, a *Affect, flag byte) {
	w := e.readiedWeapon(e.Acting())
	if w == nil || w.Plus == 0 && w.Words[2] != 0xB1 {
		e.Rv.Damage = 0
	}
}

// SaveBonusVsType42 = ovr13:268A (c, a, flag): like 24D9 for save types 4 and 2.
func (e *RuleState) SaveBonusVsType42(c *Character, a *Affect, flag byte) {
	if e.Rv.SaveType == SaveVsSpell || e.Rv.SaveType == saveVsWand {
		e.Rv.Save += levelBonus(c.Stats[StatCon])
	}
}

// heal3 = ovr13:2700 (c, a, flag): +3 hit points up to the maximum.
func (e *RuleState) heal3(c *Character, a *Affect, flag byte) {
	c.HP += 3
	if c.HP > c.MaxHP {
		c.HP = c.MaxHP
	}
}

// standUpBleeding = ovr13:2730 (c, a, flag): a dying or unconscious character gets back up with
// the hit points it had left to bleed (6 minus the rounds), then 0x5F runs
// for 1d4+1 and this affect (0x63) is removed.
func (e *RuleState) standUpBleeding(c *Character, a *Affect, flag byte) {
	var n byte
	if c.Status == StatusDying && c.Combat.Dying < 6 {
		n = 6 - c.Combat.Dying
	}
	if c.Status == StatusUnconscious {
		n = 6
	}
	if n > 0 && e.Revive(c, n) {
		e.AddAffect(c, affFallsDead, e.Dice(1, 4)+1, 0xFF, 1)
		a.Call = 0
		e.RemoveAffect(c, affStandUpBleeding, a)
	}
}

// regenUnlessFireAcid = ovr13:27D9 (c, a, flag): damage that is neither fire (1) nor acid (0x10)
// starts affect 0x66 for 3d6.
func (e *RuleState) regenUnlessFireAcid(c *Character, a *Affect, flag byte) {
	if e.Rv.DmgType&DmgFire == 0 && e.Rv.DmgType&dmgAcid == 0 {
		e.AddAffect(c, affReviveFull, e.Dice(3, 6), 0xFF, 1)
	}
}

// addAffect3B = ovr13:2814 (c, a, flag): adds 0x3B for 3 rounds unless 0x62 or 0x3B is there.
func (e *RuleState) addAffect3B(c *Character, a *Affect, flag byte) {
	if e.FindAffect(c, affHeal3) == nil && e.FindAffect(c, AffRegenerating) == nil {
		e.AddAffect(c, AffRegenerating, 3, 0xFF, 1)
	}
}

// reviveFull = ovr13:2866 (c, a, flag): back up at full hit points, else re-add 0x66.
func (e *RuleState) reviveFull(c *Character, a *Affect, flag byte) {
	if !e.Revive(c, c.MaxHP) {
		e.AddAffectNow(c, affReviveFull, a.Data, 1)
	}
}

// gaseousEscape = ovr13:28BF (c, a, flag): turns gaseous and leaves the fight.
func (e *RuleState) gaseousEscape(c *Character, a *Affect, flag byte) {
	a.Call = 0
	e.KillCharacter(c, StatusRunning, "Turns gaseous and escapes")
}

// avoidDistant3C = ovr13:2901 (c, a, flag).
func (e *RuleState) avoidDistant3C(c *Character, a *Affect, flag byte) { e.avoidDistantAttack(c, 0x3C) }

// resistRoll = ovr13:2915 (base): the chance base - 5*(11 - caster level)
// percent (as a byte) to drop the affect being added.
func (e *RuleState) resistRoll(base byte) {
	lvl := e.casterLevel(int(e.CurSpell)) & 0xFF
	// an item casts at level 6 (casterLevel), as the original (ruling 15)
	thr := byte(int(base) - 5*(0xB-lvl))
	if e.Rv.AffVal != 0 || e.Rv.DmgType&DmgMagic != 0 {
		if byte(e.Dice(1, 100)) <= thr {
			e.blockAffect(affAny)
		}
	}
}

// Resist50 = ovr13:2976 (c, a, flag).
func (e *RuleState) Resist50(c *Character, a *Affect, flag byte) { e.resistRoll(0x32) }

// resist15 = ovr13:2984 (c, a, flag).
func (e *RuleState) resist15(c *Character, a *Affect, flag byte) { e.resistRoll(0xF) }

// block90 = ovr13:2992 (c, a, flag): 90%: blocks 0x35 and 0xB.
func (e *RuleState) block90(c *Character, a *Affect, flag byte) {
	if e.Dice(1, 100) <= 0x5A {
		e.blockAffect(affAsleep)
		e.blockAffect(AffCharmed)
	}
}

// blockSleepCharm = ovr13:29B6 (c, a, flag): blocks 0xB and 0x35.
func (e *RuleState) blockSleepCharm(c *Character, a *Affect, flag byte) {
	e.blockAffect(AffCharmed)
	e.blockAffect(affAsleep)
}

// blockHeld = ovr13:29CB (c, a, flag): blocks Hold Person.
func (e *RuleState) blockHeld(c *Character, a *Affect, flag byte) { e.blockAffect(AffHeld) }

// blockCold = ovr13:29D9 (c, a, flag): cold damage (bit 1) is blocked.
func (e *RuleState) blockCold(c *Character, a *Affect, flag byte) {
	if e.Rv.DmgType&dmgCold != 0 {
		e.blockAffect(affAny)
	}
}

// blockPoisonHold = ovr13:29F0 (c, a, flag): blocks poison and Hold Person; saves of type 0 always pass.
func (e *RuleState) blockPoisonHold(c *Character, a *Affect, flag byte) {
	e.blockAffect(AffPoisoned)
	e.blockAffect(AffHeld)
	if e.Rv.SaveType == SaveVsParalysis {
		e.Rv.Save = 0x64
	}
}

// blockFire = ovr13:2A11 (c, a, flag): fire damage (bit 0) is blocked.
func (e *RuleState) blockFire(c *Character, a *Affect, flag byte) {
	if e.Rv.DmgType&DmgFire != 0 {
		e.blockAffect(affAny)
	}
}

// resistFireLessPerDie = ovr13:2A28 (c, a, flag): fire damage loses 1 per die, never below one per die.
func (e *RuleState) resistFireLessPerDie(c *Character, a *Affect, flag byte) {
	if e.Rv.DmgType&DmgFire != 0 {
		for k := byte(1); k <= e.Rv.DiceN; k++ {
			e.Rv.Damage--
			if e.Rv.Damage < int16(e.Rv.DiceN) {
				e.Rv.Damage = int16(e.Rv.DiceN)
			}
		}
	}
}

// halveElectric = ovr13:2A6D (c, a, flag): electricity damage (bit 2) is halved.
func (e *RuleState) halveElectric(c *Character, a *Affect, flag byte) {
	if e.Rv.DmgType&dmgElectric != 0 {
		e.halveDamage()
	}
}

// halveUnlessPlainWeapon = ovr13:2A87 (c, a, flag): halves damage unless the weapon is a plain
// (flags non-zero, bit 0 clear) base item.
func (e *RuleState) halveUnlessPlainWeapon(c *Character, a *Affect, flag byte) {
	if w := e.readiedWeapon(e.Acting()); w != nil {
		if f := e.itemBaseFlags(w); f == 0 || f&1 != 0 {
			e.halveDamage()
		}
	}
}

// halveVsMagicWeapon = ovr13:2AE6 (c, a, flag): halves damage when the weapon's +0x32 is positive.
func (e *RuleState) halveVsMagicWeapon(c *Character, a *Affect, flag byte) {
	if w := e.readiedWeapon(e.Acting()); w != nil && int8(w.Plus) > 0 {
		e.halveDamage()
	}
}

// holyWaterDamage = ovr13:2B20 (c, a, flag): holy water readied by the
// selected character makes the damage 1d6+1.
func (e *RuleState) holyWaterDamage(c *Character, a *Affect, flag byte) {
	if w := e.Acting().Slots[SlotWeapon]; w != nil && e.isKind(w.Type, ItemIDHolyWater) {
		e.Rv.Damage = int16(e.RollDice(1, 6) + 1)
	}
}

// halveCold = ovr13:2B63 (c, a, flag): cold damage (bit 1) is halved.
func (e *RuleState) halveCold(c *Character, a *Affect, flag byte) {
	if e.Rv.DmgType&dmgCold != 0 {
		e.halveDamage()
	}
}

// onlyMagicOrStrong = ovr13:2B7D (c, a, flag): damage is cancelled unless the weapon has +0x32
// set, or the attacker is a monster (race 0) of level (+0x73) 4 or more.
func (e *RuleState) onlyMagicOrStrong(c *Character, a *Affect, flag byte) {
	if w := e.readiedWeapon(e.Acting()); w != nil && w.Plus != 0 {
		return
	}
	s := e.Acting()
	if int8(s.Race) > raceMonster || int8(s.MaxLevel) < 4 {
		e.Rv.Damage = 0
	}
}

// dodgeBoulders = ovr13:2BC8 (c, a, flag): a boulder or giant boulder thrown
// by the selected character may be avoided (50%) from afar.
func (e *RuleState) dodgeBoulders(c *Character, a *Affect, flag byte) {
	if w := e.Acting().Slots[SlotWeapon]; w != nil && e.isKind(w.Type, ItemIDBoulder, ItemIDGiantBoulder) {
		e.avoidDistantAttack(c, 0x32)
	}
}

// SpitAcid = ovr13:2C1C (c, a, flag): spits acid (25%) at a target under 4 squares away:
// 8d4 damage, half on a save; the affects 0x79 and 0x50 end.
func (e *RuleState) SpitAcid(c *Character, a *Affect, flag byte) {
	t := c.Combat.Target
	if t == nil || t.Okay == 0 { // fixed, as the Pascal reference for nil; a fallen target: CHANGES row, unclear
		return
	}
	if e.Dice(1, 100) > 0x19 || e.RangeTo(c, t) >= 4 {
		return
	}
	e.ClearCombatInfo(c)
	e.CombatSay(c, "Spits Acid", true)
	e.shoot(e.CombatX(c), e.CombatY(c), e.CombatX(t), e.CombatY(t), 0, SpellSpitAcid)
	dmg := e.RollDice(8, 4)
	saved := e.SavingThrow(t, saveVsBreath, 0)
	e.Damage(t, dmg, saveHalf, saved)
	e.RemoveAffect(c, affSpitAcid, a)
	e.RemoveAffect(c, affAcidDamage, nil)
}

// oilDamage = ovr13:2D38 (c, a, flag): oil readied by the selected character
// makes the damage 3d8; fire or magic types (mask 9) add the dice count.
func (e *RuleState) oilDamage(c *Character, a *Affect, flag byte) {
	if w := e.Acting().Slots[SlotWeapon]; w != nil && e.isKind(w.Type, ItemIDOil) {
		e.Rv.Damage = int16(e.RollDice(3, 8))
	}
	if e.Rv.DmgType&(DmgFire|DmgMagic) != 0 {
		e.Rv.Damage += int16(e.Rv.DiceN)
	}
}

// halveOrCancelUnmagic = ovr13:2D88 (c, a, flag): halves damage with a +0x32 = 0, +0x31 = 0xB1
// weapon, cancels it for another +0x32 = 0 weapon or no weapon.
func (e *RuleState) halveOrCancelUnmagic(c *Character, a *Affect, flag byte) {
	w := e.readiedWeapon(e.Acting())
	switch {
	case w == nil:
		e.Rv.Damage = 0
	case w.Plus == 0 && w.Words[2] == 0xB1:
		e.halveDamage()
	case w.Plus == 0:
		e.Rv.Damage = 0
	}
}

// block30 = ovr13:2DE4 (c, a, flag): 30%: blocks 0xB and 0x35.
func (e *RuleState) block30(c *Character, a *Affect, flag byte) {
	if e.Dice(1, 100) <= 0x1E {
		e.blockAffect(AffCharmed)
		e.blockAffect(affAsleep)
	}
}

// blockMind = ovr13:2E08 (c, a, flag): blocks 0xB, 0x35, 0x34 and 0x37; saves of type 0 pass.
func (e *RuleState) blockMind(c *Character, a *Affect, flag byte) {
	e.blockAffect(AffCharmed)
	e.blockAffect(affAsleep)
	e.blockAffect(AffHeld)
	e.blockAffect(AffPoisoned)
	if e.Rv.SaveType == SaveVsParalysis {
		e.Rv.Save = 0x64
	}
}

// gazeReflected = ovr13:2E37 (c, a, flag): sets DS:681E when the gaze target carries a ring
// (0x98 with 0xFC) or 0x76 item in slots 1..3 of a used item.
func (e *RuleState) gazeReflected(c *Character, a *Affect, flag byte) {
	t := c.Combat.Target
	for _, p := range e.Carried(t) { // extra: a mirror works from the party inventory (it has no slot)
		if e.AffectVeto != 0 {
			break
		}
		if p.Readied != 0 || e.DollSlot(t, p) < 0 {
			for k := 1; k <= 3; k++ {
				if p.Words[k-1] == 0x98 && p.Words[2] == 0xFC || p.Words[k-1] == 0x76 {
					e.AffectVeto = 1
				}
			}
		}
	}
}

// itemGrantsAffect = ovr13:2EDA (c, a, flag): an item stands in for a; its +0x3D is the affect it
// grants (added on use with call 0, then run once; removed with flag set).
func (e *RuleState) itemGrantsAffect(c *Character, a *Affect, flag byte) {
	typ := AffectType(affectItem(a).Spells[1])
	if flag != 0 {
		e.RemoveAffect(c, typ, nil)
		return
	}
	e.AddAffect(c, typ, 0, 0xC, 0)
	e.callAffect(typ, c, nil, 0)
}

// giantStrengthGirdle = ovr13:2F57 (c, a, flag): girdle of giant strength (affect 0x26). Removed
// (flag set): the affect goes unless the character already has 18/00 and
// the stored value cannot be told apart; added: strength 18/00.
func (e *RuleState) giantStrengthGirdle(c *Character, a *Affect, flag byte) {
	e.magicStrength(c, flag, 0x12, 0x64, true)
}

// magicStrength is the girdle's strength (str, str00) while an item is readied
// (flag clear) and gone with it (flag set); say: "is stronger" when it takes.
func (e *RuleState) magicStrength(c *Character, flag byte, str Score, str00 byte, say bool) {
	if flag != 0 {
		found := false
		giant := c.Stats[StatStr] == str && c.Str00 == str00
		for _, p := range append([]*Affect(nil), c.Affects...) {
			if found {
				break
			}
			if p.Type == affStrength {
				s, s00 := strDecode(p)
				// p.Raw[0] is the type byte (the original compares the first byte, not the data)
				if giant && p.Type < 0x80 || s == str && s00 == str00 && !giant {
					e.RemoveAffect(c, affStrength, p)
					found = true
				}
			}
		}
		return
	}
	out, ok := e.SetStrength(c, str, str00)
	if ok && say {
		if e.BeginFlag == 5 {
			e.BeginFlag = 2
			e.CombatSay(c, "is stronger", true)
			e.BeginFlag = 5
		} else {
			e.CombatSay(c, "is stronger", true)
		}
	}
	e.AddAffect(c, affStrength, 0, out, 1)
}

// CursedAlignment = ovr13:30A5 (c, a, flag): cursed item (a stands in as the item): with the
// wrong alignment (+0x3D low nibble vs char +0xA0) it is unreadied and
// hurts by its high nibble.
func (e *RuleState) CursedAlignment(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		return
	}
	it := affectItem(a)
	if byte(it.Spells[1]&0xF) != c.Alignment {
		it.Readied = 0
		n := int(it.Spells[1] >> 4)
		e.Rv.DmgType = DmgMagic
		e.Damage(c, n, 0, false)
	}
}

// needsGiantStrength = ovr13:3134 (c, a, flag): the item needs strength 19 or more.
func (e *RuleState) needsGiantStrength(c *Character, a *Affect, flag byte) {
	if flag == 0 && c.Stats[StatStr] < 0x13 {
		affectItem(a).Readied = 0
		e.notice("Must have Giant Strength")
	}
}

// removeEnds17 = ovr13:317A (c, a, flag): removing it ends affect 0x17.
func (e *RuleState) removeEnds17(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		e.RemoveAffect(c, affSpiritualHammer, nil)
	}
}

// noEffect = ovr13:319B (c, a, flag): empty handler.
func (e *RuleState) noEffect(c *Character, a *Affect, flag byte) {}
