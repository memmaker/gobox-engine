package rules1e

import (
	"gobox/core"
	"gobox/core/dnd"
)

// Monster attacks that aim with a spell's targeting (fold T4d, M1): the
// later games' gazes, spits and breaths pick their target with the target
// hook as a spell of the game's would, then fly at it.

// The spells whose targeting the attacks borrow (the unified ids of the
// later games' 10, 23, 51, 61 and 68).
const (
	aimCharm     = spCharmPerson   // a chosen foe in sight
	aimHold      = SpHoldPersonCl  // a chosen square (PoD's cone breaths)
	aimBolt      = SpLightningBolt // a square for a bolt
	aimGaze      = SpellID(61)     // POR's power 61: a chosen foe (gazes, spits, SotSB's cone breaths)
	aimTouch     = spPoison        // a foe close by (breath of fire, confusing gaze)
	gazeReachMax = 2               // a confusing gaze reaches less than this
)

// aimWithSpell = the later games' DS target hook with spell's targeting: the
// one the attack goes for, nil when there is none.
func (e *RuleState) aimWithSpell(spell SpellID) *Character {
	if !e.spellTargetHook(int(spell), 1, CastReply{Fight: true}) {
		return nil
	}
	return e.spellTarget
}

// shootAt: the attack flies from one combatant to another.
func (e *RuleState) shootAt(from, to *Character, look core.SpellID) {
	e.shoot(e.CombatX(from), e.CombatY(from), e.CombatX(to), e.CombatY(to), 0, look)
}

// faceAndShoot: c turns to the first target and its attack flies to the
// target square.
func (e *RuleState) faceAndShoot(c *Character, look core.SpellID) {
	e.TurnTo(c, byte(e.directionTo(c, e.spellTarget)))
	e.shoot(e.CombatX(c), e.CombatY(c), e.TgtX, e.TgtY, 0, look)
}

// foeGaze = PoD's O07C gazeAt: the gazer picks a foe with the spell's
// targeting; nil when there is no foe to gaze at.
func (e *RuleState) foeGaze(c *Character, spell SpellID) *Character {
	t := e.aimWithSpell(spell)
	if t == nil || t.Side == c.Side {
		return nil
	}
	return t
}

// reflectGaze = PoD's O07C reflectGaze: the gaze flies at t and, when the
// gazer's gaze is reflectable and t holds a mirror, back at the gazer: the
// one the gaze ends on.
func (e *RuleState) reflectGaze(c, t *Character) *Character {
	e.CombatSay(c, "gazes...", true)
	e.shootAt(c, t, dnd.SpellStoneGaze)
	if e.FindAffect(c, affReflectableGaze) != nil && e.FindAffect(t, affMirror) != nil {
		e.CombatSay(t, "reflects it!", true)
		e.shootAt(t, c, dnd.SpellStoneGaze)
		return c
	}
	return t
}

// gazeImmune: the gaze immunity some carry (PoD's affect 126).
func (e *RuleState) gazeImmune(t *Character) bool { return e.FindAffect(t, affGazeImmunePod) != nil }

// stoningGaze = PoD's O07C:134E: the foe it looks at is turned to stone
// unless it saves or is immune to gazes.
func (e *RuleState) stoningGaze(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		return
	}
	t := e.foeGaze(c, aimGaze)
	if t == nil {
		return
	}
	t = e.reflectGaze(c, t)
	if !e.SavingThrow(t, saveVsPetrify, 0) && !e.gazeImmune(t) {
		e.killWithMessage(t, StatusStoned, "is turned to stone")
	}
}

// The hold a paralysing gaze gives.
const (
	gazeHoldTime  = 10
	gazeHoldData  = 12
	gazeHoldSave  = 3 // its saving throw's bonus
	gazeHoldCall  = 0
	gazeHoldMode  = 1
	gazeSaveBonus = 0
)

// paralysingGaze = PoD's O07C:2C72: the foe it looks at is held unless it
// saves (at +3) or is immune to gazes.
func (e *RuleState) paralysingGaze(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		return
	}
	t := e.foeGaze(c, aimCharm)
	if t == nil {
		return
	}
	e.Rv.DmgType = dmgAcid // the later games' gaze bit
	t = e.reflectGaze(c, t)
	saved := e.SavingThrow(t, saveVsPetrify, gazeHoldSave) || e.gazeImmune(t)
	e.AddAffectMsg(t, AffHeld, gazeHoldTime, gazeHoldData, gazeHoldCall, gazeHoldMode, saved, "paralyzed")
}

// confusingGaze = SotSB's O0A4 and PoD's O07C:1A05 (PoD's superset: the
// gaze bit and gaze immunity): a foe close by is confused for a few rounds
// unless it saves.
func (e *RuleState) confusingGaze(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		return
	}
	t := e.aimWithSpell(aimTouch)
	if t == nil || e.RangeTo(c, t) >= gazeReachMax {
		return
	}
	e.Rv.DmgType = dmgAcid
	e.CombatSay(c, "gazes...", true)
	e.shootAt(c, t, dnd.SpellStoneGaze)
	if !e.SavingThrow(t, SaveVsSpell, gazeSaveBonus) && !e.gazeImmune(t) {
		e.AddAffectMsg(t, affConfused, e.Dice(1, 10)+2, c.Side, 1, 1, false, "stares blankly")
	}
}

// charmGazeLater = PoD's O0BB:6B64: the gaze casts Charm Person on the
// chosen target (POR's CharmGaze: a fixed reach and its own saving throw).
func (e *RuleState) charmGazeLater(c *Character, a *Affect, flag byte) {
	if e.aimWithSpell(aimCharm) == nil {
		return
	}
	e.areaSpell = 1
	e.Rv.DmgType = dmgAcid
	e.CombatSay(c, "gazes", true)
	e.faceAndShoot(c, dnd.SpellCharmGaze)
	e.CurSpell = spCharmPerson
	e.spellEffect(spCharmPerson)
	e.areaSpell = 0
}

// rakshasaLevel: spells of a lower level do not get through a rakshasa's
// protection.
const rakshasaLevel = 8

// rakshasaProtection = PoD's O07C:0ED5.
func (e *RuleState) rakshasaProtection(c *Character, a *Affect, flag byte) {
	if flag == 0 && e.CurSpell > 0 && e.T().Spells[e.CurSpell][sotsbSpellLevelField] < rakshasaLevel {
		e.blockAffect(affAny)
	}
}

// breathFireNear is the fire a close breath of fire does.
const breathFireNear = 7

// breathFire = SotSB's O0A4 breathFireHandler: fire at a foe close by, half
// on a saving throw against breath.
func (e *RuleState) breathFire(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		return
	}
	e.Rv.DmgType = DmgFire | dmgBreath
	t := e.aimWithSpell(aimTouch)
	if t == nil {
		return
	}
	e.TurnTo(c, byte(e.directionTo(c, t)))
	e.CombatSay(c, "breathes fire", true)
	e.shootAt(c, t, dnd.SpellBreathWeapon)
	e.Damage(t, breathFireNear, saveHalf, e.SavingThrow(t, saveVsBreath, 0))
}

// SotSB's spit: its reach, and the chance (in tens of percent) of the first round.
const (
	spitReachSotsb      = 11
	spitFirstRoundSotsb = 1
	spitChanceStep      = 10
)

// spitAcidSotsb = SotSB's O0A4 spitAcidHandler: the nearer the foe, the
// likelier the spit (one in ten in the first round); 4d8 acid, half on a
// saving throw against breath. A variant of POR's SpitAcid (25%, 8d4,
// reach 4), CotAB has none.
func (e *RuleState) spitAcidSotsb(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		return
	}
	e.Rv.DmgType = dmgAcid
	t := e.aimWithSpell(aimGaze)
	if t == nil {
		return
	}
	chance := 0
	if r := e.RangeTo(c, t); r <= spitReachSotsb {
		chance = (spitReachSotsb - r) * spitChanceStep
		if e.CombatRound == 0 {
			chance = spitFirstRoundSotsb * spitChanceStep
		}
	}
	if chance <= 0 {
		return
	}
	if e.Dice(1, 100) <= chance {
		e.CombatSay(c, "spits...", true)
		e.shootAt(c, t, dnd.SpellSpitAcid)
		e.Damage(t, e.RollDice(4, 8), saveHalf, e.SavingThrow(t, saveVsBreath, 0))
		return
	}
	e.CombatSay(c, "spits and misses", true)
}

// breathsPerFight: a dragon breathes three times a fight.
const breathsPerFight = 3

// breathCharges: the breaths left, three again at the fight's start;
// false when none is left.
func (e *RuleState) breathCharges(a *Affect) bool {
	if e.CombatRound == 0 || a.Data > breathsPerFight {
		a.Data = breathsPerFight
	}
	return a.Data > 0
}

// breathCone is one cone breath: its shape, what it is and its damage
// (nil: the breather's maximum hit points).
type breathCone struct {
	aim           SpellID
	width, length int
	dmgType       byte
	what          string
	damage        func(e *RuleState, c *Character) int
	anyone        bool // the cone may catch friends (PoD's poison breath), else only foes
}

func maxHPBreath(e *RuleState, c *Character) int { return int(c.MaxHP) }

// The cone breaths (SotSB's O0A4 breathWeaponHandler, PoD's O0BB:6874
// breathHook and O0BB:652D poisonBreathHook).
var (
	coldCone      = breathCone{aimGaze, 3, 7, dmgCold | dmgBreath, "cold", maxHPBreath, false}
	fireCone      = breathCone{aimGaze, 3, 9, DmgFire | dmgBreath, "fire", maxHPBreath, false}
	coldConePod   = breathCone{aimHold, 3, 7, dmgCold | dmgBreath, "cold", maxHPBreath, false}
	fireConePod   = breathCone{aimHold, 3, 9, DmgFire | dmgBreath, "fire", maxHPBreath, false}
	acidConePod   = breathCone{aimHold, 1, 6, dmgBreath, "acid", maxHPBreath, false}
	poisonConePod = breathCone{aimGaze, 255, 3, dmgBreath, "poison", maxHPBreath, true}
	bigFireCone   = breathCone{aimHold, 3, 9, DmgFire | dmgBreath, "fire", func(e *RuleState, c *Character) int {
		return e.RollDice(bigFireDice, 10) + bigFireBonus
	}, false}
)

// The big fire breath: 24d10+12.
const (
	bigFireDice  = 24
	bigFireBonus = 12
)

// coneBreath: three times a fight, a cone at the chosen square, as long as
// it catches only foes; each caught takes the damage, half on a saving
// throw against breath; the turn is over.
func (e *RuleState) coneBreath(c *Character, a *Affect, b breathCone) {
	e.areaSpell = 1
	defer func() { e.areaSpell = 0 }()
	if !e.breathCharges(a) {
		return
	}
	e.Rv.DmgType = b.dmgType
	x, y := e.CombatX(c), e.CombatY(c)
	if e.aimWithSpell(b.aim) == nil {
		return
	}
	e.coneTargets(x, y, e.TgtX, e.TgtY, b.width, b.length)
	for i := 1; i <= e.NSpellTargets && !b.anyone; i++ {
		if t := e.tgt(i); t != nil && t.Side == c.Side {
			return
		}
	}
	if e.NSpellTargets == 0 {
		return
	}
	e.CombatSay(c, "breathes "+b.what, true)
	e.TurnTo(c, byte(e.directionTo(c, e.spellTarget)))
	e.shootAt(c, e.spellTarget, dnd.SpellBreathWeapon)
	dmg := b.damage(e, c)
	for i := 1; i <= e.NSpellTargets; i++ {
		if t := e.tgt(i); t != nil {
			e.Damage(t, dmg, saveHalf, e.SavingThrow(t, saveVsBreath, 0))
		}
	}
	a.Data--
	e.ClearCombatInfo(c)
}

// lightningReachPod is the reach of PoD's lightning breath.
const lightningReachPod = 9

// lightningBreathPod = PoD's O0BB:6728: three times a fight, a bolt that
// does the breather's maximum hit points.
func (e *RuleState) lightningBreathPod(c *Character, a *Affect, flag byte) {
	if !e.breathCharges(a) {
		return
	}
	e.Rv.DmgType = dmgElectric | dmgBreath
	if e.aimWithSpell(aimBolt) == nil {
		return
	}
	e.CombatSay(c, "breathes lightning", true)
	e.faceAndShoot(c, dnd.SpellBreathWeapon)
	e.boltPath(lightningReachPod, int(c.MaxHP), int(saveVsBreath), 0, dnd.SpellBreathWeapon)
	a.Data--
	e.ClearCombatInfo(c)
}

func coneHandler(b breathCone) AffectHandler {
	return func(e *RuleState, c *Character, a *Affect, flag byte) { e.coneBreath(c, a, b) }
}

// aimHandlers are the handlers above, by the names the games' affectVariants give.
var aimHandlers = map[HandlerName]AffectHandler{
	"stoningGazePod":        (*RuleState).stoningGaze,
	"paralysingGazePod":     (*RuleState).paralysingGaze,
	"confusingGaze":         (*RuleState).confusingGaze,
	"charmGazePod":          (*RuleState).charmGazeLater,
	"rakshasaProtectionPod": (*RuleState).rakshasaProtection,
	"breathFireSotsb":       (*RuleState).breathFire,
	"spitAcidSotsb":         (*RuleState).spitAcidSotsb,
	"breathWeaponSotsb":     coneHandler(coldCone),
	"breathWeapon2Sotsb":    coneHandler(fireCone),
	"whiteBreathPod":        coneHandler(coldConePod),
	"breathFireConeMaxHP":   coneHandler(fireConePod),
	"breathFireCone24d10":   coneHandler(bigFireCone),
	"blackBreathPod":        coneHandler(acidConePod),
	"greenBreathPod":        coneHandler(poisonConePod),
	"blueBreathPod":         (*RuleState).lightningBreathPod,
}
