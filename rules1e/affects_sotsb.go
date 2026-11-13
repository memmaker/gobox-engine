package rules1e

import "gobox/core/dnd"

// sotsbHandlers are SotSB's affect handlers (and the ones SotSB and PoD share) (fold T4c): its variants of
// POR's affects and its own, by the names SotsbAffectVariants gives them,
// each ported exactly from sotsb-go/rules and checked against it there.
var sotsbHandlers = map[HandlerName]AffectHandler{
	"swallowsWholeSotsb":      (*RuleState).swallowsWholeSotsb,
	"itemGrantsAffectLater":   (*RuleState).itemGrantsAffectLater,
	"strength1800Item":        (*RuleState).strength1800Item,
	"giantStrengthItem":       (*RuleState).giantStrengthItem,
	"wizardry":                (*RuleState).wizardry,
	"dispellingEvil":          (*RuleState).dispellingEvil,
	"electricityImmuneSotsb":  (*RuleState).electricityImmuneSotsb,
	"gnomeVsFoes":             (*RuleState).gnomeVsFoes,
	"barkskin":                (*RuleState).barkskinRuled, // ruling: spells_sotsbfirst.go
	"delayedBlastSotsb":       (*RuleState).delayedBlastSotsb,
	"friendlySotsb":           (*RuleState).noEffect,
	"invisibleSotsb":          (*RuleState).invisibleSotsb,
	"dwarfVsFoes":             (*RuleState).dwarfVsFoes,
	"befuddledSotsb":          (*RuleState).clearCombatBlock,
	"dispelEvil":              (*RuleState).dispelEvil,
	"confused":                (*RuleState).confused,
	"blinking":                (*RuleState).blinking,
	"dwarfDodges":             (*RuleState).dwarfDodges,
	"gnomeDodges":             (*RuleState).gnomeDodges,
	"fireShieldHot":           (*RuleState).fireShieldHot,
	"fireShieldCold":          (*RuleState).fireShieldCold,
	"globe":                   (*RuleState).globe,
	"needsPlus3WeaponSotsb":   (*RuleState).needsPlus3WeaponSotsb,
	"minorGlobe":              (*RuleState).minorGlobe,
	"stoningTouch":            (*RuleState).stoningTouch,
	"feeblemindedSotsb":       (*RuleState).feeblemindedSotsb,
	"invisibleToAnimalsSotsb": (*RuleState).invisibleToAnimalsSotsb,
	"highlighted":             (*RuleState).highlighted,
	"mirrorSotsb":             (*RuleState).noEffect,
	"dragonBreathImmune":      (*RuleState).dragonBreathImmune,
	"magicalBootsSotsb":       (*RuleState).magicalBootsSotsb,
	"baneWeapon":              (*RuleState).baneWeapon,
	"diseaseImmuneSotsb":      (*RuleState).noEffect,
	"berserk":                 (*RuleState).berserk,
	"magicalStone":            (*RuleState).magicalStone,
	"fireHeals":               (*RuleState).fireHeals,
	"fearAuraSotsb":           (*RuleState).fearAuraSotsb,
	"sotsb84":                 (*RuleState).noEffect,
	"sotsb85":                 (*RuleState).noEffect,
	"sotsb87":                 (*RuleState).noEffect,
	"boltsHarmlessSotsb":      (*RuleState).boltsHarmlessSotsb,
	"fearless":                (*RuleState).fearless,
	"hardyAll":                (*RuleState).hardyAll,
	"sotsb101":                (*RuleState).noEffect,
	"sotsb102":                (*RuleState).noEffect,
	"plainWeaponsImmuneSotsb": (*RuleState).plainWeaponsImmuneSotsb,
	"stunnedSotsb":            (*RuleState).clearCombatBlock,
	"berserkFoe":              (*RuleState).berserkFoe,
	"alwaysInvisibleSotsb":    (*RuleState).alwaysInvisibleSotsb,
	"frightened":              (*RuleState).frightened,
	"zapsAttackers":           (*RuleState).zapsAttackers,
	"giantStrengthSpell":      (*RuleState).giantStrengthSpell,
	"breathFireCloseSotsb":    (*RuleState).breathFireCloseSotsb,
}

// CreatureFlags (SotSB +349, PoD +420) say what kind of foe a creature is.
type CreatureFlags uint8

const (
	creatureEnchanted   CreatureFlags = 1
	creatureGiantKin    CreatureFlags = 4   // dwarves dodge it
	creatureRangerFoe   CreatureFlags = 8   // rangers do extra damage
	creatureGnomeFoe    CreatureFlags = 32  // gnomes dodge it
	creatureDwarfFoe    CreatureFlags = 128 // dwarves hit it better
	creatureAnimalsSeen CreatureFlags = 69  // the kinds invisibility to animals works for
)

// MonsterFlags (SotSB +350, PoD +421) are more kind bits.
type MonsterFlags uint8

const (
	monsterBane     MonsterFlags = 1 // a bane weapon does 6d6 against it
	monsterGnomeFoe MonsterFlags = 4 // gnomes hit it better
)

// SotSB's control values and damage bit beyond POR's.
const (
	controlCharmed = 0xB3
	controlBerserk = 0xB2
	controlHired   = 0x7F // above: a hired NPC
	dmgBreath      = 0x20
)

// SotSB's handler numbers.
const (
	sotsbDispelSlip      = 7   // dispellingEvil: enchanted attackers' to-hit down
	sotsbGlobeMax        = 4   // the globe stops spells up to this level
	sotsbMinorGlobeMax   = 3   // the minor globe
	sotsbPlusNeeded      = 3   // needsPlus3Weapon
	sotsbFeebleScore     = 3   // feeblemind's intelligence and wisdom
	sotsbFearLevel       = 5   // the fear aura spares this level and above
	sotsbPlainLevel      = 4   // monsters below it cannot hurt with a plain weapon
	sotsbParalyseTime    = 100 // paralysing touch
	sotsbHeldData        = 12  // paralysing touch's data
	sotsbRunAwayTime     = 10  // confusion's flight
	sotsbBerserkTime     = 1   // confusion's berserk fit
	sotsbSlowTime        = 3   // electricity slows the fire-healed
	sotsbSpellLevelField = 1   // spell record: level
	sotsbItemUsesBolts   = 0x80
	sotsbGiantStrength   = 18  // giant strength items: 18/00, or 18 + their value
	sotsbStrength00      = 100 // 18/00
	sotsbGiantStepMax    = 6   // the highest value: strength 24
	sotsbAlwaysTime      = 0xFF
	sotsbCloseBreath     = 8 // breathFireClose's damage
	sotsbAlwaysData      = 0xFF
)

// --- shared pieces ----------------------------------------------------------

// unseenBy: the attacker cannot see the character and hits it less.
func (e *RuleState) unseenBy() {
	e.AffectVeto = 1
	e.Rv.HitRoll -= 4
}

// startFleeing: the character turns to run, computer controlled, no target.
func startFleeing(c *Character) {
	ci := combatInfo(c)
	ci.Turned = 1
	c.Auto = 1
	if c.Control <= controlHired {
		c.Control = controlCharmed
	}
	ci.Target = nil
}

// goBerserk: computer controlled, against the nearest combatant.
func (e *RuleState) goBerserk(c *Character) {
	c.Auto = 1
	if c.Control <= controlHired || c.Control == controlCharmed {
		c.Control = controlCharmed
	} else {
		c.Control = controlBerserk
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

// stopSpellsUpTo: spells of the level and below do not get through.
func (e *RuleState) stopSpellsUpTo(flag byte, level SpellProp) {
	if flag == 0 && e.CurSpell > 0 && e.T().Spells[e.CurSpell][sotsbSpellLevelField] <= level {
		e.blockAffect(affAny)
	}
}

// constitutionSaveBonus: the hardy races' bonus, a point per 3.5 points of
// constitution (below 4 the original adds an unset local; none here).
func constitutionSaveBonus(c *Character) byte {
	switch con := c.Stats[StatCon]; {
	case con >= 18:
		return 5
	case con >= 14:
		return 4
	case con >= 11:
		return 3
	case con >= 7:
		return 2
	case con >= 4:
		return 1
	}
	return 0
}

// --- item powers -------------------------------------------------------------

// itemGrantsAffectLater = O0A4:3908: a readied item grants the affect its
// +0x3D names (data 255, with its removal call); unlike POR's it is not run now.
func (e *RuleState) itemGrantsAffectLater(c *Character, a *Affect, flag byte) {
	typ := AffectType(affectItem(a).Spells[1])
	if flag != 0 {
		e.RemoveAffect(c, typ, nil)
		return
	}
	e.AddAffect(c, typ, 0, sotsbAlwaysData, 1)
}

// strength1800Item = SotSB itemReadyEffect's statAffects for item power 3:
// strength 18/00 while readied, silently (POR's girdle says "is stronger").
func (e *RuleState) strength1800Item(c *Character, a *Affect, flag byte) {
	e.magicStrength(c, flag, sotsbGiantStrength, sotsbStrength00, false)
}

// giantStrengthItem = statAffects for item power 5 (CotAB, SotSB, PoD):
// +0x3D 0 gives 18/00, 1..6 strength 19..24, other values nothing.
func (e *RuleState) giantStrengthItem(c *Character, a *Affect, flag byte) {
	switch arg := affectItem(a).Spells[1]; {
	case arg == 0:
		e.magicStrength(c, flag, sotsbGiantStrength, sotsbStrength00, false)
	case arg <= sotsbGiantStepMax:
		e.magicStrength(c, flag, Score(sotsbGiantStrength+arg), 0, false)
	}
}

// wizardry: the ring of wizardry. Felix's ruling (T4c 08:10): in every game
// it doubles every magic-user level's spells, as rules1e's ClassStats does
// (Powers.Wizardry, which who reads from the readied items); readying or
// unreadying it (Readied is set first) works the slots out again.
func (e *RuleState) wizardry(c *Character, a *Affect, flag byte) {
	c.SpellSlots[SpellClassMage] = countSlots(e.T().ClassStats(who(c)).MageSlots)
}

// --- the handlers --------------------------------------------------------------

// dispellingEvil = O0A4:0394: enchanted attackers are hampered.
func (e *RuleState) dispellingEvil(c *Character, a *Affect, flag byte) {
	if flag == 0 && e.Acting().CreatureFlags&creatureEnchanted != 0 {
		e.Rv.HitRoll -= sotsbDispelSlip
	}
}

// electricityImmuneSotsb = O0A4:03B7.
func (e *RuleState) electricityImmuneSotsb(c *Character, a *Affect, flag byte) {
	if flag == 0 && e.Rv.DmgType&dmgElectric != 0 {
		e.blockAffect(affAny)
	}
}

// gnomeVsFoes = O0A4:03D6.
func (e *RuleState) gnomeVsFoes(c *Character, a *Affect, flag byte) {
	if flag == 0 && targetOf(c).MonsterFlags&monsterGnomeFoe != 0 {
		e.Rv.HitRoll++
	}
}

// invisibleSotsb = O0A4:0A03: attackers who cannot see it are hampered.
func (e *RuleState) invisibleSotsb(c *Character, a *Affect, flag byte) {
	if flag == 0 && e.FindAffect(e.Acting(), affDetectInvisible) == nil {
		e.unseenBy()
	}
}

// dwarfVsFoes = O0A4:0A38.
func (e *RuleState) dwarfVsFoes(c *Character, a *Affect, flag byte) {
	if flag == 0 && targetOf(c).CreatureFlags&creatureDwarfFoe != 0 {
		e.Rv.HitRoll++
	}
}

// dispelEvil = O0A4:0C8B: the enchanted foe is sent away unless it saves,
// and the spell is spent.
func (e *RuleState) dispelEvil(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		return
	}
	t := targetOf(c)
	if t.CreatureFlags&creatureEnchanted != 0 && !e.SavingThrow(t, SaveVsSpell, 0) {
		e.killWithMessage(t, StatusGone, "is dispelled")
		e.RemoveAffect(c, affDispellingEvil, nil)
		e.RemoveAffect(c, affDispelEvil, nil)
		return
	}
	e.CombatSay(t, "resists dispel evil", true)
}

// confused = O0A4:0DED: each round the confused run, stand dazed, go
// berserk or rage; a saving throw may end it.
func (e *RuleState) confused(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		return
	}
	switch r := e.Dice(1, 100); {
	case r <= 10:
		e.RemoveAffect(c, affConfused, nil)
		startFleeing(c)
		e.AddAffectMsg(c, affFrightened, sotsbRunAwayTime, 0, 1, 1, false, "runs away")
	case r <= 60:
		e.hitMessage(c, true, "is confused")
		e.ClearCombatInfo(c)
	case r <= 80:
		e.AddAffectMsg(c, affBerserk, sotsbBerserkTime, c.Side, 1, 1, false, "goes berserk")
		e.callAffect(affBerserk, c, nil, 0)
	default:
		e.hitMessage(c, true, "is enraged")
	}
	if e.SavingThrow(c, SaveVsSpell, -2) {
		e.RemoveAffect(c, affConfused, nil)
	}
}

// blinking = O0A4:0FD2: blinked away before its initiative came, it cannot be
// hit (POR: after it came).
func (e *RuleState) blinking(c *Character, a *Affect, flag byte) {
	if flag == 0 && combatInfo(c).Delay == 0 {
		e.AffectVeto = 1
		e.Rv.HitRoll = -1
	}
}

// dwarfDodges = O0A4:159E.
func (e *RuleState) dwarfDodges(c *Character, a *Affect, flag byte) {
	if flag == 0 && e.Acting().CreatureFlags&creatureGiantKin != 0 {
		e.Rv.HitRoll -= 4
	}
}

// gnomeDodges = O0A4:15C3.
func (e *RuleState) gnomeDodges(c *Character, a *Affect, flag byte) {
	if flag == 0 && e.Acting().CreatureFlags&creatureGnomeFoe != 0 {
		e.Rv.HitRoll -= 4
	}
}

// fireShieldHot = O0A4:163D: cold is resisted, unsaved fire hurts double.
func (e *RuleState) fireShieldHot(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		return
	}
	if e.Rv.DmgType&dmgCold != 0 {
		e.Rv.Save += 2
		return
	}
	if e.Rv.DmgType&DmgFire != 0 && !e.Rv.SaveMade {
		e.Rv.Damage *= 2
	}
}

// fireShieldCold = O0A4:1678: fire is resisted, cold hurts double. The
// original tests "nothing happened since the game was saved" (b7523) where
// its twin tests the saving throw; mid-fight that is false, so cold always
// doubles here.
func (e *RuleState) fireShieldCold(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		return
	}
	if e.Rv.DmgType&DmgFire != 0 {
		e.Rv.Save += 2
		return
	}
	if e.Rv.DmgType&dmgCold != 0 {
		e.Rv.Damage *= 2
	}
}

// globe = O0A4:1708.
func (e *RuleState) globe(c *Character, a *Affect, flag byte) { e.stopSpellsUpTo(flag, sotsbGlobeMax) }

// minorGlobe = O0A4:1A2C.
func (e *RuleState) minorGlobe(c *Character, a *Affect, flag byte) {
	e.stopSpellsUpTo(flag, sotsbMinorGlobeMax)
}

// needsPlus3WeaponSotsb = O0A4:1930.
func (e *RuleState) needsPlus3WeaponSotsb(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		return
	}
	if w := e.readiedWeapon(e.Acting()); w != nil && int8(w.Plus) >= sotsbPlusNeeded {
		return
	}
	e.Rv.Damage = 0
}

// stoningTouch = O0A4:1B79.
func (e *RuleState) stoningTouch(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		return
	}
	t := targetOf(c)
	if !e.SavingThrow(t, saveVsPetrify, 0) {
		e.killWithMessage(t, StatusStoned, "is turned to stone")
	}
}

// feeblemindedSotsb = O0A4:21BC.
func (e *RuleState) feeblemindedSotsb(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		return
	}
	c.Stats[StatInt] = sotsbFeebleScore
	c.Stats[StatWis] = sotsbFeebleScore
	e.CombatSay(c, "is made stupid", true)
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

// invisibleToAnimalsSotsb = O0A4:2284. The original looks at the invisible
// one's own kind, not the attacker's; kept.
func (e *RuleState) invisibleToAnimalsSotsb(c *Character, a *Affect, flag byte) {
	if flag == 0 && c.CreatureFlags&creatureAnimalsSeen != 0 && e.FindAffect(c, affDetectInvisible) == nil {
		e.unseenBy()
	}
}

// highlighted = O0A4:240F: easier to hit.
func (e *RuleState) highlighted(c *Character, a *Affect, flag byte) {
	if flag == 0 {
		c.AC[0] -= 2
		c.AC[1] -= 2
	}
}

// dragonBreathImmune = O0A4:2444.
func (e *RuleState) dragonBreathImmune(c *Character, a *Affect, flag byte) {
	if flag == 0 && e.Rv.DmgType&dmgBreath != 0 {
		e.blockAffect(affAny)
	}
}

// magicalBootsSotsb = O0A4:2465: harder to hit.
func (e *RuleState) magicalBootsSotsb(c *Character, a *Affect, flag byte) {
	if flag == 0 {
		c.AC[0] += 2
		c.AC[1] += 2
	}
}

// baneWeapon = O0A4:2497: 6d6 and the strength bonus against its foes.
func (e *RuleState) baneWeapon(c *Character, a *Affect, flag byte) {
	if flag == 0 && targetOf(c).MonsterFlags&monsterBane != 0 {
		e.Rv.Damage = int16(e.RollDice(6, 6) + int(strDamage(c)))
		e.Rv.HitRoll++
	}
}

// berserk = O0A4:250D: attacks the nearest one, of either side.
func (e *RuleState) berserk(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		if c.Control == controlCharmed {
			c.Control = 0
			c.Auto = 0
		}
		c.Side = SideParty
		return
	}
	e.goBerserk(c)
	if e.inCombat() {
		combatInfo(c).CanCast = 0
		c.Side = targetOf(c).Side ^ 1
		e.CombatSay(c, "goes berserk", true)
	}
}

// magicalStone = O0A4:2674.
func (e *RuleState) magicalStone(c *Character, a *Affect, flag byte) {
	if flag == 0 {
		e.Rv.Save++
	}
}

// fireHeals = O0A4:2689: magical fire heals, electricity slows, nothing
// else gets through (whatever the flag, as the original).
func (e *RuleState) fireHeals(c *Character, a *Affect, flag byte) {
	switch {
	case e.Rv.DmgType&DmgFire != 0 && e.Rv.DmgType&DmgMagic != 0:
		e.CombatSay(c, "is healed", true)
		c.HP = byte(min(int(c.HP)+int(e.Rv.Damage), int(c.MaxHP)))
	case e.Rv.DmgType&dmgElectric != 0:
		e.AddAffect(c, AffSlow, sotsbSlowTime, sotsbAlwaysData, 0)
		e.CombatSay(c, "is slowed", true)
	}
	e.blockAffect(affAny)
}

// fearAuraSotsb = O0A4:2A36: foes below the level who are not fearless run in terror.
func (e *RuleState) fearAuraSotsb(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		return
	}
	for _, foe := range e.Party {
		if foe.Side == c.Side || e.FindAffect(foe, affFearless) != nil ||
			e.FindAffect(foe, affFrightened) != nil || int8(foe.MaxLevel) >= sotsbFearLevel {
			continue
		}
		e.AddAffectMsg(foe, affFrightened, 0, sotsbAlwaysData, 1, 1, false, "is terrified")
		if e.FindAffect(foe, affFrightened) != nil {
			startFleeing(foe)
		}
	}
}

// boltsHarmlessSotsb = O0A4:2BE7.
func (e *RuleState) boltsHarmlessSotsb(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		return
	}
	if w := e.readiedWeapon(e.Acting()); w != nil && e.Base(w.Type).Usage&sotsbItemUsesBolts != 0 {
		e.Rv.Damage = 0
	}
}

// fearless = O0A4:2EB6.
func (e *RuleState) fearless(c *Character, a *Affect, flag byte) {
	if flag == 0 {
		e.blockAffect(affFrightened)
	}
}

// hardyAll = O0A4:2EEF: against paralysis and poison, rods and spells.
func (e *RuleState) hardyAll(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		return
	}
	if t := e.Rv.SaveType; t == SaveVsSpell || t == saveVsWand || t == SaveVsParalysis {
		e.Rv.Save += constitutionSaveBonus(c)
	}
}

// plainWeaponsImmuneSotsb = O0A4:3553: a weapon without a plus does nothing in
// the hands of a player race or a low-level monster.
func (e *RuleState) plainWeaponsImmuneSotsb(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		return
	}
	sel := e.Acting()
	if w := e.readiedWeapon(sel); w != nil && w.Plus != 0 {
		return
	}
	if sel.Race != raceMonster && sel.Race <= RaceHuman || int8(sel.MaxLevel) < sotsbPlainLevel {
		e.Rv.Damage = 0
	}
}

// berserkFoe = O0A4:3637: attacks the nearest; afterwards back to the
// side its data keeps.
func (e *RuleState) berserkFoe(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		if c.Control == controlCharmed {
			c.Control = 0
		}
		c.Side = a.Data
		return
	}
	e.goBerserk(c)
	c.Side = targetOf(c).Side ^ 1
}

// alwaysInvisibleSotsb = O0A4:3757.
func (e *RuleState) alwaysInvisibleSotsb(c *Character, a *Affect, flag byte) {
	if flag == 0 {
		e.AddAffect(c, affInvisible, sotsbAlwaysTime, sotsbAlwaysData, 0)
	}
}

// frightened = O0A4:37D5: when the fear goes, so does the flight.
func (e *RuleState) frightened(c *Character, a *Affect, flag byte) {
	if flag == 0 {
		return
	}
	if c.Control == controlCharmed {
		c.Control = 0
		c.Auto = 0
	}
	combatInfo(c).Turned = 0
}

// zapsAttackers = O0A4:3826: a close attacker takes double the damage it
// dealt, as magic.
func (e *RuleState) zapsAttackers(c *Character, a *Affect, flag byte) {
	sel := e.Acting()
	if flag != 0 || e.RangeTo(sel, c) >= 2 {
		return
	}
	dealt := e.Rv.Damage
	e.Rv.Damage *= 2
	if e.Rv.Damage == 0 {
		return
	}
	dmgType := e.Rv.DmgType
	e.Rv.DmgType = DmgMagic
	e.CombatSay(sel, "gets zapped", true)
	e.Damage(sel, int(e.Rv.Damage), 0, false)
	e.Rv.Damage, e.Rv.DmgType = dealt, dmgType
}

// giantStrengthSpell = O0A4:38BD: the spell's message; the strength comes
// from the game's statAffects.
func (e *RuleState) giantStrengthSpell(c *Character, a *Affect, flag byte) {
	if flag == 0 {
		e.CombatSay(c, "gains giant strength", true)
	}
}

// breathFireCloseSotsb = O0A4:2094: once the fight is old enough (its first
// attack count no less than the round), fire at the foe close by: 8 points,
// half on a save.
func (e *RuleState) breathFireCloseSotsb(c *Character, a *Affect, flag byte) {
	if flag != 0 || e.CombatRound == 0 || byte(c.BaseAttack[0]) < e.CombatRound {
		return
	}
	e.Rv.DmgType = DmgFire | dmgBreath
	t := targetOf(c)
	if e.RangeTo(c, t) >= 2 {
		return
	}
	e.CombatSay(c, "Breathes Fire", true)
	e.shoot(e.CombatX(c), e.CombatY(c), e.CombatX(t), e.CombatY(t), 0, dnd.SpellBreathWeapon)
	e.Damage(t, sotsbCloseBreath, saveHalf, e.SavingThrow(t, saveVsBreath, 0))
}

// Swallowing (sotsb O0A4:1C26).
const (
	swallowRoll = 75 // the attack roll that swallows
	belchChance = 50 // percent: the swallower belches, else the victim is digested
)

// swallowsWholeSotsb = sotsb O0A4:1C26: a very good roll swallows the
// target; its body lies where the swallower stands. Its flicker is not drawn.
func (e *RuleState) swallowsWholeSotsb(c *Character, a *Affect, flag byte) {
	if flag != 0 || e.Rv.HitRoll < swallowRoll {
		return
	}
	t := targetOf(c)
	if t == nil {
		return
	}
	e.CombatSay(t, "is swallowed whole", true)
	t.Status, t.Okay, t.HP = StatusDead, 0, 0
	e.removeCombatAffects(t)
	e.AffectGroup(t, grpDown)
	e.CombatSay(t, "is killed", false)
	if e.Dice(1, 100) <= belchChance {
		e.CombatSay(c, "belches in satisfaction", true)
	} else {
		e.CombatSay(t, "is digested instantly", true)
	}
	e.combatRemoveUnder(t, c)
}
