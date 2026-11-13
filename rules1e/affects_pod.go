package rules1e

import (
	"gobox/rules"
	"strconv"
)

// podHandlers are PoD's affect handlers (fold T4c): its variants of
// POR's affects and its own, by the names PodAffectVariants gives them,
// each ported exactly from pod-go/rules and checked against it there.
var podHandlers = map[HandlerName]AffectHandler{
	// O07C's empty entries: affects that only mark the character.
	"friendlyPod":     (*RuleState).noEffect,
	"mirrorPod":       (*RuleState).noEffect,
	"magicalBootsPod": (*RuleState).noEffect, // O07C:1BA0 doubles a local: does nothing
	"pod87":           (*RuleState).noEffect,
	// O07C:00F5: the turn is over.
	"befuddledPod": (*RuleState).clearCombatBlock,
	"stunnedPod":   (*RuleState).clearCombatBlock,

	"electricityImmunePod":  (*RuleState).electricityImmunePod,
	"needsPlus3WeaponPod":   func(e *RuleState, c *Character, a *Affect, flag byte) { e.needsWeaponPlusPod(plusNeeded3Pod) },
	"plainWeaponsImmunePod": func(e *RuleState, c *Character, a *Affect, flag byte) { e.needsWeaponPlusPod(plusNeededPlainPod) },
	"needsPlus2WeaponPod":   func(e *RuleState, c *Character, a *Affect, flag byte) { e.needsWeaponPlusPod(plusNeeded2Pod) },
	"halfFromPiercingPod":   (*RuleState).halfFromPiercingPod,
	"halfFromCrushingPod":   (*RuleState).halfFromCrushingPod,
	"luckStonePod":          func(e *RuleState, c *Character, a *Affect, flag byte) { e.halveBelowPlusPod(plusNeeded3Pod) },
	"halfFromWeakPod":       func(e *RuleState, c *Character, a *Affect, flag byte) { e.halveBelowPlusPod(plusNeeded4Pod) },
	"invisiblePod":          (*RuleState).invisiblePod,
	"prayerPod":             (*RuleState).prayerPod,
	"weakenedPod":           (*RuleState).weakenedPod,
	"diseaseImmunePod":      (*RuleState).diseaseImmunePod,
	"dragonFearPod":         (*RuleState).dragonFearPod,
	"fireResistDruidPod":    (*RuleState).fireResistDruidPod,
	"mindBlankPod":          (*RuleState).mindBlankPod,
	"dancingPod":            func(e *RuleState, c *Character, a *Affect, flag byte) { c.AC[0] -= acStepPod; c.AC[1] -= acStepPod },
	"ironSkinPod":           func(e *RuleState, c *Character, a *Affect, flag byte) { c.AC[0] += acStepPod; c.AC[1] += acStepPod },
	"gazeImmunePod":         (*RuleState).gazeImmunePod,
	// O07C:060D runs at the flag-0 call: POR's poisonEnds then.
	"slowPoisonPod": func(e *RuleState, c *Character, a *Affect, flag byte) {
		if flag == 0 {
			e.poisonEnds(c, a, flag)
		}
	},
	"dragonFearAttackPod":   (*RuleState).dragonFearAttackPod,
	"bulletRakePod":         (*RuleState).bulletRakePod,
	"fireDamage2d6Pod":      func(e *RuleState, c *Character, a *Affect, flag byte) { e.fireDamagePod(c, 2) },
	"fireTouchPod":          func(e *RuleState, c *Character, a *Affect, flag byte) { e.fireDamagePod(c, 2) },
	"fireDamage4d6Pod":      func(e *RuleState, c *Character, a *Affect, flag byte) { e.fireDamagePod(c, 4) },
	"slainByBoltsPod":       (*RuleState).slainByBoltsPod,
	"invisibleToAnimalsPod": (*RuleState).invisibleToAnimalsPod,
	"vorpalWeaponPod":       (*RuleState).vorpalWeaponPod,
	// O07C:0E5B: POR's agesOnce at flag 0, the aged bit one higher. PoD's one
	// call doubles moves and attacks; rules1e's speed group asks per value.
	"hastePod": func(e *RuleState, c *Character, a *Affect, flag byte) {
		if flag == 0 {
			e.agesOnceBit(c, a, hasteAgedBitPod)
		}
	},
	// O07C:1030: POR's halveAttacks at flag 0.
	"slowPod": func(e *RuleState, c *Character, a *Affect, flag byte) {
		if flag == 0 {
			e.halveAttacks(c, a, flag)
		}
	},
	"weaponSequencePod": (*RuleState).weaponSequencePod,
}

// monsterNoHeadPod: a vorpal blade cannot behead it (PoD MonsterFlags 32).
const monsterNoHeadPod MonsterFlags = 32

// vorpalWeaponPod = O07C:2E7E: in a fight a critical hit beheads a foe that
// has a head to lose.
func (e *RuleState) vorpalWeaponPod(c *Character, a *Affect, flag byte) {
	if !e.inCombat() || flag != 0 || !e.Rv.Critical {
		return
	}
	if t := targetOf(c); t.MonsterFlags&monsterNoHeadPod == 0 {
		e.killWithMessage(t, StatusDead, "is beheaded")
	}
}

// weaponSequencePod's weapons by the hand's blow of the round: the dice and
// the name the blow shows. Each is +sequencePlusPod.
var weaponSequencePodWeapons = [...]struct {
	dice, sides int
	name        string
}{
	{1, 10, "Two-Handed Sword"},
	{2, 4, "Bastard Sword"},
	{1, 8, "Long Sword"},
	{1, 8, "Scimitar"},
	{1, 4, "Dagger"},
	{1, 6, "Spear"},
}

const sequencePlusPod = 4

// weaponSequencePod = O07C:2EE2: each blow of the round (hand 1's count,
// SlotShots) strikes with the next weapon of the sequence: its dice, +4 and
// the strength bonus; its name shows in the text window (no flag test).
// Blows past the sixth keep the damage and show nothing.
func (e *RuleState) weaponSequencePod(c *Character, a *Affect, flag byte) {
	blow := int(e.SlotShots[1])
	if blow < 1 || blow > len(weaponSequencePodWeapons) {
		return
	}
	w := weaponSequencePodWeapons[blow-1]
	e.Rv.Damage = int16(int(strDamage(c)) + e.Dice(w.dice, w.sides) + sequencePlusPod)
	e.Emit(rules.Narrated{Text: w.name + "+" + strconv.Itoa(sequencePlusPod), More: true})
}

// acStepPod is the armour class dancing costs and iron skin gives (O07C:30DF,
// O07C:3221; no flag test).
const acStepPod = 4

// diseaseImmunePod = O07C:1C1F.
func (e *RuleState) diseaseImmunePod(c *Character, a *Affect, flag byte) {
	if flag == 0 {
		e.blockAffect(affCauseDisease)
	}
}

// dragonFearLossPod is what dragon fear takes off damage and to-hit.
const dragonFearLossPod = 2

// dragonFearPod = O07C:319D: two less damage and to hit (no flag test).
func (e *RuleState) dragonFearPod(c *Character, a *Affect, flag byte) {
	if e.Rv.Damage > dragonFearLossPod {
		e.Rv.Damage -= dragonFearLossPod
	} else {
		e.Rv.Damage = 0
	}
	e.Rv.HitRoll -= dragonFearLossPod
}

// fireResistPerDataPod: the druid's fire resistance soaks this much fire
// per point of data.
const fireResistPerDataPod = 12

// fireResistDruidPod = O07C:30FE: the fire is soaked up while the data
// lasts, twelve points for each; then the affect ends.
func (e *RuleState) fireResistDruidPod(c *Character, a *Affect, flag byte) {
	if e.Rv.DmgType&DmgFire == 0 {
		return
	}
	pool := int(a.Data) * fireResistPerDataPod
	if pool > int(e.Rv.Damage) {
		pool -= int(e.Rv.Damage)
		e.Rv.Damage = 0
		a.Data = byte(pool / fireResistPerDataPod)
		if a.Data == 0 {
			e.RemoveAffect(c, affFireResistDruidPod, nil)
		}
		return
	}
	e.Rv.Damage -= int16(pool)
	e.RemoveAffect(c, affFireResistDruidPod, nil)
}

// mindBlankPod = O07C:3240: no charm, confusion, hold, feeblemind or fear
// takes hold.
func (e *RuleState) mindBlankPod(c *Character, a *Affect, flag byte) {
	for _, typ := range []AffectType{AffCharmed, affConfused, AffHeld, affFeebleminded, affFrightened} {
		e.blockAffect(typ)
	}
}

// gazeImmuneSavePod is the saving throw a gaze immunity makes; PoD's gazes
// carry the acid bit.
const gazeImmuneSavePod = 20

// gazeImmunePod = O07C:3278.
func (e *RuleState) gazeImmunePod(c *Character, a *Affect, flag byte) {
	if e.Rv.DmgType&dmgAcid != 0 {
		e.Rv.Save = gazeImmuneSavePod
	}
}

// invisiblePod = O07C:0886: attackers who cannot see it are hampered (no
// veto, unlike POR's).
func (e *RuleState) invisiblePod(c *Character, a *Affect, flag byte) {
	if flag == 0 && e.FindAffect(e.Acting(), affDetectInvisible) == nil {
		e.Rv.HitRoll -= 4
	}
}

// prayerSidePod: the praying side in a prayer's data (POR keeps it in bit 4).
const prayerSidePod = 128

// prayerPod = O07C:11E4: the praying side gains, the other loses.
func (e *RuleState) prayerPod(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		return
	}
	if c.Side == a.Data&prayerSidePod/prayerSidePod {
		e.plusOneSaveHit(c, a, flag)
		return
	}
	e.Rv.HitRoll--
	e.Rv.Save--
}

// Weakness: its timer and the strength floor.
const (
	weakenTimePod      = 60
	weakestStrengthPod = 3
)

// weakenedPod = O07C:105C: strength drops while it lasts; at the bottom
// the victim is helpless (again each time: PoD does not look for it).
func (e *RuleState) weakenedPod(c *Character, a *Affect, flag byte) {
	if flag != 0 || !e.AddAffectNow(c, affWeakened, a.Data, weakenTimePod) {
		return
	}
	if c.Stats[StatStr] > weakestStrengthPod {
		e.CombatSay(c, "is weakened", true)
		c.Stats[StatStr]--
		return
	}
	e.AddAffect(c, affHelpless, 0, 0xFF, 0)
}

// The weapon plus PoD's needsWeaponPlus handlers ask for, and the ones
// below which luckStone and halfFromWeak halve.
const (
	plusNeededPlainPod = 1
	plusNeeded2Pod     = 2
	plusNeeded3Pod     = 3
	plusNeeded4Pod     = 4
)

// A monster without a weapon hurts a needsWeaponPlus foe with hit dice of
// at least (plus-1)*monsterDicePerPlusPod + monsterDiceForPlus1Pod.
const (
	monsterDicePerPlusPod  = 2
	monsterDiceForPlus1Pod = 5
)

// PoD's base item byte +7 (ItemBase.Material) says how a weapon hurts:
// up to 1 pierces or cuts, 128 and above crushes.
const (
	weaponKindPiercingPod   = 1
	weaponKindCrushedMinPod = 128
)

// blockDamageTypePod = O07C:00DD (types): damage of any of the types is stopped.
func (e *RuleState) blockDamageTypePod(types byte) {
	if e.Rv.DmgType&types != 0 {
		e.blockAffect(affAny)
	}
}

// electricityImmunePod = O07C:0230.
func (e *RuleState) electricityImmunePod(c *Character, a *Affect, flag byte) {
	if flag == 0 {
		e.blockDamageTypePod(dmgElectric)
	}
}

// needsWeaponPlusPod = O07C:1536 (plus): no damage from the acting
// attacker's weapon below the plus; a monster without a weapon hurts only
// with enough hit dice (5, 7, 9 for +1, +2, +3), anyone else not at all.
func (e *RuleState) needsWeaponPlusPod(plus int) {
	sel := e.Acting()
	if w := e.readiedWeapon(sel); w != nil {
		if int(int8(w.Plus)) < plus {
			e.Rv.Damage = 0
		}
		return
	}
	if sel.raceNumber() == 0 {
		if (plus-1)*monsterDicePerPlusPod+monsterDiceForPlus1Pod > int(sel.MaxLevel) {
			e.Rv.Damage = 0
		}
		return
	}
	e.Rv.Damage = 0
}

// halfFromPiercingPod = O07C:17F2: half damage from a piercing or cutting
// weapon (no flag test).
func (e *RuleState) halfFromPiercingPod(c *Character, a *Affect, flag byte) {
	if w := e.readiedWeapon(e.Acting()); w != nil && e.itemBaseFlags(w) <= weaponKindPiercingPod {
		e.halveDamage()
	}
}

// halfFromCrushingPod = O07C:1841.
func (e *RuleState) halfFromCrushingPod(c *Character, a *Affect, flag byte) {
	if w := e.readiedWeapon(e.Acting()); w != nil && e.itemBaseFlags(w) >= weaponKindCrushedMinPod {
		e.halveDamage()
	}
}

// halveBelowPlusPod = O07C:1EC3 (luck stone, +3) and O07C:1F06 (+4): half
// damage from weapons below the plus.
func (e *RuleState) halveBelowPlusPod(plus int8) {
	if w := e.readiedWeapon(e.Acting()); w != nil && int8(w.Plus) < plus {
		e.halveDamage()
	}
}

// downAffectsPod = PoD DS:0C6D 1..16: the affects going out of action ends.
var downAffectsPod = []AffectType{affSticksToSnakes, affElectricityImmunePod, AffCharmed, affSilenced,
	affSpiritualHammer, affBefuddledPod, affCoughing, affHelpless, affConfused, AffHeld, affAsleep,
	affBerserk, affStunnedPod, affBerserkFoe, affFrightened, affDancingPod}

// The second-hand blows of a bullet rake: one time in four many, else few.
const (
	rakeManyBlowsPod = 8
	rakeFewBlowsPod  = 4
)

// bulletRakePod = O07C bulletRakeHandler (no flag test).
func (e *RuleState) bulletRakePod(c *Character, a *Affect, flag byte) {
	if e.Dice(1, 4) == 1 {
		c.Attacks[1] = rakeManyBlowsPod
		return
	}
	c.Attacks[1] = rakeFewBlowsPod
}

// fireDamagePod = O07C:2A92 (dice): in a fight, the target takes dice d6
// magic fire, half on a saving throw against spells (no flag test).
func (e *RuleState) fireDamagePod(c *Character, dice int) {
	if !e.inCombat() {
		return
	}
	t := combatInfo(c).Target
	if t == nil {
		return
	}
	e.Rv.DmgType = DmgFire | DmgMagic
	dmg := e.RollDice(dice, 6)
	e.Damage(t, dmg, saveHalf, e.SavingThrow(t, SaveVsSpell, 0))
}

// itemTypeBolts2Pod is PoD's second kind of bolts (its item type 74).
const itemTypeBolts2Pod = 74

// slainByBoltsPod = O07C slainByBoltsHandler: a bolt of the second kind kills.
func (e *RuleState) slainByBoltsPod(c *Character, a *Affect, flag byte) {
	if w := e.readiedWeapon(e.Acting()); w != nil && w.Type == itemTypeBolts2Pod {
		e.killWithMessage(c, StatusDead, "is slain")
	}
}

// invisibleToAnimalsPod = O07C:1207: as invisiblePod, attackers are
// hampered without a veto (SotSB's vetoes: invisibleToAnimalsSotsb). Like
// SotSB's it tests the invisible one's own kind.
func (e *RuleState) invisibleToAnimalsPod(c *Character, a *Affect, flag byte) {
	if flag == 0 && c.CreatureFlags&creatureAnimalsSeen != 0 && e.FindAffect(c, affDetectInvisible) == nil {
		e.Rv.HitRoll -= 4
	}
}

// dragonFearSavePod is the saving throw's penalty against a dragon's fear.
const dragonFearSavePod = -4

// dragonFearAttackPod = O07C:2209: every standing foe who fails a saving
// throw against petrification (at -4) and is not already afraid is
// terrified (as an area effect). A variant: SotSB's fear aura makes the low
// levels flee (fearAuraSotsb).
func (e *RuleState) dragonFearAttackPod(c *Character, a *Affect, flag byte) {
	e.areaSpell = 1
	for _, foe := range e.Party {
		if foe.Side != c.Side && foe.Status == StatusOkay && !e.SavingThrow(foe, saveVsPetrify, dragonFearSavePod) &&
			e.FindAffect(foe, affDragonFearPod) == nil {
			e.AddAffectMsg(foe, affDragonFearPod, 0, 0, 1, saveNegate, false, "is terrified")
		}
	}
	e.areaSpell = 0
}
