package rules1e

import "gobox/rules"

// Affect handlers of ovr13, offsets 0x0000..0x1C00. Each handler is a
// DS:6828 table entry: handler(char, affect, flag); flag 1 = removing.
// The rule globals DS:6816.. are e.rv (fleeRoll and friends see rules.go).
// DS:6826 (a far pointer scratch) is a Go local; where the original left
// it holding the target of the combat block, we use c.Combat.Target.

// combatInfo is c.Combat; outside a fight the original reads garbage, we
// read a blank block.
func combatInfo(c *Character) *CombatInfo {
	if c.Combat == nil {
		return &CombatInfo{}
	}
	return c.Combat
}

// targetOf is the character at char+0x108 -> +0xA.
func targetOf(c *Character) *Character {
	if t := combatInfo(c).Target; t != nil {
		return t
	}
	return &Character{}
}

// nameMatches: is the name one of names (the original's 16-byte Pascal strings)? Returns the number of matches (the originals count each).
func (e *RuleState) nameMatches(c *Character, names []string) int {
	m := 0
	for i := range names {
		if c.CharName == names[i] {
			m++
		}
	}
	return m
}

// blockAffect = ovr13:0000 (type): the affect being added is refused when
// type is 0 or the one being added; the damage goes with it.
func (e *RuleState) blockAffect(typ AffectType) {
	if typ == 0 || e.Rv.AffVal == typ {
		e.Rv.Damage = 0
		e.Rv.AffVal = 0
	}
}

// AddAffectNow = ovr13:001F (char, type, data, time): adds the affect with
// the remove-call flag, false while the temple is curing.
func (e *RuleState) AddAffectNow(c *Character, typ AffectType, data byte, time int) bool {
	if e.Rv.Cure != 0 {
		return false
	}
	e.AddAffect(c, typ, time, data, 1)
	return true
}

// killWithMessage = ovr13:0058 (char, status, message): the message, then
// the character goes down with the status unless already dead.
const msgPoisonDeath = "dies from poison"

func (e *RuleState) killWithMessage(c *Character, status byte, msg string) {
	msg = rules.Cut(msg, msgMax)
	e.CombatSay(c, msg, true)
	if s := c.Status; s == StatusStoned || s == StatusDead || s == StatusGone {
		return
	}
	c.Status = status
	c.Okay = 0
	c.HP = 0
	e.removeCombatAffects(c)
	e.AffectGroup(c, grpDown)
	if c.Okay == 0 {
		e.combatRemoveFromMap(c)
	}
}

// readiedWeapon = ovr13:101D (char): the weapon found by 00BA:2E8B, else
// the first readied item.
func (e *RuleState) readiedWeapon(c *Character) *Item {
	var it *Item
	if c.Slots[SlotWeapon] != nil {
		ok := e.AmmoFor(c, &it)
		if !ok || it == nil {
			it = c.Slots[SlotWeapon]
		}
	}
	return it
}

// avoidDistantAttack = ovr13:0FAC (char, percent): the selected character
// holding a ranged weapon may avoid an attack from afar.
func (e *RuleState) avoidDistantAttack(c *Character, pct byte) {
	sel := e.Acting()
	if sel.Slots[SlotWeapon] != nil && e.RangeTo(sel, c) > 1 && e.Dice(1, 100) <= int(pct) {
		e.CombatSay(c, "Avoids it", true)
		e.Rv.Damage = 0
		e.Rv.HitRoll = -1
		e.slotHits[1]--
	}
}

// poisonAttack = ovr13:1545 (char, bonus): the target of the combat block
// is poisoned unless it saves; no save means death.
func (e *RuleState) poisonAttack(c *Character, bonus int) {
	t := targetOf(c)
	// Felix's ruling, all games (PoD's): through AddAffectMsg, so an immunity blocks it
	if e.SavingThrow(t, SaveVsParalysis, bonus) {
		return
	}
	e.AddAffectMsg(t, AffPoisoned, 0, 0xFF, 0, saveNegate, false, "is poisoned")
	if e.FindAffect(t, AffPoisoned) != nil {
		e.killWithMessage(t, StatusDead, msgPoisonDeath) // extra: the original says "is killed"
	}
}

// paralyseTarget = ovr13:15EB (char, time, extra): the target is paralysed
// unless it saves. extra (bp+6) is never used by the original.
func (e *RuleState) paralyseTarget(c *Character, time int, extra byte) {
	t := targetOf(c)
	if !e.SavingThrow(t, SaveVsParalysis, 0) {
		e.hitMessage(t, true, "is Paralyzed")
		e.AddAffect(t, AffHeld, time, 0xC, 0)
	}
}

// hitBonusUp = ovr13:010D: to-hit and DS:6825 up.
func (e *RuleState) hitBonusUp(c *Character, a *Affect, flag byte) {
	e.fleeRoll += 5
	e.Rv.HitRoll++
}

// hitBonusDown = ovr13:011D: the reverse.
func (e *RuleState) hitBonusDown(c *Character, a *Affect, flag byte) {
	if e.fleeRoll < 5 {
		e.fleeRoll = 0
	} else {
		e.fleeRoll -= 5
	}
	e.Rv.HitRoll--
}

// bonusVsRace4 = ovr13:013B: +2 to hit and damage against a target of race 4.
func (e *RuleState) bonusVsRace4(c *Character, a *Affect, flag byte) {
	if targetOf(c).Kind == 4 {
		e.Rv.HitRoll += 2
		e.Rv.Damage += 2
	}
}

// startsToTrain = ovr13:017F: "starts to train".
func (e *RuleState) startsToTrain(c *Character, a *Affect, flag byte) {
	e.CombatSay(c, "starts to train", true)
	e.AddAffect(c, affTrainingManual, 0xA8C0, 0xFF, 1)
}

// bonusByTargetRace = ovr13:01C3: bonus by the target's race, fire-ish magic damage.
func (e *RuleState) bonusByTargetRace(c *Character, a *Affect, flag byte) {
	var k byte
	switch targetOf(c).Kind {
	case 0xA:
		k = 1
	case 9, 0xC:
		k = 2
	case 4:
		k = 3
	}
	e.Rv.HitRoll += int8(k)
	e.Rv.Damage += int16(k)
	e.Rv.DmgType = DmgFire | DmgMagic
}

// IsHardier = ovr13:0233: "is hardier" (the trained Constitution grows). From 15 on, the hit points above the dice
// divided by the constitution bonus summed over the class levels (at least 1) are gained when the constitution is
// below 17 or c has fighter levels (por-re EffTrainManual).
func (e *RuleState) IsHardier(c *Character, a *Affect, flag byte) {
	e.CombatSay(c, "is hardier", true)
	c.Stats[StatCon]++
	con := c.Stats[StatCon]
	if con >= 0x14 {
		e.AddAffect(c, affRegenerate, 0x3C, conRegenData, 1)
		return
	}
	gain := c.MaxHP - c.HPRoll
	if con < 15 {
		return
	}
	var bonus byte
	for i, l := range c.Levels {
		if int8(l) <= 0 {
			continue
		}
		switch {
		case i == ClsFighter:
			bonus += byte(con-15) * byte(l)
		case con > 15:
			bonus += byte(l) * 2
		default:
			bonus += byte(l)
		}
	}
	gain /= max(bonus, 1)
	if con < 17 || int8(c.Levels[ClsFighter]) > 0 {
		c.MaxHP += gain
		c.HP += gain
	}
}

// protFromEvil = ovr13:0377: against an evil attacker (the selected
// character's alignment 2/5/8, alignmentNames), save +2, its to-hit -2.
func (e *RuleState) protFromEvil(c *Character, a *Affect, flag byte) {
	if v := e.Acting().Alignment; v == 2 || v == 5 || v == 8 {
		e.Rv.Save += 2
		e.Rv.HitRoll -= 2
	}
}

// protFromGood = ovr13:03AC: the same against a good one (0/3/6).
func (e *RuleState) protFromGood(c *Character, a *Affect, flag byte) {
	if v := e.Acting().Alignment; v == 0 || v == 3 || v == 6 {
		e.Rv.Save += 2
		e.Rv.HitRoll -= 2
	}
}

// resistColdHalf = ovr13:03E1: cold damage halved, save +3.
func (e *RuleState) resistColdHalf(c *Character, a *Affect, flag byte) {
	if e.Rv.DmgType&dmgCold != 0 {
		e.Rv.Damage >>= 1
		e.Rv.Save += 3
	}
}

// charmSide = ovr13:0400: charm: the side follows the affect.
func (e *RuleState) charmSide(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		c.Side = (a.Data & 0x40) >> 6
		if c.Control == 0xB3 {
			c.Control = 0
			c.Auto = 0 // as PoD's O07C charm (T4d: one handler)
		}
		return
	}
	if a.Data&0x20 == 0 {
		a.Data = a.Data + 0x20 + c.Side<<6
		c.Side = a.Data >> 7
		c.Auto = 1
		if c.Control <= 0x7F {
			c.Control = 0xB3
		}
		combatInfo(c).Target = nil
		e.CountSides()
	}
	e.fleeRoll = 0x64
}

// bestStrength = ovr13:04BA: strength affects (the best one counts).
func (e *RuleState) bestStrength(c *Character, a *Affect, flag byte) {
	if a.Data > 0x7F {
		return
	}
	c.Stats[StatStr], c.Str00 = strDecode(a)
	var best *Affect
	var bs Score
	var bs00 byte
	for _, n := range c.Affects {
		if (n.Type == affEnlarged || n.Type == affStrength) && n != a {
			s, s00 := strDecode(n)
			if s > bs || bs == 0x12 && s00 > bs00 {
				bs, bs00, best = s, s00, n
			}
		}
	}
	if best != nil {
		best.Data = strValue(c.Stats[StatStr], c.Str00)
		c.Stats[StatStr], c.Str00 = bs, bs00
	}
}

// setCharisma = ovr13:05D6.
func (e *RuleState) setCharisma(c *Character, a *Affect, flag byte) { c.Stats[StatCha] = Score(a.Data) }

// poisonDamage = ovr13:05EB: poison: damage each round.
func (e *RuleState) poisonDamage(c *Character, a *Affect, flag byte) {
	if e.AddAffectNow(c, affPoisonDamage, a.Data, 0xA) && c.HP > 1 {
		e.Rv.DmgType = 0
		e.Damage(c, 1, 0, false)
	}
}

// armourAtLeast = ovr13:064E: armour class at least 0x39, save +1.
func (e *RuleState) armourAtLeast(c *Character, a *Affect, flag byte) {
	if c.AC[0] < 0x39 {
		c.AC[0] = 0x39
	}
	e.Rv.Save++
	if e.CurSpell == SpMagicMissile {
		e.Rv.Damage = 0
	}
}

// hitBonusVsNames = ovr13:0679: +1 to hit per matching name in DS:0332 (race 1, class 1).
func (e *RuleState) hitBonusVsNames(c *Character, a *Affect, flag byte) {
	t := targetOf(c)
	if t.Kind == 1 && t.Size == 1 {
		e.Rv.HitRoll += int8(e.nameMatches(t, hitNames1[:]) + e.nameMatches(t, hitNamesExtra[:]))
	}
}

// resistFireHalf = ovr13:06E4: fire damage halved, save +3.
func (e *RuleState) resistFireHalf(c *Character, a *Affect, flag byte) {
	if e.Rv.DmgType&DmgFire != 0 {
		e.Rv.Damage >>= 1
		e.Rv.Save += 3
	}
}

// IsSilenced = ovr13:070F: "is silenced" (the spell being cast is lost).
// PoD's (O07C) also skips the removal and forgets the spell being cast (T4d:
// one handler).
func (e *RuleState) IsSilenced(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		return
	}
	ci := combatInfo(c)
	if ci.CanUse != 0 {
		e.CombatSay(c, "is silenced", true)
	}
	if ci.Spell > 0 {
		e.forgetSpell(c, SpellID(ci.Spell))
	}
	ci.Spell = 0
	ci.CanUse = 0
	ci.CanCast = 0
}

// poisonEnds = ovr13:0775: poison ends; death by poison when the poison was still in.
func (e *RuleState) poisonEnds(c *Character, a *Affect, flag byte) {
	if e.FindAffect(c, AffPoisoned) != nil { // Felix's ruling, all games (PoD's): unconscious
		e.killWithMessage(c, StatusUnconscious, "collapses from poison")
	}
	e.Rv.Cure = 1
	e.RemoveAffect(c, affPoisonDamage, nil)
	e.Rv.Cure = 0
}

// porHammer is POR's Spiritual Hammer (CotAB's too): a +1
// hammer named by word 0xF3 that ends the spell when put away.
var porHammer = Item{Type: ItemHammer, Words: [3]WordID{0, WordID(ItemHammer), 0xF3}, Plus: 1, Spells: [3]ItemPower{0, ItemPower(affSpiritualHammer), ItemPower(affEndsSpiritualHammer)}}

// SotsbHammer is SotSB's: POR's in SotSB's numbering, type 6 (Hammer) and
// name word 121 (sotsb-re gdconsts.pas itemNames; POR's type 0x14 is
// SotSB's Two-Handed Sword).
var SotsbHammer = Item{Type: 6, Words: [3]WordID{0, 6, 121}, Plus: 1, Spells: [3]ItemPower{0, ItemPower(affSpiritualHammer), ItemPower(affEndsSpiritualHammer)}}

// PodHammer is PoD's: its type 6 and name word 121, no ending power.
var PodHammer = Item{Type: 6, Words: [3]WordID{0, 6, 121}, Plus: 1, Spells: [3]ItemPower{0, ItemPower(affSpiritualHammer), ItemPower(AffItem0)}}

// hammer is the game's Spiritual Hammer (rules.json spiritualHammer).
func (e *RuleState) hammer() *Item {
	if h := e.T().Hammer; h != nil {
		return h
	}
	return &porHammer
}

// SpiritualHammerItem = ovr13:07E0 (PoD O07C:068B): the Spiritual Hammer comes and goes.
func (e *RuleState) SpiritualHammerItem(c *Character, a *Affect, flag byte) {
	e.spiritualHammer(c, flag)
}

// spiritualHammer adds or removes the hammer; it reports the hammer it added.
func (e *RuleState) spiritualHammer(c *Character, flag byte) bool {
	h := e.hammer()
	var it *Item
	for _, i := range e.Carried(c) {
		if i.Type == h.Type && i.Words[2] == h.Words[2] {
			it = i
			break
		}
	}
	if flag != 0 && it != nil {
		e.RemoveItem(c, it)
	}
	n := *h
	added := flag == 0 && it == nil && !e.Overloaded(&n, c) // a party member's room is the party inventory's
	if added {
		e.AddItemCopy(c, &n)
		e.CombatSay(c, "Gains an item", true)
	}
	e.ApplyItems(c)
	return added
}

// spiritualHammerPod = PoD O07C:068B: as POR's, then it readies the item
// listed after the hammer when no weapon is readied (its search steps past
// the find), not the hammer.
func (e *RuleState) spiritualHammerPod(c *Character, a *Affect, flag byte) {
	if !e.spiritualHammer(c, flag) || c.Slots[0] != nil {
		return
	}
	h := e.hammer()
	items := e.Carried(c)
	for i, it := range items {
		if it.Type == h.Type && it.Words[2] == h.Words[2] {
			if i+1 < len(items) {
				saved := e.Who
				e.Who = e.PartyIndex(c)
				e.ReadyItem(items[i+1])
				e.Who = saved
			}
			break
		}
	}
}

// blindedToHit = ovr13:0918: to-hit -4; DS:681E unless the selected one has affect 0x18.
func (e *RuleState) blindedToHit(c *Character, a *Affect, flag byte) {
	if e.FindAffect(e.Acting(), affDetectInvisible) == nil {
		e.AffectVeto = 1
	}
	e.Rv.HitRoll -= 4
}

// hitBonusVsTargetNames = ovr13:0947: +1 to hit per matching name in DS:0372 (target: race 1, class 1).
func (e *RuleState) hitBonusVsTargetNames(c *Character, a *Affect, flag byte) {
	t := targetOf(c)
	if t.Kind == 1 && t.Size == 1 {
		e.Rv.HitRoll += int8(e.nameMatches(t, hitNames2[:]) + e.nameMatches(t, hitNamesExtra[:]))
	}
}

// Mirror image data: the images in the high nibble (the caster's level, or
// nothing, below); a spell record's field spellImageField non-zero: an image
// can take that spell.
const (
	imageStep       = 16
	imageLow        = 15
	spellImageField = 14
)

// mirrorImage = SotSB O0A4:0A83 (Felix's ruling, all games, POR's ovr13:09C0
// replaced): a blow or a single-target spell may strike an image instead.
func (e *RuleState) mirrorImage(c *Character, a *Affect, flag byte) {
	if flag != 0 || e.Dice(1, int(a.Data/imageStep)+1) <= 1 {
		return
	}
	if e.CurSpell > 0 && (e.areaSpell != 0 || e.T().Spells[e.CurSpell][spellImageField] == 0) {
		return
	}
	e.blockAffect(affAny)
	e.CombatSay(c, "lost an image", true)
	images := a.Data/imageStep - 1
	if images == 0 {
		e.RemoveAffect(c, affMirrorImage, nil)
		return
	}
	a.Data = a.Data&imageLow | images*imageStep
}

// damageLessQuarter = ovr13:0A3D: damage down by a quarter.
func (e *RuleState) damageLessQuarter(c *Character, a *Affect, flag byte) {
	e.Rv.Damage -= e.Rv.Damage >> 2
}

// CoughingEnds = PoD O07C:0A11 (Felix's ruling, all games; POR's ovr13:0A65
// replaced): in a fight no spells or items, and both armour classes 2 worse.
func (e *RuleState) CoughingEnds(c *Character, a *Affect, flag byte) {
	if e.BeginFlag != beginFlagCombat {
		return
	}
	ci := combatInfo(c)
	if ci.CanUse != 0 {
		e.CombatSay(c, "is coughing", true)
	}
	ci.CanUse = 0
	ci.CanCast = 0
	e.ApplyItems(c)
	c.AC[0] -= 2
	c.AC[1] -= 2
}

// collapses = ovr13:0B1B: "collapses" (status 6) and the character is reset.
func (e *RuleState) collapses(c *Character, a *Affect, flag byte) {
	a.Call = 0
	if e.Rv.Cure == 0 {
		e.killWithMessage(c, StatusDead, "collapses")
	}
	c.Side = a.Data >> 4
	c.Auto = 1
	c.Undead = 0
	c.FightLevel = byte(max(c.Levels[ClsFighter], c.Levels[ClsPaladin], c.Levels[ClsRanger]))
	c.BaseMove = 0xC
	if c.Control == 0xB3 {
		c.Control = 0
	}
	c.Kind = 0
}

// minus4All = ovr13:0BAC: -4 to hit, armour and save.
func (e *RuleState) minus4All(c *Character, a *Affect, flag byte) {
	e.Rv.HitRoll -= 4
	c.AC[0] -= 4
	c.AC[1] -= 4
	e.Rv.Save -= 4
}

// weakAndSick = ovr13:0BCF: affects 0x2B and 0x2C together.
func (e *RuleState) weakAndSick(c *Character, a *Affect, flag byte) {
	e.callAffect(affWeakened, c, a, flag)
	e.callAffect(affSick, c, a, flag)
}

// plusOneSaveHit = ovr13:0C06: +1 save and to-hit.
func (e *RuleState) plusOneSaveHit(c *Character, a *Affect, flag byte) {
	e.Rv.Save++
	e.Rv.HitRoll++
}

// minus4HitSave = ovr13:0C15: -4 to hit and save.
func (e *RuleState) minus4HitSave(c *Character, a *Affect, flag byte) {
	e.Rv.HitRoll -= 4
	e.Rv.Save -= 4
}

// alwaysMissDelayed = ovr13:0C26: DS:681E and the attack always misses.
func (e *RuleState) alwaysMissDelayed(c *Character, a *Affect, flag byte) {
	if int8(combatInfo(c).Delay) > 0 {
		e.AffectVeto = 1
		e.Rv.HitRoll = -1
	}
}

// agesOnce = ovr13:0C4B: "ages" once (age +1), DS:681A doubled.
func (e *RuleState) agesOnce(c *Character, a *Affect, flag byte) { e.agesOnceBit(c, a, hasteAgedBit) }

// The haste data bit that says the hasted aged already (PoD keeps it one
// higher: hasteAgedBitPod).
const (
	hasteAgedBit    = 0x10
	hasteAgedBitPod = 0x20
)

// agesOnceBit is agesOnce with the game's aged bit.
func (e *RuleState) agesOnceBit(c *Character, a *Affect, bit byte) {
	if a.Data&bit == 0 {
		a.Data += bit
		e.CombatSay(c, "ages", true)
		c.Age++
		e.ageCheckDue = true
	}
	e.turnValue <<= 1
}

// avoidUnmagicWeapon = ovr13:1086: the selected character with a weapon of strength 0 may avoid
// the attack.
func (e *RuleState) avoidUnmagicWeapon(c *Character, a *Affect, flag byte) {
	if w := e.readiedWeapon(e.Acting()); w != nil && w.Plus == 0 {
		e.avoidDistantAttack(c, 0x64)
	}
}

// halveAttacks = ovr13:10C1: DS:681A halved.
func (e *RuleState) halveAttacks(c *Character, a *Affect, flag byte) { e.turnValue >>= 1 }

// isWeakened = ovr13:10DE: "is weakened": strength down while it lasts (affect 0x2B).
func (e *RuleState) isWeakened(c *Character, a *Affect, flag byte) {
	if !e.AddAffectNow(c, affWeakened, a.Data, 0x3C) {
		return
	}
	if c.Stats[StatStr] > 3 {
		e.CombatSay(c, "is weakened", true)
		c.Stats[StatStr]--
	} else if e.FindAffect(c, affHelpless) == nil {
		e.AddAffect(c, affHelpless, 0, 0xFF, 0)
	}
}

// sickDamage = ovr13:1168: affect 0x2C: damage while it lasts, or nausea at 1 hit point.
func (e *RuleState) sickDamage(c *Character, a *Affect, flag byte) {
	if !e.AddAffectNow(c, affSick, a.Data, 0xA) {
		return
	}
	if c.HP > 1 {
		e.Rv.DmgType = 0
		e.Damage(c, 1, 0, false)
	} else if e.FindAffect(c, affHelpless) == nil {
		e.AddAffect(c, affHelpless, 0, 0xFF, 0)
	}
}

// minus4VsNames = ovr13:11F9: -4 to hit against the names in DS:03F2 (selected: race 1, race bits 2).
func (e *RuleState) minus4VsNames(c *Character, a *Affect, flag byte) {
	// fixed, as the Pascal reference: the original matched the defender's own target and only for an
	// attacker of race 1 and size 2, so of the giants only ogres counted
	e.Rv.HitRoll -= int8(4 * e.nameMatches(e.Acting(), hitNames3[:]))
}

// minus4VsBugbearGnoll = ovr13:1276: -4 to hit when the selected is a BUGBEAR or GNOLL of race 1.
func (e *RuleState) minus4VsBugbearGnoll(c *Character, a *Affect, flag byte) {
	sel := e.Acting()
	if sel.Kind == 1 &&
		(sel.CharName == "BUGBEAR" || sel.CharName == "GNOLL") {
		e.Rv.HitRoll -= 4
	}
}

// prayerSide = ovr13:12B2: same side as the affect: plusOneSaveHit, else -1 to hit and save.
func (e *RuleState) prayerSide(c *Character, a *Affect, flag byte) {
	if c.Side == (a.Data&0x10)>>4 {
		e.plusOneSaveHit(c, a, flag)
	} else {
		e.Rv.HitRoll--
		e.Rv.Save--
	}
}

// healOverTime = ovr13:12FB: affect 0x32 and a point of healing.
func (e *RuleState) healOverTime(c *Character, a *Affect, flag byte) {
	if e.AddAffectNow(c, affMummyDisease, a.Data, 0x3840) && e.Heal(c, 1, true) {
		e.healedMessage(c)
	}
}

// clearCombatBlock = ovr13:133F: the combat block is cleared; the one
// lose-turn handler (sotsb O0A4:0274, pod O07C:00F5 endTurn are the same).
func (e *RuleState) clearCombatBlock(c *Character, a *Affect, flag byte) { e.ClearCombatInfo(c) }

// turnInvisible = ovr13:1363.
// PoD's ring (O07C) also takes it away when put off in a fight (T4d: one
// handler).
func (e *RuleState) turnInvisible(c *Character, a *Affect, flag byte) {
	if flag != 0 {
		if e.inCombat() {
			e.RemoveAffect(c, affInvisible, nil)
		}
		return
	}
	e.AddAffect(c, affInvisible, 1, 0xC, 0)
}

// rots = ovr13:1399: "Rots...": charisma drops, then the affect or death.
func (e *RuleState) rots(c *Character, a *Affect, flag byte) {
	if e.Rv.Cure != 0 {
		return
	}
	e.CombatSay(c, "Rots...", true)
	if c.Stats[StatCha] > 2 {
		c.Stats[StatCha] -= 2
	} else {
		c.Stats[StatCha] = 0
	}
	if c.Stats[StatCha] < 3 {
		c.Stats[StatCha] = 3
	}
	if a.Data>>4 > 1 {
		a.Data -= 0x10
		e.AddAffectNow(c, affRotting, a.Data, 0xA8C0)
	} else {
		e.killWithMessage(c, StatusGone, "Dies rots away")
	}
}

// heldNoMoves = ovr13:1452: the combat block loses +6; DS:681A cleared when DS:681D is set.
func (e *RuleState) heldNoMoves(c *Character, a *Affect, flag byte) {
	combatInfo(c).Moves = 0
	if e.modMoves != 0 {
		e.turnValue = 0
	}
}

// addAffect62 = ovr13:1472.
func (e *RuleState) addAffect62(c *Character, a *Affect, flag byte) {
	e.AddAffect(c, affHeal3, 0, 0xFF, 0)
}

// ResistFirePerDie = ovr13:1490: fire damage is cut by 2 per die (not below one per die), save +4.
func (e *RuleState) ResistFirePerDie(c *Character, a *Affect, flag byte) {
	if e.Rv.DmgType&DmgFire == 0 {
		return
	}
	for i := byte(1); i <= e.Rv.DiceN; i++ {
		// fixed, as the Pascal reference: in bytes, 0 or 1 point minus 2 wrapped to ~250 and stood above the floor
		e.Rv.Damage = int16(max(int(e.Rv.Damage)-2, int(e.Rv.DiceN)))
	}
	e.Rv.Save += 4
	if e.Rv.DmgType&DmgMagic == 0 {
		e.blockAffect(affAny)
	}
}

// regenerate = ovr13:14EB: affect 0x3E and a point of healing.
func (e *RuleState) regenerate(c *Character, a *Affect, flag byte) {
	if e.AddAffectNow(c, affRegenerate, a.Data, 0x3C) && e.Heal(c, 1, true) {
		e.healedMessage(c)
	}
}

// poisonBite = ovr13:165D, 1671, 1685: poison bites with save bonus 0, 4, 2.
func (e *RuleState) poisonBite(c *Character, a *Affect, flag byte)  { e.poisonAttack(c, 0) }
func (e *RuleState) poisonBite4(c *Character, a *Affect, flag byte) { e.poisonAttack(c, 4) }
func (e *RuleState) poisonBite2(c *Character, a *Affect, flag byte) { e.poisonAttack(c, 2) }

// paralyse2d8 = ovr13:1699: paralysis for 2d8.
// Felix's ruling, all games: 2d8 rounds, a foe close by (PoD's reach), held
// through AddAffectMsg (immunities block it, a hold already there is renewed
// by the renewal rule).
func (e *RuleState) paralyse2d8(c *Character, a *Affect, flag byte) {
	t := combatInfo(c).Target
	if flag != 0 || t == nil || e.RangeTo(c, t) >= touchReach {
		return
	}
	time := e.Dice(2, 8)
	saved := e.SavingThrow(t, SaveVsParalysis, 0)
	e.AddAffectMsg(t, AffHeld, time, heldData, 0, saveNegate, saved, "paralyzed by poison")
}

// touchReach: a touch reaches less than this; heldData is a hold's data.
const (
	touchReach = 2
	heldData   = 0xC
)

// paralyseNotElf = ovr13:16BB: paralysis for 0x3F unless the target is an elf.
func (e *RuleState) paralyseNotElf(c *Character, a *Affect, flag byte) {
	if targetOf(c).Race != RaceElf {
		e.paralyseTarget(c, 0x3F, 0)
	}
}

// paralyse1d9 = ovr13:16E6: paralysis for 1d9+10.
func (e *RuleState) paralyse1d9(c *Character, a *Affect, flag byte) {
	e.paralyseTarget(c, e.Dice(1, 9)+10, 0xFE)
}

// poisonBiteMinus2 = ovr13:170B: poison with save bonus 0xFE (-2).
func (e *RuleState) poisonBiteMinus2(c *Character, a *Affect, flag byte) { e.poisonAttack(c, 0xFE) }

// missMinus4 = ovr13:171F: DS:681E, to-hit -4.
func (e *RuleState) missMinus4(c *Character, a *Affect, flag byte) {
	e.AffectVeto = 1
	e.Rv.HitRoll -= 4
}

// mostlyInvisible = ovr13:1730: invisibility 95% of the time.
func (e *RuleState) mostlyInvisible(c *Character, a *Affect, flag byte) {
	if e.Dice(1, 100) <= 0x5F {
		e.AddAffect(c, affInvisible, 1, 0xC, 0)
	}
}

// bloodDrain = ovr13:176F: blood drain (affect 0x4B on the drinker, 0x3A on the victim,
// the victim is combatant a.Raw[3]).
func (e *RuleState) bloodDrain(c *Character, a *Affect, flag byte) {
	t := e.Combatant(a.Data)
	if t == nil {
		t = &Character{}
	}
	if flag != 0 || c.Okay == 0 || t.Okay == 0 {
		e.RemoveAffect(t, affHeldFast, nil)
		if flag == 0 {
			a.Call = 0
			e.RemoveAffect(c, affBloodDrain, a)
		}
		if c.Okay != 0 && t.Okay != 0 {
			combatInfo(c).Turned = 1
		}
		return
	}
	e.CombatSay(c, "Sucks some Blood", true)
	dmg := e.RollDice(1, 4)
	e.Rv.DmgType = 0
	e.Damage(t, dmg, 0, false)
	e.ClearCombatInfo(c)
	time := int(a.Time)
	if t.Okay != 0 && time>>4 > dmg {
		time -= dmg << 4
		a.Time = uint16(time)
	} else {
		e.RemoveAffect(c, affBloodDrain, a)
	}
}

// holdTarget = ovr13:18B5: the target is held (affect 0x4B on c with the target's index).
func (e *RuleState) holdTarget(c *Character, a *Affect, flag byte) {
	t := targetOf(c)
	e.AddAffect(c, affBloodDrain, 0xCF, byte(e.CombatIndex(t)), 1)
	e.AddAffect(t, affHeldFast, 0, 0xFF, 0)
	e.callAffect(affHeldFast, t, nil, 0)
}

// holdWeakTarget = ovr13:1935: hold: a target under 0x13 strength is held fast.
func (e *RuleState) holdWeakTarget(c *Character, a *Affect, flag byte) {
	t := targetOf(c)
	if t.Stats[StatStr] < 0x13 {
		e.AddAffect(c, affStrikeTied, 0, byte(e.CombatIndex(t)), 1)
		e.AddAffect(t, affHeldFast, 0, 0xFF, 0)
		e.callAffect(affHeldFast, t, nil, 0)
		e.CombatSay(t, "is held fast", true)
	}
}

// getUpOrWait = ovr13:19D9: back on its feet or affect 0x4E (the local result is unused).
func (e *RuleState) getUpOrWait(c *Character, a *Affect, flag byte) {
	if !e.Revive(c, c.HP) {
		e.AddAffectNow(c, affGetUp, a.Data, 1)
	}
}

// magicDamage2d10 = ovr13:1A19: 2d10 magic damage to the target.
func (e *RuleState) magicDamage2d10(c *Character, a *Affect, flag byte) {
	e.Rv.DmgType = DmgFire | DmgMagic
	t := targetOf(c)
	e.Damage(t, e.RollDice(2, 10), 0, false)
}

// acidDamage1d4 = ovr13:1A4C: 1d4 acid damage to the target.
func (e *RuleState) acidDamage1d4(c *Character, a *Affect, flag byte) {
	e.Rv.DmgType = dmgAcid
	t := targetOf(c)
	e.Damage(t, e.RollDice(1, 4), 0, false)
}

// fear = ovr13:1AAA: fear: the enemies save or are paralysed / weakened (by +0x73).
func (e *RuleState) fear(c *Character, a *Affect, flag byte) {
	for _, p := range e.Party {
		if p.Okay == 0 || byte(B2i(IsParty(c))) != p.Side {
			continue
		}
		saved := e.SavingThrow(p, SaveVsSpell, 0)
		switch v := int8(p.MaxLevel); {
		case v >= 0 && v <= 3:
			e.AddAffectMsg(p, AffHeld, 0, 0xC, 0, 1, saved, "is paralyzed with fear")
		case v == 4 || v == 5:
			e.AddAffectMsg(p, AffCurse, 0, 0xC, 0, 1, saved, "is weakened by fear")
		}
	}
}

// fearLifted = ovr13:1BA8: the fear (0x52) is lifted from everyone and the enemies are
// paralysed for 1d4 rounds.
func (e *RuleState) fearLifted(c *Character, a *Affect, flag byte) {
	for _, p := range e.Party {
		e.RemoveAffect(p, affFearLifted, nil)
		if p.Okay == 0 || byte(B2i(IsParty(c))) != p.Side {
			continue
		}
		var bonus int
		if byte(e.Rv.Side[p.Side]) >= 6 {
			bonus++
		}
		if p.Race == RaceHuman {
			bonus += 2
		}
		saved := e.SavingThrow(p, SaveVsParalysis, bonus)
		e.AddAffectMsg(p, AffHeld, e.Dice(1, 4), 0xC, 0, 1, saved, "is paralyzed with fear")
	}
}
