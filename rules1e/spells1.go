package rules1e

// ovr23 (spells), first part (0x0000..0x100F) and the tail (0x2540..0x37B5).
// The spell effects between are in spells2.go. The spell record is
// DS:31FA + 16*spell (see spellRec): [0] class (casterLevel), [1] list
// group, [2]/[3] duration base/per level, [4]/[5] amount base/per level,
// [6] minimum 1 flag, [7] targeting kind, [8] save mode, [9] save type,
// [0xA] affect type.

import (
	"fmt"
	"gobox/core"
	"gobox/core/goldbox"
)

// tgt / setTgt = DS:6C91 + 4*i: the spell target list (1-based). Slot 1 is
// DS:6C95, kept in e.spellTarget.
func (e *RuleState) tgt(i int) *Character {
	if i == 1 {
		return e.spellTarget
	}
	return e.spellTargets[i]
}

func (e *RuleState) SetTgt(i int, c *Character) {
	if i == 1 {
		e.spellTarget = c
	} else {
		e.spellTargets[i] = c
	}
}

// SpellName = DS:28E9 + 0x29*spell.
func (e *RuleState) SpellName(spell int) string {
	if spell < 0 || spell >= len(SpellNames) {
		return "" // an item or monster power (57-67): no name (the original reads past its table)
	}
	return SpellNames[spell]
}

// SpellDuration = ovr23:0729 (spell): the duration of the spell in rounds (or
// turns); never 0 when the record says so.
func (e *RuleState) SpellDuration(spell int) int {
	s := byte(spell)
	var v byte
	if !e.UsingItem {
		v = e.SpellRec(SpellID(s), 2) + byte(int(e.SpellRec(SpellID(s), 3))*e.casterLevel(spell))
	} else {
		v = e.SpellRec(SpellID(s), 2) + byte(int(e.SpellRec(SpellID(s), 3))*6)
	}
	if v == 0 && e.SpellRec(SpellID(s), 6) != 0 {
		v = 1
	}
	if v == 0xFF {
		v = 1
	}
	return int(v)
}

// spellAmount = ovr23:07D6 (spell): the amount (damage or rounds) of the
// spell; a near procedure of the overlay.
func (e *RuleState) spellAmount(spell int) int {
	var v int
	switch spell {
	case 0x28:
		v = e.Dice(1, 6) * 10
	case 0x39, 0x3D:
		v = e.Dice(5, 4)
	case 0x3B:
		v = e.Dice(1, 4)*10 + 0x28
	case 0x3F:
		if e.BeginFlag == 5 {
			v = e.Dice(2, 10) * 10
		} else {
			v = (e.Dice(1, 10) + 10) * 10
		}
	case 0x43:
		v = 0x5A0
	case int(spMindBlank): // PoD spellAmount
		v = mindBlankRoundsPod
	case int(spOttosIrresistibleDance):
		v = e.Dice(1, ottoDanceDiceSides) + ottoDanceBonusRounds
	default:
		v = int(e.SpellRec(SpellID(spell), 4)) + int(e.SpellRec(SpellID(spell), 5))*e.casterLevel(spell)
	}
	return int(int16(v))
}

// applySpell = ovr23:08C3 (spell, level, call, dmg, dmgType, msg): applies
// the spell record to every target: saving throw, damage and affect.
// level 0 means the caster level; dmg is also the flag that sets the damage
// type (DS:6819); dmg is cleared for the rest when an attack misses.
func (e *RuleState) applySpell(spell, level, call, dmg, dmgType int, msg string) {
	msg = goldbox.Cut(msg, msgMax)
	if dmg == 0 {
		e.Rv.DmgType = 0
	} else {
		e.Rv.DmgType = byte(dmgType)
	}
	if e.NSpellTargets == 0 {
		return
	}
	lvl := byte(level)
	if level == 0 {
		lvl = byte(e.cappedCasterLevel(spell))
	}
	s := byte(spell)
	for i := 1; i <= e.NSpellTargets; i++ {
		t := e.tgt(i)
		if t == nil {
			continue
		}
		saved := false
		if e.SpellRec(SpellID(s), 8) != 0 {
			saved = e.SavingThrow(t, int(e.SpellRec(SpellID(s), 9)), 0)
		}
		if e.SpellRec(e.CurSpell, 2) == 0xFF {
			e.ApplyItems(t)
			e.AffectGroup(t, 0xB)
			if !e.AttackHits(e.Acting(), t, int(t.AC[0])) {
				dmg = 0
				saved = true
			}
		}
		if dmg > 0 {
			e.Damage(t, dmg, e.SpellRec(SpellID(s), 8), saved)
		}
		if e.SpellRec(SpellID(s), spellAffectField) > 0 {
			e.AddAffectMsg(t, e.spellAffect(SpellID(s)), e.spellAmount(spell), lvl, byte(call), e.SpellRec(SpellID(s), 8), saved, msg)
		}
	}
	e.Rv.DmgType = 0
}

// spellAffectField is the spell record's affect: the game's affect number.
const spellAffectField = 0xA

// spellAffect is the spell's affect as an id (the record holds the game's
// number; POR's numbers are the ids). A number outside the game's affectIds
// is a broken rules.json.
func (e *RuleState) spellAffect(spell SpellID) AffectType {
	id, err := e.T().affectID(e.SpellRec(spell, spellAffectField))
	if err != nil {
		panic(fmt.Sprintf("spell %d: %v", spell, err))
	}
	return id
}

// CastReply is what a cast was given before it ran: the fight's targeting
// (DS:6B84 repointed by Combat to ovr14:20C3) and castAsk's answers out
// of combat.
type CastReply struct {
	Fight bool       // the fight's targeting (chooseSpellTargets), else spellTargetSetup's
	Yes   bool       // yes to "Lose it?" / "Use it?"
	On    *Character // the member the spell goes on (nil: none)
}

// spellTargetHook = lcall [DS:6B84].
func (e *RuleState) spellTargetHook(spell, mode int, r CastReply) bool {
	if r.Fight {
		return e.chooseSpellTargets(spell, mode)
	}
	return e.spellTargetSetup(spell, mode, r.On)
}

// spellTargetSetup = ovr23:0A91 (spell, mode, var ok): fills the target
// list by the targeting kind of the spell (1 the caster, 2 a chosen member,
// 4 the whole party); ok is false for any other kind or when the choice is
// cancelled. This is slot 0 of the DS:6B84 table; mode is not used.
func (e *RuleState) spellTargetSetup(spell, mode int, on *Character) (ok bool) {
	if e.SpellCaster == nil {
		e.SpellCaster = e.Acting()
	}
	e.spellTarget = e.Acting()
	e.NSpellTargets = 1
	ok = true
	switch e.SpellRec(SpellID(spell), 7) {
	case 1:
	case 2:
		e.SpellCaster = on // SelectCharacter("Cast Spell on whom"), asked before the cast (castAsk)
		if e.SpellCaster == nil {
			e.spellTarget = nil
			e.NSpellTargets = 0
			ok = false
		} else {
			e.spellTarget = e.SpellCaster
		}
	case 4:
		for i, c := range e.Party {
			e.SetTgt(i+1, c)
		}
		if len(e.Party) > 0 {
			e.NSpellTargets = len(e.Party)
		}
	default:
		ok = false
	}
	return
}

// applySpellToSide = ovr23:0F4F (side, msg): drops the targets of the other side
// (and, for Bless in combat, those with an enemy next to them), then
// applies the spell record to the rest.
func (e *RuleState) applySpellToSide(side int, msg string) {
	msg = goldbox.Cut(msg, msgMax)
	e.areaSpell = 1
	for i := 1; i <= e.NSpellTargets; i++ {
		t := e.tgt(i)
		if t == nil || int(t.Side) != side ||
			e.CurSpell == SpBless && e.BeginFlag == 5 && e.NearEnemies(t, 1) > 0 {
			e.SetTgt(i, nil)
		}
	}
	e.applySpell(int(e.CurSpell), 0, 0, 0, 0, msg)
}

// CastSpell = ovr23:0C20 (spell, mode, b, var quit): casts the spell of
// the selected character. b non-zero announces it ("casts"); quit is set
// when a combat-only item was used up. A combat-only spell or item outside
// a fight asks to lose it / use it, r holds the answers. cancelled: the
// player cancelled the aim.
func (e *RuleState) CastSpell(spell SpellID, mode, b byte, r CastReply, quit *bool) (cancelled bool) {
	caster := e.Acting()
	ok := true
	if e.BeginFlag != 5 && e.SpellRec(spell, 7) == 0 { // the yes/no was asked before (castAsk)
		yes := r.Yes
		if !e.UsingItem {
			if yes {
				e.forgetSpell(caster, SpellID(spell))
			}
		} else if yes {
			*quit = true
		}
		b = 0
		ok = false
	}
	if b != 0 && !e.UsingItem {
		e.announceCast(caster, "casts", int(spell))
	}
	posed := false
	for ok {
		*quit = e.spellTargetHook(int(spell), int(mode), r) // the asm passes the caller's var
		if *quit {
			ok = false
			if e.BeginFlag == 5 {
				x, y := e.CombatX(caster), e.CombatY(caster)
				dir := 0
				for !e.InArc(x, y, e.TgtX, e.TgtY, dir) {
					dir++
				}
				e.TurnTo(caster, byte(dir)) // the cast's pose is the presenter's (CastSpell)
				if e.Fight.On {             // BUG-CLASSES 37b: the pose before the flight, as the original's
					e.Emit(core.Casting{From: unit(e.CombatIndex(caster)), Spell: SpellIDs[spell], Ats: []core.BoardPos{{X: e.TgtX, Y: e.TgtY}}})
					posed = true
				}
				switch spell {
				case 0x2F:
					e.Sound(string(core.CueFireball))
				case 0x33:
					e.Sound(string(core.CueLightning))
				default:
					e.Sound(string(core.CueSpell))
				}
				e.shoot(x, y, e.TgtX, e.TgtY, 0, SpellIDs[spell])
			}
			e.removeInvisibility(caster)
			slot := -1
			if !e.UsingItem {
				slot = e.forgetSpell(caster, SpellID(spell))
			}
			e.CurSpell = spell
			if e.Fight.On && !posed {
				e.Emit(core.Casting{From: unit(e.CombatIndex(caster)), Spell: SpellIDs[spell], Ats: []core.BoardPos{{X: e.TgtX, Y: e.TgtY}}}) // no name: the original's combat cast skips its lookup (ovr23:0C37), 57-67 are powers with none
			}
			e.spellEffect(spell)
			if e.spellKept && slot >= 0 {
				caster.Memorized[slot] = spell
			}
			e.spellKept = false
			e.CurSpell = 0
			e.areaSpell = 0
		} else if e.BeginFlag != 5 {
			ok = false
		} else {
			if mode == 0 { // extra: the player cancelled the aim (the original asked "Abort Spell?" and lost
				return true // the spell and the turn): the spell is kept and the turn goes on
			}
			e.notice("Spell Aborted")
			if !e.UsingItem {
				e.forgetSpell(caster, SpellID(spell))
			}
			ok = false
		}
	}
	return false
}

// spellRemoveCurse = ovr23:2540: the target (DS:6C95) loses the curse
// affect, or else the curse of its first cursed item.
func (e *RuleState) spellRemoveCurse() {
	t := e.spellTarget
	if e.CureAffect(t, AffBestowCurse) {
		e.hitMessage(t, true, "is un-cursed")
		return
	}
	found := false
	for _, it := range t.Items {
		if found {
			break
		}
		if it.Cursed != 0 {
			found = true
			it.Cursed = 0
		}
	}
	if found {
		e.hitMessage(t, true, "'s item is un-cursed")
	}
}

// affects is a spell whose whole effect is applySpell with its message:
// ovr23:10E9 "is affected" and its siblings (the ovr23 address is on
// each spellEffect case).
func (e *RuleState) affects(level, call int, msg string) {
	e.applySpell(int(e.CurSpell), level, call, 0, 0, msg)
}

// fxFireball = ovr23:266B: a burst of radius 2 round the target square,
// level d6 of fire (spFireballPower rolls its own level); the near list only
// becomes the target list when the area holds no word at +0x1CC.
func (e *RuleState) fxFireball() {
	var lvl int
	if e.CurSpell == spFireballPower {
		lvl = int(byte(e.Dice(1, 3)*2 + 1))
	} else {
		lvl = e.casterLevel(int(e.CurSpell))
	}
	if e.Area.In3D == 0 {
		e.BuildNear(e.TgtX, e.TgtY, 2, 0xFF, 1)
		for i, n := range e.Near {
			e.SetTgt(i+1, e.Combatant(n.Index))
		}
		e.NSpellTargets = len(e.Near)
	}
	e.applySpell(int(e.CurSpell), 0, 0, e.RollDice(lvl, 6), 9, "")
}

// removeAffectFromSide = ovr23:2764 (type, side, msg): the first casterLevel
// targets of that side lose the affect; the others are dropped. A nil
// target (garbage in the asm) is dropped as well.
func (e *RuleState) removeAffectFromSide(typ AffectType, side int, msg string) {
	msg = goldbox.Cut(msg, msgMax)
	e.areaSpell = 1
	lvl := byte(e.casterLevel(int(e.CurSpell)))
	for i := 1; i <= e.NSpellTargets; i++ {
		t := e.tgt(i)
		if t != nil && int(t.Side) == side && lvl > 0 {
			lvl--
			if e.CureAffect(t, typ) {
				e.SetTgt(i, nil)
			}
		} else {
			e.SetTgt(i, nil)
		}
	}
	e.applySpell(int(e.CurSpell), 0, 0, 0, 0, msg)
	for i := 1; i <= e.NSpellTargets; i++ {
		if t := e.tgt(i); t != nil {
			e.AffectGroup(t, 0x12)
		}
	}
}

// fxHaste = ovr23:2892: slow is removed from the caster's side.
func (e *RuleState) fxHaste() { e.removeAffectFromSide(AffSlow, int(e.Acting().Side), "is Hasted") }

// strikeSquare = ovr23:28BC (x, y, dmg, saveType): whoever stands on the
// square takes dmg (half on a save, DS:6819 = 0xC); true when the ground
// there stops a bolt.
func (e *RuleState) strikeSquare(x, y, dmg, saveType int) bool {
	blocked := false
	who, ground := e.MapAt(x, y)
	if ground > 0 && e.T().Ground[ground][0] == 0xFF {
		blocked = true
	}
	if who > 0 {
		t := e.Combatant(byte(who))
		e.Rv.DmgType = dmgElectric | DmgMagic
		e.Damage(t, int(byte(dmg)), 2, e.SavingThrow(t, saveType, 0))
		e.Rv.DmgType = 0
	}
	return blocked
}

// boltPath = ovr23:295B (reach, dmg, saveType, bounce): the bolt goes on
// past the target square, away from the caster, for reach*2 steps, hitting
// each square; where it is stopped it comes back toward the caster (once,
// when bounce is set, those squares count extra) with the direction
// reversed.
func (e *RuleState) boltPath(reach, dmg, saveType, bounce int, what core.SpellID) {
	if sel := e.Acting(); e.CombatX(sel) == e.TgtX && e.CombatY(sel) == e.TgtY {
		return
	}
	e.areaSpell = 1
	e.BoltWalk(reach, bounce, nil, func(x0, y0, x, y int) bool {
		e.shoot(x0, y0, x, y, 0, what)
		return e.strikeSquare(x, y, dmg, saveType)
	})
	e.areaSpell = 0
}

// BoltWalk is boltPath's line (split out, extra: the aim preview walks it
// dry): from the target square away from the selected caster; step (if
// any) sees each square crossed, strike each square a leg ends on (from
// the leg's start) and says whether the ground there stops the bolt.
// e.tgtX/tgtY move to where it turns, as in the original.
func (e *RuleState) BoltWalk(reach, bounce int, step func(x, y int), strike func(x0, y0, x, y int) bool) {
	prevWho, _ := e.MapAt(e.TgtX, e.TgtY)
	dirSign := 1
	sel := e.Acting()
	cx, cy := e.CombatX(sel), e.CombatY(sel)
	flag := bounce
	if cx == e.TgtX && cy == e.TgtY {
		return
	}
	var who Slot // who is uninitialised in the asm until the first step
	var ground Ground
	_, ground = e.MapAt(e.TgtX, e.TgtY)
	r := byte(reach * 2)
	for r > 0 {
		x0, y0 := e.TgtX, e.TgtY
		x1 := int(int16(e.TgtX + (e.TgtX-cx)*dirSign*int(r)))
		y1 := int(int16(e.TgtY + (e.TgtY-cy)*dirSign*int(r)))
		l := goldbox.NewLine(x0, y0, x1, y1)
		for {
			lastX, lastY := byte(l.X), byte(l.Y)
			if !(x0 == x1 && y0 == y1) {
				for {
					more := l.Step()
					who, ground = e.MapAt(int(byte(l.X)), int(byte(l.Y)))
					if more && step != nil {
						step(int(byte(l.X)), int(byte(l.Y)))
					}
					if !more {
						break
					}
					if who > 0 && who != prevWho {
						break
					}
					if ground == 0 || e.T().Ground[ground][0] > 1 {
						break
					}
					if byte(l.Steps) >= r {
						break
					}
				}
			}
			prevWho = who
			if ground == 0 {
				r = 0
			}
			blocked := strike(int(lastX), int(lastY), int(byte(l.X)), int(byte(l.Y)))
			if blocked {
				e.TgtX, e.TgtY = int(byte(l.X)), int(byte(l.Y))
				back := goldbox.NewLine(e.TgtX, e.TgtY, cx, cy)
				for back.Step() {
				}
				if flag != 0 && byte(back.Steps) <= 8 {
					l.Steps += 8
				}
				dirSign = -dirSign
				flag = 0
				prevWho = 0
			}
			if byte(l.Steps) < r {
				r -= byte(l.Steps)
			} else {
				r = 0
			}
			if blocked || r == 0 {
				break
			}
		}
	}
}

// fxLightning = ovr23:2BB7: level d6 down the line, bouncing.
func (e *RuleState) fxLightning() {
	d := byte(e.Dice(e.casterLevel(int(e.CurSpell)), 6))
	e.strikeSquare(e.TgtX, e.TgtY, int(d), 4)
	e.boltPath(8, int(d), 4, 1, SpellIDs[e.CurSpell])
}

// fxSlow = ovr23:2C09.
func (e *RuleState) fxSlow() {
	e.removeAffectFromSide(AffHaste, B2i(IsParty(e.Acting())), "is Slowed")
}

// fxSpeedy = ovr23:2E02.
func (e *RuleState) fxSpeedy() {
	if !e.CureAffect(e.spellTarget, AffSlow) {
		e.applySpell(int(e.CurSpell), 0, 0, 0, 0, "is Speedy")
	}
}

// fxCureParalysis = ovr23:2E4D: affect 0x37 is cured (and the affect 0x16
// that goes with it removed); otherwise a small heal.
func (e *RuleState) fxCureParalysis() {
	t := e.spellTarget
	if e.CureAffect(t, AffPoisoned) {
		e.Rv.Cure = 1
		e.RemoveAffect(t, affSlowPoison, nil)
		e.Rv.Cure = 0
		return
	}
	if e.cureDiseaseAffects() {
		return
	}
	if e.Heal(t, e.Dice(1, 4)+8, false) {
		e.hitMessage(t, true, "is healed")
	}
}

// fxStronger = ovr23:2EE5: strength 0x15 for the spell's duration.
func (e *RuleState) fxStronger() {
	t := e.spellTarget
	out, ok := e.SetStrength(t, 0x15, 0)
	if ok {
		e.CombatSay(t, "is stronger", true)
	}
	e.AddAffect(t, affStrength, e.spellAmount(int(e.CurSpell)), out, 1)
}

// fxBolt = ovr23:2F4D: Dice(1,6)+0x14 on the target square, then a fixed
// 0x14 along the line.
func (e *RuleState) fxBolt() {
	e.strikeSquare(e.TgtX, e.TgtY, e.Dice(1, 6)+0x14, 4)
	e.boltPath(3, 0x14, 4, 0, SpellIDs[e.CurSpell])
}

// fxHeal2d4 = ovr23:2FD0.
func (e *RuleState) fxHeal2d4() {
	t := e.spellTarget
	if e.Heal(t, e.Dice(2, 4)+2, false) {
		e.hitMessage(t, true, "is healed")
	}
}

// fxDamage2d4 = ovr23:3058: 2d4+2 of damage type 8.
func (e *RuleState) fxDamage2d4() {
	e.applySpell(int(e.CurSpell), 0, 0, e.RollDice(2, 4)+2, 8, "")
}

// fxNone = ovr23:3092: does nothing.
func (e *RuleState) fxNone() {}

// BreathWeapon = ovr23:30D9 (char, affect, flag): the breath weapon affect
// (type 0x58). Half the time (when DS:6DE3 is set) nothing happens; the
// monster breathes at the target square with its hit points as damage.
func (e *RuleState) BreathWeapon(c *Character, a *Affect, flag int) {
	if e.CombatRound != 0 && e.Dice(1, 100) <= 0x32 {
		return
	}
	cx, cy := e.CombatX(c), e.CombatY(c)
	e.Emit(core.Sounded{Cue: "breathe_fire"})
	e.CombatSay(c, "Breathes!", true)
	e.spellTargetHook(SpLightningBolt, 1, CastReply{Fight: true}) // aimed like a bolt; the result (DS:70C6) is not used
	e.TgtX = cx + goldbox.Sign(e.TgtX-cx)
	e.TgtY = cy + goldbox.Sign(e.TgtY-cy)
	if cx+1 == e.TgtX {
		e.TgtX++
	}
	if cy+1 == e.TgtY {
		e.TgtY++
	}
	e.removeInvisibility(c)
	e.shoot(cx, cy, e.TgtX, e.TgtY, 0, SpellBreathWeapon)
	e.strikeSquare(e.TgtX, e.TgtY, int(c.MaxHP), 3)
	e.boltPath(0xA, int(c.MaxHP), 3, 0, SpellBreathWeapon)
	if a.Data > 0xFD {
		a.Data--
	} else {
		e.RemoveAffect(c, AffBreathWeapon, a)
	}
	e.ClearCombatInfo(c)
}

// IsScroll = ovr23:323D (item): the base item type is a scroll (11..13).
func (e *RuleState) IsScroll(it *Item) bool {
	if it == nil {
		return false
	}
	t := e.Base(it.Type).Slot
	return t > 0xA && t < 0xE
}

// scrollUsed = ovr23:328A (char, item, spell): the spell is wiped from the
// scroll; the scroll goes when the item byte +0x30 falls under 0xD2.
func (e *RuleState) scrollUsed(c *Character, it *Item, spell SpellID) {
	s := spell.Spell()
	idx := 0
	for i := 1; i <= 3; i++ {
		if it.Spells[i-1].Spell() == s {
			idx = i
		}
	}
	if idx == 0 {
		return
	}
	it.Spells[idx-1] = 0
	it.Words[1]--
	if it.Words[1] < 0xD2 {
		e.RemoveItem(c, it)
	}
}

// announceCast = ovr23:331A (char, msg, spell): announces the cast, in the
// side panel in combat, else in the message box.
func (e *RuleState) announceCast(c *Character, msg string, spell int) {
	switch {
	case msg != "casts":
		e.Emit(core.Sounded{Cue: core.SoundCue("study")}) // memorized, scribed
	case e.BeginFlag != 5:
		e.Emit(core.Sounded{Cue: core.SoundCue("cast_spell")}) // in combat the spell's flight has the sound
	}
	msg = goldbox.Cut(msg, msgMax)
	if e.BeginFlag == 5 {
		e.CombatSay(c, "Casts a Spell", true)
		e.CombatSay(nil, "Spell: "+e.SpellName(spell), true)
		return
	}
	e.CombatSay(c, msg+" "+e.SpellName(spell), true) // extra: the original's flash in the message area is a toast, as CharMessage's
}

// SpellTableInit = ovr23:3413: the spell globals; the breath affect handler
// it sets (type 0x58, DS:6988) is in Tables.Affects; its DS:6B84 table of
// spell effects is spellEffect's switch.
func (e *RuleState) SpellTableInit() {
	e.Rv.Cure = 0 // DS:681F
	e.UsingItem = false
	e.SpellCaster = nil
}

// SpellEffect is spellEffect for the games' parity tests of their spells.
func (e *RuleState) SpellEffect(spell SpellID) { e.spellEffect(spell) }

// spellEffect runs the effect of spell: the DS:6B84 table of spell
// procedures spellTableInit filled, a switch on the spell here.
func (e *RuleState) spellEffect(spell SpellID) {
	switch spell {
	case SpBless:
		e.spellBless()
	case spCurse:
		e.spellBane()
	case SpCureLightWounds:
		e.spellCureLightWounds()
	case spCauseLightWounds:
		e.spellDamage1d8()
	case SpDetectMagicCl:
		e.affects(0, 0, "is affected") // ovr23:10E9
	case spProtFromEvilCl:
		e.affects(0, 0, "is protected") // ovr23:1123
	case spProtFromGoodCl:
		e.affects(0, 0, "is protected")
	case spResistCold:
		e.affects(0, 0, "is cold-resistant") // ovr23:1162
	case spBurningHands:
		e.spellDamageLevel()
	case spCharmPerson:
		e.spellCharmPerson()
	case spDetectMagicMU:
		e.affects(0, 0, "is affected")
	case spEnlarge:
		e.spellStrengthPercent()
	case spReduce:
		e.spellReduce()
	case spFriends:
		e.spellFriends()
	case SpMagicMissile:
		e.spellMagicMissile()
	case spProtFromEvilMU:
		e.affects(0, 0, "is protected")
	case spProtFromGoodMU:
		e.affects(0, 0, "is protected")
	case spReadMagic:
		e.affects(0, 0, "is affected")
	case spShield:
		e.affects(0, 0, "is shielded") // ovr23:1499
	case spShockingGrasp:
		e.spellBurningHands()
	case spSleep:
		e.spellSleep()
	case spFindTraps:
		e.affects(0, 0, "is affected")
	case SpHoldPersonCl:
		e.SpellHold()
	case spResistFire:
		e.affects(0, 0, "is fire resistant") // ovr23:17E9
	case spSilence:
		e.affects(0, 0, "is silenced") // ovr23:1822
	case spSlowPoison:
		e.spellRaiseDead()
	case spSnakeCharm:
		e.spellUndeadTargets()
	case spSpiritualHammer:
		e.spellApplyAnd17()
	case spDetectInvisibility:
		e.affects(0, 0, "is affected")
	case spInvisibility:
		e.affects(0, 0, "is invisible") // ovr23:1A16
	case SpKnock:
		e.affects(0, 0, "Knock-Knock") // ovr23:1A4F
	case spMirrorImage:
		e.spellMirrorImage()
	case spRayOfEnfeeblement:
		e.affects(0, 0, "is weakened") // ovr23:1ACC
	case spStinkingCloud:
		e.spellStinkingCloud()
	case spStrength:
		e.spellEnlarge()
	case spAnimateDead:
		e.spellAnimateDead()
	case spCureBlindness:
		e.spellCureBlind()
	case spCauseBlindness:
		e.affects(0, 0, "is blind") // ovr23:225D
	case spCureDisease:
		e.spellCureDisease()
	case spCauseDisease:
		e.affects(0, 1, "is diseased") // ovr23:234C
	case spDispelMagicCl:
		e.spellDispel()
	case spPrayer:
		e.spellPrayer()
	case spRemoveCurse:
		e.spellRemoveCurse()
	case spBestowCurse:
		e.affects(0, 0, "has been cursed!") // ovr23:2604
	case spBlink:
		e.affects(0, 0, "is blinking") // ovr23:263D
	case spDispelMagicMU:
		e.spellDispel()
	case spFireball:
		e.fxFireball()
	case spHaste:
		e.fxHaste()
	case spHoldPersonMU:
		e.SpellHold()
	case spInvisibility10:
		e.affects(0, 0, "is invisible")
	case SpLightningBolt:
		e.fxLightning()
	case spProtFromEvil10:
		e.affects(0, 0, "is protected")
	case spProtFromGood10:
		e.affects(0, 0, "is protected")
	case spProtFromNormalMissiles:
		e.affects(0, 0, "is protected")
	case spSlow:
		e.fxSlow()
	case 57:
		e.fxSpeedy()
	case 58:
		e.fxCureParalysis()
	case 59:
		e.fxStronger()
	case 60:
		e.fxBolt()
	case 61:
		e.affects(0, 0, "is paralyzed") // ovr23:2F99
	case 62:
		e.fxHeal2d4()
	case 63:
		e.affects(0, 0, "is invisible") // ovr23:302A
	case spFireballPower:
		e.fxFireball()
	case 65:
		e.fxDamage2d4()
	case 66:
		e.fxNone()
	case 67:
		e.affects(0xFF, 1, "is Reading") // ovr23:30A2
	default:
		if !e.spellEffectRuled(spell) && !e.spellEffectCotabFirst(spell) && !e.spellEffectSotsbFirst(spell) && !e.spellEffectPodFirst(spell) {
			panic(fmt.Sprintf("spellEffect: spell %d has no effect", spell))
		}
	}
}
