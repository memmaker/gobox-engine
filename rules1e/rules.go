package rules1e

// Rules: affects, to-hit rolls, saving throws, damage and healing, ported
// from ovr25 and the START helpers it leans on (00BA:16EC ... 2637).

import (
	"gobox/rules"
	"slices"
	"strconv"
)

// CombatInfo is the 0x16-byte block at char+0x108 (allocated by ovr11).
type CombatInfo struct {
	Spell      SpellID    // +0x00 spell being cast (lost when hit)
	CanCast    byte       // +0x01
	CanUse     byte       // +0x02 may use an item
	Delay      byte       // +0x03 signed initiative; 0x14 = delayed
	AttackSlot byte       // +0x04 attack slot in use (1, 2)
	Sweeps     byte       // +0x05 sweeps left against weak foes
	Moves      byte       // +0x06 half squares left
	Guarding   byte       // +0x07
	Attacked   byte       // +0x08 has attacked this turn
	Facing     byte       // +0x09 0..7
	Target     *Character // +0x0A
	Dying      byte       // +0x0E rounds of bleeding left
	Attackers  byte       // +0x0F times attacked this round
	Turned     byte       // +0x10 turned by a cleric
	TurnUsed   byte       // +0x11 has tried to turn undead
	Flank      byte       // +0x12 accumulated turn from attackers
	Extra      byte       // +0x13 not one of the saved party
	Fleeing    byte       // +0x14
	Pattern    byte       // +0x15 AI move pattern, row of DS:02AE
	Ranged     bool       // extra: the computer chose its Launcher or Missiles slot for this turn's attacks (AiPickWeapon)
}

// Rule globals (DS:6816..682A).
type ruleVars struct {
	// Save DS:6816 saving throw total, DiceN DS:681C, HitRoll DS:6822, SaveType DS:682A.
	rules.Roll
	AffVal  AffectType // DS:6817 affect being added (0 = blocked)
	Damage  int16      // DS:6818 (a signed word, as PoD has it; POR and the others kept a byte)
	DmgType byte       // DS:6819 1 fire 2 cold 4 electricity 8 magic 0x10 acid
	Cure    byte       // DS:681F set while the temple removes affects
	Side    [2]int8    // DS:6814 standing members per side
	// Critical (PoD A7E9): the last attack roll was above criticalAbove and
	// hit (a natural 1 leaves it as it was). Only PoD's vorpal blade reads it.
	Critical bool
	// extra: the attack under way (attackSlots, Strike): who strikes and what
	// with. The affects and the rule set's hooks called during it read them
	// (WieldOf); outside an attack By is nil.
	By   *Character
	With Wield
}

// criticalAbove: an attack roll above it that hits is critical (PoD).
const criticalAbove = 16

// AffectHandler is one entry of the DS:6828 table (filled by ovr13:31A2):
// handler(char, affect, flag); flag 1 = the affect is being removed.
type AffectHandler func(e *RuleState, c *Character, a *Affect, flag byte)

// callAffect = ovr25:0000 (type, char, affect, flag): the handler e's tables
// name for the id (Tables.Affects, with the game's affectVariants).
func (e *RuleState) callAffect(typ AffectType, c *Character, a *Affect, flag byte) {
	if h := affectHandler[e.T().Affects[typ]]; h != nil {
		h(e, c, a, flag)
	}
}

// CallAffect is callAffect for the games' parity tests of their handlers.
func (e *RuleState) CallAffect(typ AffectType, c *Character, a *Affect, flag byte) {
	e.callAffect(typ, c, a, flag)
}

// FindAffect = 00BA:2199 (out, type, char): first affect of the type.
func (e *RuleState) FindAffect(c *Character, typ AffectType) *Affect {
	for _, a := range c.Affects {
		if a.Type == typ {
			return a
		}
	}
	return nil
}

// RemoveAffect = ovr25:0026 (char, type, affect): affect nil means the first
// one of the type. Its handler runs first when the node asks for it.
func (e *RuleState) RemoveAffect(c *Character, typ AffectType, a *Affect) {
	if a == nil {
		a = e.FindAffect(c, typ)
	}
	if a == nil {
		return
	}
	if a.Call != 0 {
		e.callAffect(typ, c, a, 1)
	}
	for i, x := range c.Affects {
		if x == a {
			c.Affects = append(c.Affects[:i], c.Affects[i+1:]...)
			return
		}
	}
}

// AddAffect = ovr25:0E52 (char, type, time, data, call): append a node.
func (e *RuleState) AddAffect(c *Character, typ AffectType, time int, data, call byte) {
	a := &Affect{AffectRecord: AffectRecord{Type: typ, Time: uint16(time), Data: data, Call: call}}
	c.Affects = append(c.Affects, a)
	e.Emit(rules.Sounded{Cue: rules.SoundCue(daspAffects[typ])})
}

// auraAffects = ovr25 cs:0140: affects that also work on everyone near the
// one who has them (prayer, chant ...).
func auraAffect(t AffectType) bool {
	return t == affSilenced || t == affProtEvilRadius || t == affProtGoodRadius || t == affPrayer
}

// checkAffect = ovr25:0160 (char, type): run the handler (flag 0) when the
// character has the affect, or for an aura affect when anyone in range
// (anyone at all outside combat) has it.
func (e *RuleState) checkAffect(c *Character, typ AffectType) {
	a := e.FindAffect(c, typ)
	found := a != nil
	if !found && auraAffect(typ) {
		saved := e.nearList // DS:671B, 0xD8 bytes
		for _, p := range e.Party {
			if found {
				break
			}
			pa := e.FindAffect(p, typ)
			if pa == nil {
				continue
			}
			a = pa
			if e.BeginFlag != 5 {
				found = true
				continue
			}
			r := 1
			if typ == affPrayer {
				r = 6
			}
			e.nearCombatants(p, e.CombatX(p), e.CombatY(p), r)
			me := e.CombatIndex(c)
			for _, n := range e.Near {
				if int(n.Index) == me {
					found = true
				}
			}
		}
		e.nearList = saved
	}
	if found {
		e.callAffect(typ, c, a, 0)
	}
}

// affectGroups = the lists of ovr25:02FA: which affects take part in each
// kind of check (the group number is the caller's).
var affectGroups = [][]AffectType{
	0x00: {affGazeReflected},
	0x01: {affMinorGlobe, affBlink, affInvisible, AffMonsterInvisible},
	0x02: {affWeakeningTouch, affWeakeningTouch2, affDisease, affParalyseNotElf, affMagicDamage2d10, affAcidDamage, affHoldTarget, affRake, affHoldWeakTarget},
	0x03: {affPoisonBite, affPoisonBite4, affPoisonBite2, affParalyse2d8, affParalyseNotElf, affParalyse1d9, affPoisonBiteMinus2, affMagicDamage2d10, affWeakeningTouch2, affDisease},
	0x04: {affEnfeebled, affBonusVsRace4, AffBonusByTargetRace, AffRangerVsGiants},
	0x05: {affMirrorImage, affProtMissiles, affAvoidDistant, affDodgeBoulders, affStartRegenerating, affHalveUnlessPlain, affHalveVsMagicWeapon, affOnlyMagicOrStrong, affHalveOrCancelUnmagic, affOnlyMagicWeapons, affHalveVsWeaponFlags, affResistNormalWeapons, affOilDamage, affHolyWaterDamage},
	0x06: {affResistFireLessPerDie, AffFireResistance, affOilDamage, affResistNormalWeapons, affResistFireElectric, affResistCold, affResistFire, affResist50, affResist15, affBlockFire, affHalveElectric, affHalveCold, affShield, affHalveFire, affStartRegenerating, affMirrorImage},
	0x07: {affSnakeCharm, AffHeld, affAsleep, affHelpless},
	0x08: {affStandUpBleeding, affFear, affFearLifted, affDisplaced, affCamouflaged, affTurnInvisible},
	0x09: {affResist50, affResist15, affBlock90, affBlockSleepCharm, affBlockHold, affBlockCold, affBlockPoisonHold, affBlockFire, affBlock30, affBlockMind},
	0x0A: {AffBless, AffCurse, AffBlind, AffBestowCurse, affPrayer, affBonusVsRace4, AffBonusByTargetRace, affHitBonusVsNames, affHitBonusVsTargetNames},
	0x0B: {AffBlind, affShield, AffProtEvil, affProtGood, affProtEvilRadius, affProtGoodRadius, affCoughing},
	0x0C: {AffProtEvil, affProtGood, affResistCold, affShield, affResistFire, AffBlind, AffBestowCurse, affProtEvilRadius, affProtGoodRadius, affPrayer, AffFireResistance, affBlockPoisonHold, affBlockMind, affSaveBonusVsType0, affSaveBonusVsType42, affBarkskin},
	0x0D: {affStandUpBleeding, affRegenUnlessFireAcid, affGaseousEscape, affBloodDrain, affStrikeTied},
	0x0E: {affStoneGaze, affCharmGaze, 0x58, affSpitAcid},
	0x0F: {affSilenced, affCoughing, affStrikeTied, affBloodDrain, AffCharmed},
	0x10: {affInvisible, AffMonsterInvisible, affBlink, affMinus4VsNames, affMinus4VsBugbearGnoll, affDisplaced},
	0x11: {AffBless, AffCurse, AffCharmed},
	0x12: {AffHaste, AffSlow, affHeldFast},
	0x13: {affHeal3, affSpiritualHammer, affCamouflaged, affTurnInvisible, AffCharmed},
}

// Affect check groups by meaning (the ovr25:02FA selector; the later games'
// selectors O0E3:0414, O145:0466, O0E2:04DD number 0..0x15 the same). The
// groups past POR's 0x13 are rules1e's numbers: the games gave 0x16 different
// meanings.
const (
	grpGazeReflected   = 0x00 // POR's only: an empty group in the later games
	grpAttackerExtra   = 0x01
	grpOnHit           = 0x02
	grpOnDamage        = 0x03
	grpDamageBonus     = 0x04
	grpWeaponDamage    = 0x05
	grpDamage          = 0x06
	grpTurn            = 0x07
	grpRound           = 0x08
	grpAddAffect       = 0x09
	grpAttacker        = 0x0A
	grpArmour          = 0x0B
	grpSave            = 0x0C
	grpDown            = 0x0D
	grpAI              = 0x0E
	grpUse             = 0x0F
	grpDefender        = 0x10
	grpMorale          = 0x11
	grpSpeed           = 0x12
	grpRest            = 0x13
	grpUnsavedDamage   = 0x14 // damage not saved against (fire shields)
	grpConfusion       = 0x15 // never asked for by any game
	grpMovement        = 0x16 // SotSB's 0x16; never asked for
	grpAlwaysInvisible = 0x17 // COTAB's 0x16: at the fight's start
	grpCombatStart     = 0x18 // PoD's 0x18: at the fight's start
	grpUnused          = 0x19 // COTAB's 0x17: never asked for
)

// groups is the game's affect groups (Tables.Groups, rules.json
// affectGroups); POR's when the game names none.
func (e *RuleState) groups() [][]AffectType {
	if g := e.T().Groups; g != nil {
		return g
	}
	return affectGroups
}

// AffectGroup = ovr25:02FA (char, group).
func (e *RuleState) AffectGroup(c *Character, group int) {
	r := e.T().Resist
	if r != nil && (group == grpDamage || group == grpAddAffect) {
		e.magicResistance(c, r)
	}
	if g := e.groups(); group < len(g) {
		for _, t := range g[group] {
			e.checkAffect(c, t)
		}
	}
	if r != nil && r.KindImmunity && group == grpDamage {
		e.kindImmunity(c)
	}
}

// MagicResistRule is a later game's own magic resistance (rules.json
// magicResistance), run before the damage and add-affect groups (SotSB
// O145:01E1, PoD O0E2:0406); nil: none (POR, CotAB: their affects resist).
type MagicResistRule struct {
	Keep         [2]AffectType // stopped only as damage: the affects at the original's numbers 91 and 82
	DamageGate   bool          // SotSB: plain (unmagical) damage is never resisted, even with an affect
	KindImmunity bool          // PoD O0E2:0487: after the damage group, the poison immune and breathless take no poison / type 64
}

var (
	SotsbMagicResist = MagicResistRule{Keep: [2]AffectType{affCloudkill, affFear}, DamageGate: true}
	PodMagicResist   = MagicResistRule{Keep: [2]AffectType{affBlueBreathPod, affSlainByBoltsPod}, KindImmunity: true}
)

// magicResistance: the character's resistance, 5% better for each caster
// level under 11, stops the affect being added and the damage.
func (e *RuleState) magicResistance(c *Character, r *MagicResistRule) {
	if c.MagicResistance == 0 {
		return
	}
	if r.DamageGate {
		if e.Rv.AffVal == 0 || e.Rv.Damage != 0 && e.Rv.DmgType&DmgMagic == 0 {
			return
		}
	} else if e.Rv.AffVal == 0 && e.Rv.DmgType&DmgMagic == 0 {
		return
	}
	chance := int(c.MagicResistance) + (resistTopLevel-e.casterLevel(int(e.CurSpell)))*resistPerLevel
	if e.Dice(1, 100) <= chance {
		e.Rv.Damage = 0
		if e.Rv.AffVal != r.Keep[0] && e.Rv.AffVal != r.Keep[1] {
			e.Rv.AffVal = 0
		}
	}
}

// Monster kinds immune to a damage kind (PoD O0E2:0487).
const (
	dmgPoison                        = 0x80
	monsterPoisonImmune MonsterFlags = 16 // immune to poison (monsterNoBreath: to damage type 64)
)

// kindImmunity = PoD O0E2:0487: poison does nothing to the poison immune,
// damage type 64 nothing to those that do not breathe.
func (e *RuleState) kindImmunity(c *Character) {
	if e.Rv.DmgType&dmgPoison != 0 && c.MonsterFlags&monsterPoisonImmune != 0 ||
		e.Rv.DmgType&dmgSlaying != 0 && c.MonsterFlags&monsterNoBreath != 0 {
		e.Rv.Damage = 0
		e.Rv.AffVal = 0
	}
}

// StinkingCloud = ovr25:0AE7 (char): a character standing in a stinking
// cloud coughs (save) or is overcome by nausea.
func (e *RuleState) StinkingCloud(c *Character) {
	if c.Okay == 0 {
		return
	}
	_, ground := e.GroundAt(c, 8)
	if ground != e.grounds().StinkingCloud {
		return
	}
	for _, t := range []AffectType{affCoughing, affHelpless, affAnimated, affBlockPoisonHold, affBlockMind} {
		if e.FindAffect(c, t) != nil {
			return
		}
	}
	saved := e.SavingThrow(c, SaveVsParalysis, 0)
	var t AffectType
	if saved {
		t = affCoughing
		e.AddAffectMsg(c, t, 1, 0xFF, 0, 0, saved, "starts to cough")
	} else {
		t = affHelpless
		e.AddAffectMsg(c, t, e.Dice(1, 4)+1, 0xFF, 0, 0, saved, "chokes and gags from nausea")
	}
	// ovr25 passes the out pointer of its last FindAffect (0x7D), which is nil.
	e.callAffect(t, c, nil, 0)
}

// RollHit = ovr25:0C4F (char, bonus): d20 (a natural 20 counts 100) plus
// bonus must beat the defender's armour value at +0x111.
func (e *RuleState) RollHit(c *Character, bonus int) bool {
	if !e.attackDie() {
		return false
	}
	e.Rv.Critical = e.Rv.HitRoll > criticalAbove
	e.AffectGroup(c, grpDefender)
	hit := TrapHits(int(e.Rv.HitRoll), int(byte(bonus)), stats{c}.AC(0))
	e.Rv.Critical = e.Rv.Critical && hit
	return hit
}

// attackDie rolls the attack die into e.rv.HitRoll; false: a natural 1.
func (e *RuleState) attackDie() bool {
	rng := rules.RNG{State: e.RandSeed}
	die := AttackDie(&rng)
	e.RandSeed = rng.State
	e.Rv.HitRoll = int8(die)
	return !AttackMissed(die)
}

// AttackHits = ovr25:0CB7 (attacker, defender, need): attacking ends
// invisibility; d20 + THAC0 + the side's modifier must reach need. w is
// what att strikes with (its THAC0).
func (e *RuleState) AttackHits(att, def *Character, w Wield, need int) bool {
	adj := int(int8(byte(e.Game.PartyHit)))
	if att.Side != SideParty {
		adj = int(int8(byte(e.Game.MonsterHit)))
	}
	e.removeInvisibility(att)
	if !e.attackDie() {
		return false
	}
	e.Rv.Critical = e.Rv.HitRoll > criticalAbove
	e.AffectGroup(att, grpAttacker)
	e.AffectGroup(def, grpDefender)
	roll, thac0, need := int(e.Rv.HitRoll), int(e.thac0(att, w)), int(byte(need))
	hit := AttackHits(roll, thac0, adj, need) &&
		(e.Rv.HitRoll == 100 || e.Rules().Hits(e, att, def, roll+thac0+adj-need)) // extra: the rule set may turn the hit away
	e.Rv.Critical = e.Rv.Critical && hit
	return hit
}

// SavingThrow = ovr25:0D63 (char, type, bonus): 1 always fails, 20 always
// saves, otherwise d20 + bonus + the character's save bonus must reach the
// target for the type (0 paralysis ... 4 spells).
func (e *RuleState) SavingThrow(c *Character, typ, bonus int) bool {
	rng := rules.RNG{State: e.RandSeed}
	die := RollDice(&rng, 1, 20)
	e.RandSeed = rng.State
	e.Rv.Save = byte(die)
	if made, fixed := SaveFixed(die); fixed {
		e.Rv.SaveMade = made
		return made
	}
	e.Rv.Save = byte(die + int(int8(c.SaveBonus)) + int(int8(bonus)))
	e.Rv.SaveType = byte(typ)
	e.AffectGroup(c, grpSave)
	// signed: ruling 1, CHANGES-FROM-ORIGINAL.md
	e.Rv.SaveMade = SaveMade(int(c.Saves[int8(typ)]), int(int8(e.Rv.Save)))
	return e.Rv.SaveMade
}

// RollDice = ovr25:0E2E (n, sides): Dice that also remembers n (DS:681C).
func (e *RuleState) RollDice(n, sides int) int { return rules.RollCounted(&e.Rv.Roll, e, n, sides) }

// KillCharacter = ovr25:0F06 (char, status, message).
func (e *RuleState) KillCharacter(c *Character, status byte, msg string) {
	msg = rules.Cut(msg, msgMax)
	if c.Okay == 0 {
		return
	}
	idx := e.CombatIndex(c)
	e.CombatSay(c, msg, true)
	c.Okay = 0
	c.Status = status
	if c.Status != StatusRunning {
		c.HP = 0
	}
	if status != StatusRunning { // a flight is EvFled (tryFlee)
		e.factFell(c)
	}
	e.Cpos[idx].Size = 0 // DS:5F2A + 4*idx
	e.CombatRedrawMap()
	if e.ClearCombatInfo(c) {
		e.removeCombatAffects(c)
	}
}

// removeInvisibility = ovr25:0FD2.
func (e *RuleState) removeInvisibility(c *Character) {
	for a := e.FindAffect(c, affInvisible); a != nil; a = e.FindAffect(c, affInvisible) {
		e.RemoveAffect(c, affInvisible, a)
	}
}

// combatAffects = DS:0C14: affects that end with the fight (or a death).
var combatAffects = [16]AffectType{affTrainingManual, AffCharmed, affCoughing, affHelpless, affAnimated, affSnakeCharm, AffHeld, affAsleep, affRepulsed, affHeldFast, AffRegenerating, affFallsDead, affHeal3, affEndsSpiritualHammer, affStrikeTied, affBloodDrain}

// removeTiedAffects = ovr25:1041 (DS:0C24): the tied-strike and blood-drain links end.
func (e *RuleState) removeTiedAffects(c *Character) {
	for _, t := range []AffectType{affStrikeTied, affBloodDrain} {
		e.RemoveAffect(c, t, nil)
	}
}

// CureAffect = ovr25:1081 (char, type): "is Cured" and the affect goes.
func (e *RuleState) CureAffect(c *Character, typ AffectType) bool {
	a := e.FindAffect(c, typ)
	if a == nil {
		return false
	}
	e.CombatSay(c, "is Cured", true)
	e.RemoveAffect(c, typ, a)
	return true
}

// strValue = ovr25:10E4 (str, str00): one byte for strength, 18/xx as
// xx+1 (18/00 = 101), others +100.
func strValue(str Score, str00 byte) byte {
	if str == 18 {
		return str00 + 1
	}
	return byte(str + 0x64)
}

// strDecode = ovr25:110D: the inverse, from an affect's saved strength.
func strDecode(a *Affect) (str Score, str00 byte) {
	str = Score(a.Data & 0x7F)
	if str <= 0x65 {
		return 18, byte(str - 1)
	}
	return str - 0x64, 0
}

// SetStrength = ovr25:115C (char, str, str00, out): magical strength
// (affects 0x0C / 0x26 remember the natural value). Returns false, with
// the request encoded, when it would lower the current strength.
func (e *RuleState) SetStrength(c *Character, str Score, str00 byte) (byte, bool) {
	cur, cur00 := c.Stats[StatStr], c.Str00
	if str < cur || str == 18 && str00 < cur00 {
		return strValue(str, str00) + 0x80, false
	}
	i := slices.IndexFunc(c.Affects, func(a *Affect) bool {
		return (a.Type == affEnlarged || a.Type == affStrength) && a.Data < 0x80
	})
	if i >= 0 {
		keep := c.Affects[i]
		v := strValue(cur, cur00)
		cur, cur00 = strDecode(keep)
		keep.Data = v + 0x80
	}
	c.Stats[StatStr] = str
	c.Str00 = str00
	e.ApplyItems(c)
	return strValue(cur, cur00), true
}

// Saving throw types (the Saves order).
const (
	SaveVsParalysis = iota // paralysis, poison, death magic
	saveVsPetrify          // petrification, polymorph
	saveVsWand             // rod, staff, wand
	saveVsBreath
	SaveVsSpell
)

// Damage type bits (rv.dmgType).
const (
	DmgFire     = 0x01
	dmgCold     = 0x02
	dmgElectric = 0x04
	DmgMagic    = 0x08
	dmgAcid     = 0x10
)

// DamageMode values for Damage (what a successful save does).
const (
	saveNegate = 1
	saveHalf   = 2
)

// Damage = ovr25:1344 (char, damage, mode, saved): reports and applies
// damage, the lost spell, and going down.
func (e *RuleState) Damage(c *Character, dmg int, mode byte, saved bool) {
	e.Rv.Damage = int16(dmg)
	e.AffectGroup(c, grpDamage)
	if saved {
		switch mode {
		case saveNegate:
			e.Rv.Damage = 0
		case saveHalf:
			e.Rv.Damage /= 2
		}
	} else {
		e.AffectGroup(c, grpUnsavedDamage)
	}
	if e.Rv.Damage == 0 {
		return
	}
	var msg string
	if e.Rv.Damage != 1 {
		msg = "takes " + strconv.Itoa(int(e.Rv.Damage)) + " points of damage "
	} else {
		msg = "takes 1 point of damage "
	}
	switch e.Rv.DmgType &^ DmgMagic {
	case 1:
		msg += "from Fire"
	case 2:
		msg += "from Cold"
	case 4:
		msg += "from Electricity"
	case 0x10:
		msg += "from Acid"
	}
	if e.Rv.DmgType&DmgMagic == e.Rv.DmgType {
		msg += "from Magic"
	}
	e.hitMessage(c, false, msg)
	e.ApplyDamage(c, int(e.Rv.Damage))
	if e.BeginFlag == 5 && c.Combat != nil {
		c.Combat.CanCast = 0
		if c.Combat.Spell > 0 {
			e.CombatSay(c, "lost a spell", true)
			e.forgetSpell(c, SpellID(c.Combat.Spell))
			c.Combat.Spell = 0
		}
	}
	if c.Okay == 0 {
		msg = "Goes Down"
		if c.Status == StatusDying {
			msg += ", and is Dying"
		}
		if s := c.Status; s >= StatusDead && s <= StatusGone {
			msg = "is killed"
		}
		e.CombatSay(c, msg, true)
		if e.BeginFlag == 5 {
			e.removeCombatAffects(c)
			e.AffectGroup(c, grpDown)
			if c.Okay == 0 {
				e.combatRemoveFromMap(c)
			}
		}
	}
}

// Heal = ovr25:175D (char, amount, magical): only the living (okay,
// animated, unconscious, dying) heal; affect 0x32 blocks natural healing.
func (e *RuleState) Heal(c *Character, amount int, magical bool) bool {
	switch c.Status {
	case 0, 1, 4, 5:
	default:
		return false
	}
	if !(c.HP < c.MaxHP && magical) {
		if e.FindAffect(c, affMummyDisease) != nil || magical {
			return false
		}
	}
	c.HP += byte(amount)
	if c.HP > c.MaxHP {
		c.HP = c.MaxHP
	}
	if c.Okay == 0 {
		if c.Status == StatusDying {
			c.Status = StatusUnconscious
		}
		if c.Status == StatusUnconscious && e.BeginFlag != 5 {
			e.callAffect(affGetUp, c, nil, 1)
		}
	}
	return true
}

// Revive = ovr25:1869 (char, hp): back on its feet if its square is free.
func (e *RuleState) Revive(c *Character, hp byte) bool {
	if !e.combatSquareFree(c, e.CombatX(c), e.CombatY(c), true) {
		return false
	}
	c.Status = StatusOkay
	c.Okay = 1
	c.HP = hp
	if e.BeginFlag == 5 {
	}
	msg := "gets back up"
	if c.Side == SideMonsters {
		msg = "stands up and grins"
	}
	e.CombatSay(c, msg, true)
	e.CountSides()
	return true
}

// hitMessage = 00BA:1FEB (char, magical, message): in combat the target
// flashes (once per message delay step for magic) with a sound.
func (e *RuleState) hitMessage(c *Character, magical bool, msg string) {
	// fixed: the original cut msg to 40 characters ("...damage from Electric"); every path wraps it
	if e.BeginFlag == 5 { // the target's flash is the presenter's (Struck, Attacked)
		if magical {
			e.Sound(e.hitEvent(c, "hit_good"))
		} else {
			e.Sound(e.hitEvent(c, "hit"))
		}
	}
	e.CombatSay(c, msg, true)
}

// ApplyDamage = 00BA:2228 (char, damage): hit points down to 0, then
// unconscious / dying (bleeding rounds) / dead at 10 below.
func (e *RuleState) ApplyDamage(c *Character, dmg int) {
	if s := c.Status; s >= StatusDead && s <= StatusGone { // fixed, as the Pascal reference: a hit on a body made it Dying again
		return
	}
	if e.BeginFlag == 5 {
		dmg = e.Rules().Damage(e, c, dmg)
	}
	e.factHarm(c, dmg)
	hp := int(c.HP)
	over, left := 0, 0
	if hp >= dmg {
		left = hp - dmg
	} else {
		over = dmg - hp
	}
	over &= 0xFF
	switch {
	case over > 9 || left == 0 && c.Status == StatusAnimated:
		c.Status = StatusDead
	case over > 0:
		c.Status = StatusDying
		if e.BeginFlag == 5 && c.Combat != nil {
			c.Combat.Dying = byte(over)
		}
	case left == 0:
		c.Status = StatusUnconscious
	}
	if s := c.Status; s == StatusOkay || s == StatusAnimated {
		c.HP = byte(left)
		return
	}
	c.Okay = 0
	c.HP = 0
	if e.BeginFlag == 5 {
		e.Rv.Side[int8(c.Side)&1]--
		if c.Combat != nil {
			c.Combat.Delay = 0
		}
	}
}

// forgetSpell = 00BA:148B (spell): the character's first memorised copy of
// the spell is gone. fixed: the original always took it from the selected
// character, so a hit that interrupted a spell robbed the attacker.
func (e *RuleState) forgetSpell(c *Character, sp SpellID) (slot int) {
	if c == nil {
		return -1
	}
	for i := 0; i <= 0x14; i++ {
		if c.Memorized[i] == sp {
			c.Memorized[i] = 0
			return i
		}
	}
	return -1
}

// healedMessage = 00BA:2348 (char).
func (e *RuleState) healedMessage(c *Character) {
	msg := "is partially healed"
	if c.HP == c.MaxHP {
		msg = "is fully healed"
	}
	e.CombatSay(c, msg, true)
}

// CountSides = 00BA:23DB: standing members of each side (DS:6814/6815).
func (e *RuleState) CountSides() {
	e.Rv.Side = [2]int8{}
	for _, c := range e.Party {
		if c.Okay != 0 {
			e.Rv.Side[int8(c.Side)&1]++
		}
	}
}

// ClearCombatInfo = 00BA:2637 (char): reset the combat block; always true.
func (e *RuleState) ClearCombatInfo(c *Character) bool {
	if c.Combat != nil {
		ci := c.Combat
		ci.Delay, ci.Spell, ci.Guarding, ci.Moves = 0, 0, 0, 0
	}
	return true
}
