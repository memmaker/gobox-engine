package rules1e

import "slices"

// spellEffectSotsbFirst runs the effect of the unified spells 104..127, the ids
// sotsb has first (spellids_games.go), ported from its spellFx (fold T3f).
// false: the spell is not one of them (or not ported yet).
func (e *RuleState) spellEffectSotsbFirst(spell SpellID) bool {
	switch spell {
	case spHarm:
		e.fxHarm()
	case spElixirOfYouth:
		e.fxYounger()
	case spBarkskin:
		e.fxBarkskin()
	case spUnnamed95Sotsb, spGlobeOfInvulnerability:
		e.applySpell(int(e.CurSpell), 0, 0, 0, 0, "is Protected") // fxMinorGlobe
	case spMassInvisibility:
		e.fxMassInvisibility()
	case spPowerWordStun:
		e.fxPowerWordStun()
	case spCureLightWoundsDruid:
		e.spellCureLightWounds()
	case spUnnamed101, spUnnamed102, spUnnamed103, spUnnamed104, spUnnamed105,
		spUnnamed106, spUnnamed107, spUnnamed108, spUnnamed109:
		e.fxNone()
	case spDeathSpell:
		e.fxDeathSpell()
	case spDisintegrate:
		e.killUnlessSaved(StatusDead, "disintegrates")
	case spDelayedBlastFireball:
		e.fxDelayedBlast()
	case spFleshToStone:
		e.killUnlessSaved(StatusStoned, "is turned to stone")
	default:
		return false
	}
	return true
}

const (
	youngestAge = 18   // fxYounger's floor
	slayMark    = 0xFF // Rv.Damage while grpAddAffect asks whether a slaying spell may land
	dmgSlaying  = 0x40 // Rv.DmgType of the slaying spells (SotSB dmgType64)

	monsterNoBreath MonsterFlags = 8 // SotSB: does not breathe (Death Spell passes it over)

	deathSpellBudgetDice  = 4
	deathSpellBudgetSides = 20
	deathSpellBudgetMul   = 2
	deathSpellBig6HP      = 51  // a 6 HD victim above this costs as 7 HD
	deathSpellOutOfReach  = 162 // more than any budget

	stunRoundsHP1 = 30 // Power Word Stun: up to this HP 4d4 rounds
	stunRoundsHP2 = 60 // 2d4
	stunRoundsHP3 = 90 // 1d4; more, none

)

// Barkskin (ruling): base armour class 6, one better per barkskinLevelStep
// caster levels; +1 on saves except against magic.
const (
	barkskinAC        = AcTen + 10 - 6 // AC 6 on the 60-minus-AC scale
	barkskinLevelStep = 4
	barkskinRounds    = 4 // duration: 4 rounds + 1 per caster level (AD&D 1e)
	barkskinCall      = 1 // the handler runs at removal (AC back)
)

// barkskinBase is the base armour class (60 minus AC) Barkskin gives at the
// caster level kept in its affect's Data.
func barkskinBase(a *Affect) int8 {
	return int8(barkskinAC + int(a.Data)/barkskinLevelStep)
}

// fxBarkskin (ruling): the target's skin becomes bark.
func (e *RuleState) fxBarkskin() {
	t := e.tgt(1)
	if t == nil {
		return
	}
	lvl := e.casterLevel(int(e.CurSpell))
	e.AddAffectMsg(t, affBarkskin, barkskinRounds+lvl, byte(lvl), barkskinCall, 0, false, "has Barkskin")
	e.ApplyItems(t)
}

// barkskinRuled (ruling): +1 on saves unless against magic; at removal the
// armour class is counted again without it (ApplyItems reads barkskinBase).
func (e *RuleState) barkskinRuled(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		c.Affects = slices.DeleteFunc(c.Affects, func(x *Affect) bool { return x == a })
		e.ApplyItems(c)
		return
	}
	if e.Rv.DmgType&DmgMagic == 0 && e.Rv.SaveType != SaveVsSpell {
		e.Rv.Save++
	}
}

// fxHarm = SotSB fxHarm: diseased, down to 1d4 hit points.
func (e *RuleState) fxHarm() {
	t := e.tgt(1)
	e.Rv.DmgType = DmgMagic
	e.AddAffectMsg(t, affCauseDisease, 0, 0, 0, 0, false, "")
	e.CombatSay(t, "is Diseased", true)
	e.Damage(t, int(t.HP-byte(e.Dice(1, 4))), 0, false)
}

// fxYounger = SotSB fxYounger: the caster is 2..5 years younger, no less than 18.
func (e *RuleState) fxYounger() {
	c := e.Acting()
	c.Age -= uint16(e.Dice(1, 4) + 1)
	if c.Age < youngestAge {
		c.Age = youngestAge
		e.CombatSay(c, "is now 18 years old", true)
		return
	}
	e.CombatSay(c, "becomes younger", true)
}

// deathSpellCost = SotSB deathSpellCost: a victim's price out of the budget, by hit dice.
func deathSpellCost(c *Character) byte {
	switch lvl := int8(c.MaxLevel); {
	case lvl == 0:
		return 1
	case lvl == 1:
		return 2
	case lvl >= 2 && lvl <= 3:
		return 8
	case lvl >= 4 && lvl <= 5:
		return 16
	case lvl == 6:
		if c.MaxHP <= deathSpellBig6HP {
			return 16
		}
		return 20
	case lvl >= 7 && lvl <= 8:
		return 20
	}
	return deathSpellOutOfReach
}

// fxDeathSpell = SotSB fxDeathSpell: victims drop dead in list order while
// the budget lasts; the breathless are passed over.
func (e *RuleState) fxDeathSpell() {
	e.areaSpell = 1
	e.Rv.DmgType = dmgSlaying
	budget := byte(e.Dice(deathSpellBudgetDice, deathSpellBudgetSides) * deathSpellBudgetMul)
	for i := 1; i <= e.NSpellTargets; i++ {
		t := e.tgt(i)
		cost := deathSpellCost(t)
		if t.MonsterFlags&monsterNoBreath == 0 && budget >= cost {
			budget -= cost
			e.killWithMessage(t, StatusDead, "drops dead")
		} else {
			e.SetTgt(i, nil)
		}
	}
}

// killUnlessSaved = SotSB killUnlessSaved: unless the target's affects stop
// it, a failed save vs spell leaves it in the status.
func (e *RuleState) killUnlessSaved(status byte, msg string) {
	e.Rv.DmgType = dmgSlaying
	t := e.tgt(1)
	e.Rv.Damage = slayMark
	e.AffectGroup(t, grpAddAffect)
	if e.Rv.Damage != 0 && !e.SavingThrow(t, SaveVsSpell, 0) {
		e.killWithMessage(t, status, msg)
	}
}

// Delayed Blast Fireball (ruling): a fireball of (caster level)d6+1 per die
// that waits 0 (none) .. delayedBlastMaxRounds rounds on the target square,
// the caster's answer to the spell's question (Fight.Choice).
const (
	delayedBlastMaxRounds = 5
	delayedBlastRadius    = 3
	delayedBlastCall      = 1    // the handler runs at removal: the blast goes off
	groundWall            = 0xFF // GroundInfo's first prop of a wall: nothing is laid on it
)

// delayedBlastDirs: the one square a waiting blast covers (8: the centre).
var delayedBlastDirs = [1]Dir{8}

// delayedBlastChoices are the spell's question's replies: the delay in rounds.
var delayedBlastChoices = []string{"0", "1", "2", "3", "4", "5"}

// delayedBlastDamage: the caster level plus that many d6, +1 per die by the ruling's count.
func (e *RuleState) delayedBlastDamage(lvl int) int { return lvl + e.RollDice(lvl, 6) }

// fxDelayedBlast = SotSB fxDelayedBlast (O113:5D90), the delay in rounds:
// none goes off at once as a fireball; else the blast waits on the target
// square as the caster's next numbered one, the delay its affect's time.
func (e *RuleState) fxDelayedBlast() {
	e.areaSpell = 1
	lvl := e.casterLevel(int(spDelayedBlastFireball))
	delay := int(min(max(e.Fight.Choice, 0), delayedBlastMaxRounds))
	if delay == 0 {
		e.applySpell(int(spDelayedBlastFireball), 0, 0, e.delayedBlastDamage(lvl), int(DmgFire|DmgMagic), "")
		return
	}
	sel := e.Acting()
	n := 0
	var same []*cloud
	for _, o := range e.Clouds {
		if o.Kind == CloudDelayedBlast {
			same = append(same, o)
		}
	}
	for i := 0; i < len(same)-1; i++ { // the last record is not counted (as createCloud)
		if same[i].Ptr00 == sel {
			n++
		}
	}
	cl := &cloud{Kind: CloudDelayedBlast, Ptr00: sel, Num: byte(n), Level: lvl, X: byte(e.TgtX), Y: byte(e.TgtY)}
	e.Clouds = append(e.Clouds, cl)
	e.AddAffect(sel, affDelayedBlastSotsb, delay, byte(n), delayedBlastCall) // Data: the caster's blast number (cloud.Num)
	_, g := e.MapAt(e.TgtX, e.TgtY)
	if g > 0 && e.GroundInfo(g)[0] < groundWall {
		cl.Covered[0] = 1
	}
	g = e.olderGround(CloudDelayedBlast, cl, e.TgtX, e.TgtY, g, true)
	cl.Ground[0] = g
	if cl.Covered[0] != 0 {
		e.Cmap.Ground[e.TgtY][e.TgtX] = e.grounds().DelayedBlast
	}
	e.CombatSay(sel, cloudSpecs[CloudDelayedBlast].message, true)
}

// groundUnderBody: the ground a body at x, y lies on, else g.
func (e *RuleState) groundUnderBody(x, y int, g Ground) Ground {
	for b := 1; b <= e.Nbodies; b++ {
		if int(int8(e.Bodies[b].X)) == x && int(int8(e.Bodies[b].Y)) == y {
			g = Ground(e.Bodies[b].Ground)
		}
	}
	return g
}

// delayedBlastSotsb = SotSB delayedBlastHandler (O0A4:30B1): when the
// caster's waiting blast ends in combat it goes off: everyone within three
// squares takes the damage, half on a save vs spell.
func (e *RuleState) delayedBlastSotsb(c *Character, a *Affect, flag byte) {
	if flag == 0 || !e.inCombat() {
		return
	}
	num := a.Data
	at := slices.IndexFunc(e.Clouds, func(cl *cloud) bool {
		return cl.Kind == CloudDelayedBlast && cl.Ptr00 == c && cl.Num == num
	})
	if at < 0 {
		return
	}
	cl := e.Clouds[at]
	x, y := int(cl.X), int(cl.Y)
	e.BuildNear(x, y, delayedBlastRadius, 0xFF, 1)
	var caught []*Character
	for _, n := range e.Near {
		caught = append(caught, e.Combatant(n.Index))
	}
	e.notice("Delayed Blast Fireball goes off")
	if cl.Covered[0] != 0 {
		e.Cmap.Ground[y][x] = e.groundUnderBody(x, y, cl.Ground[0])
	}
	e.Clouds = slices.Delete(e.Clouds, at, at+1)
	e.Rv.DmgType = DmgFire | DmgMagic
	damage := e.delayedBlastDamage(cl.Level)
	for _, t := range caught {
		if t != nil {
			e.Damage(t, damage, saveHalf, e.SavingThrow(t, SaveVsSpell, 0))
		}
	}
	e.Rv.DmgType = 0
}

// fxMassInvisibility = SotSB fxMassInvisibility: the caster's side, but
// those with an enemy next to them in combat.
func (e *RuleState) fxMassInvisibility() {
	e.areaSpell = 1
	side := e.Acting().Side
	for i := 1; i <= e.NSpellTargets; i++ {
		t := e.tgt(i)
		if t.Side != side || e.inCombat() && e.NearEnemies(t, 1) > 0 {
			e.SetTgt(i, nil)
		}
	}
	e.applySpell(int(e.CurSpell), 0, 0, 0, 0, "is Invisible (spell)")
}

// fxPowerWordStun = SotSB fxPowerWordStun: the stun lasts by the target's
// hit points.
func (e *RuleState) fxPowerWordStun() {
	t := e.tgt(1)
	var rounds int
	switch {
	case t.HP >= 1 && t.HP <= stunRoundsHP1:
		rounds = e.Dice(4, 4)
	case t.HP > stunRoundsHP1 && t.HP <= stunRoundsHP2:
		rounds = e.Dice(2, 4)
	case t.HP > stunRoundsHP2 && t.HP <= stunRoundsHP3:
		rounds = e.Dice(1, 4)
	}
	s := e.CurSpell
	e.AddAffectMsg(t, e.spellAffect(s), rounds, byte(e.casterLevel(int(s))), 0,
		e.SpellRec(s, spellFieldSaveMode), false, "is STUNNED")
}
