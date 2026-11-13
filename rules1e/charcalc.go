package rules1e

import (
	"gobox/core"
	"gobox/core/goldbox"
)

// Derived character values, ported from ovr24.

// RecomputeCharacter = ovr24:0000: hit dice, highest level, attacks,
// spell slots, known cleric spells, saving throws and thief skills from
// the class levels.
func (e *RuleState) RecomputeCharacter(c *Character) {
	s := e.T().ClassStats(who(c))
	c.Thac0Base = byte(s.Thac0Base)
	if int(int8(c.MaxLevel)) < s.Highest {
		c.MaxLevel = byte(s.Highest)
	}
	c.BaseAttack[0] = Count(s.Attacks)
	c.SpellSlots[SpellClassCleric], c.SpellSlots[SpellClassMage], c.SpellSlots[SpellClassDruid] = countSlots(s.ClericSlots), countSlots(s.MageSlots), countSlots(s.DruidSlots)
	// a paladin with cleric slots knows them too (COTAB O0FE:000F knowClassSpells); a ranger's
	// druid spells are none in POR's spell list
	if int8(c.Levels[ClsCleric]) > 0 || rowSum(e.T().PaladinSpells, int(int8(c.Levels[ClsPaladin]))) != [5]int{} {
		for k := 1; k <= LastSpell; k++ {
			rec := e.T().Spells[k][:2]
			if SpellClass(rec[0]) == SpellClassCleric && c.SpellSlot(SpellClassCleric, int(int8(rec[1]))) > 0 {
				c.SpellBook[k-1] = 1
			}
		}
	}
	for i, save := range porSave {
		c.Saves[i] = Score(s.Saves[save])
	}
	if int8(c.Levels[ClsThief]) > 0 {
		for i := range s.Skills {
			c.Thief[i] = Percent(s.Skills[i])
		}
	}
}

// wisdomSlots = ovr24:01AF: bonus cleric spells for wisdom 13..17.
func (e *RuleState) wisdomSlots(c *Character) {
	c.SpellSlots[SpellClassCleric] = countSlots(WisdomBonus(intSlots(c.SpellSlots[SpellClassCleric]), int(c.Stats[StatWis])))
}

// who is what ClassStats reads from c.
func who(c *Character) Who {
	levels, old := map[core.Class]int{}, map[core.Class]int{}
	for i, class := range PorClass {
		levels[class], old[class] = int(int8(c.Levels[i])), int(int8(c.OldLevels[i]))
	}
	race := CoreRaceMonster // monster records carry other race numbers
	if int(c.Race) < len(RecordRace) {
		race = RecordRace[c.Race]
	}
	var powers ItemPowers
	for _, it := range c.Items {
		if it.Readied != 0 && it.Spells[2] == ItemPower(itemPowerWizardry) { // the later games' ring of wizardry (POR's items have none)
			powers.Wizardry = true
		}
		if it.Readied != 0 && it.Spells[2] == ItemPower(itemPowerConstitution) { // COTAB's constitution item: the dwarf's save bonus for anyone
			powers.SaveBonus = true
		}
	}
	return Who{
		Race: race, OldLevels: old, Powers: powers,
		Levels: levels, Dex: int(c.Stats[StatDex]), Wis: int(c.Stats[StatWis]), Con: int(c.Stats[StatCon]),
		Int:         int(c.Stats[StatInt]),
		ClericSlots: intSlots(c.SpellSlots[SpellClassCleric]),
		MageSlots:   intSlots(c.SpellSlots[SpellClassMage]),
		DruidSlots:  intSlots(c.SpellSlots[SpellClassDruid]),
	}
}

// savingThrows = ovr24:0255: the best of each class's table (DS:426E), with
// COTAB's constitution bonus (ClassStats).
func (e *RuleState) savingThrows(c *Character) {
	s := e.T().ClassStats(who(c))
	for i, save := range porSave {
		c.Saves[i] = Score(s.Saves[save])
	}
}

// thiefSkills = ovr24:0329: level table DS:3CF3 plus race DS:3D3B, plus
// dexterity DS:3D4E for the first five skills.
func (e *RuleState) thiefSkills(c *Character) {
	skills := e.T().ThiefSkills(int(int8(c.Levels[ClsThief])), RecordRace[c.Race], int(c.Stats[StatDex]), ItemPowers{})
	for i := range skills {
		c.Thief[i] = Percent(skills[i])
	}
}

// Readied-item slots at char+0xCC (13 far pointers). Slot numbers come
// from the base item record; 2 is body armour, 9 and 10 the two rings, 11
// arrows, 12 quarrels.
const (
	SlotWeapon = 0
	SlotShield = 1
	// ReadiedOffHand (extra) is the Readied of a weapon in the off hand: the shield slot
	ReadiedOffHand = 4
	SlotArmor      = 2
	slotRing1      = 9
	slotRing2      = 10
	SlotArrows     = 11
	SlotBolts      = 12
)

// ItemBase is the 16-byte base record of an item type (DS:5582 + 16*type,
// from the ITEMS file).
type ItemBase struct {
	Slot       byte // +0 readied slot (0 weapon ... 9 ring, 0xB..0xD scrolls)
	Hands      byte // +1
	LargeDice  byte // +2 damage against large targets
	LargeSides byte // +3
	LargeBonus byte // +4 signed
	Speed      byte // +5 rate of fire / weapon speed
	Protection byte // +6 the armour class it gives: acGiven, then its points (acPoints)
	Material   byte // +7 flags the resistance handlers test (0x81, bit 0)
	Kept8      byte // +8 no rule reads it (0x80 on most weapons); kept so the table round-trips
	Dice       byte // +9
	Sides      byte // +A
	Bonus      byte // +B signed
	Range      byte // +C squares + 1 (1 = melee)
	Classes    byte // +D class bits that may use it (ClassMask)
	Usage      byte // +E flags: 1 arrows, 2 missile, 4 strength bonus, 8 fires ammunition, 0x10 thrown, 0x80 bolts
	_          byte // +F
}

// Armour class is counted up, as 60 minus AC (Character.AC): AcTen is
// AC 10, nothing worn. An item's Protection has AcGiven set when it counts
// towards armour class at all, the low bits its points: body armour and
// bracers on that same scale (AcTen + how much better than AC 10), a
// shield the points it adds; rings and cloaks of protection have none,
// they give only their magic plus (itemAC).
const (
	AcGiven = 0x80
	AcTen   = 50
)

// givesAC: the item counts towards armour class.
func (b *ItemBase) givesAC() bool { return b.Protection&AcGiven != 0 }

// AcPoints is Protection's points, 0 for an item that gives none.
func (b *ItemBase) AcPoints() int {
	if !b.givesAC() {
		return 0
	}
	return int(b.Protection &^ AcGiven)
}

// AcBetter is how much better than AC 10 the item makes its wearer, magic
// aside: armour's and bracers' points less acTen, a shield's as they are.
func (b *ItemBase) AcBetter() int {
	if v := b.AcPoints(); v < AcTen {
		return v
	}
	return b.AcPoints() - AcTen
}

// Base is the Base record of an item type (a blank one past the table).
func (e *RuleState) Base(typ ItemType) *ItemBase {
	if int(typ) < len(e.Items) {
		return &e.Items[typ]
	}
	return &ItemBase{}
}

// readySlot is the slot (Character.Slots) a readied item of type typ
// fills: its base record's, but bracers (extra) wear in the armour's, which
// they stand in for (itemAC: only the better of the two counts).
func (e *RuleState) readySlot(typ ItemType) int {
	if e.isKind(typ, ItemIDBracers) {
		return SlotArmor
	}
	return int(e.Base(typ).Slot)
}

// ApplyItems = 00BA:0BB8: everything that follows from the items carried
// and readied: slots, weight, hands, attack and damage, armour class,
// movement and saving throw bonus.
func (e *RuleState) ApplyItems(c *Character) {
	if e.partyMember(c) { // extra: unreadied items belong to the party inventory
		e.stow(c)
	}
	c.Slots = [14]*Item{}
	c.ItemCount = 0
	c.Hands = 0
	c.Weight = 0 // extra: items weigh nothing
	for _, it := range c.Items {
		c.ItemCount++
		if it.Readied == 0 {
			continue
		}
		typ := it.Type
		if it.Readied == ReadiedOffHand && !e.Rules().OffHand(e, c, it) { // U9: no off hand under a rule set without one (1e): the weapon comes off
			it.Readied = 0
			continue
		}
		if it.Readied == ReadiedOffHand { // a weapon in the off hand, the shield's (a rule set's: OffHand; 4e only)
			c.Slots[SlotShield] = it
			c.Hands += e.Base(typ).Hands
			continue
		}
		switch s := e.readySlot(typ); {
		case s <= 8:
			c.Slots[s] = it
		case s == 9:
			if c.Slots[slotRing1] == nil {
				c.Slots[slotRing1] = it
			} else if c.Slots[slotRing2] == nil {
				c.Slots[slotRing2] = it
			}
		}
		if e.isKind(typ, ItemIDArrow) {
			c.Slots[SlotArrows] = it
		}
		if e.isKind(typ, ItemIDQuarrel) {
			c.Slots[SlotBolts] = it
		}
		c.Hands += e.Base(typ).Hands
	}
	c.Dice, c.Sides, c.Bonus = c.BaseDice, c.BaseSides, c.BaseBonus
	var ac [5]int8
	ringOnly := false
	c.SaveBonus = 0
	c.AC[0] = ArmourClass(c.BaseAC)
	c.Move = c.BaseMove
	c.THAC0 = c.Thac0Base
	ac[0] = goldbox.DexAC(byte(c.Stats[StatDex]))
	if c.Slots[SlotWeapon] == nil {
		c.THAC0 += byte(strHit(c))
		c.Bonus[0] += Modifier(strDamage(c))
	}
	e.weaponStats(c)
	for _, it := range c.Items {
		if it.Readied == 0 {
			continue
		}
		e.armourMove(it, c)
		if e.itemAC(c, it, &ac) {
			ringOnly = true
		}
	}
	if ringOnly {
		ac[3] = 0
	}
	if int(ac[4]) < int(c.AC[0]) {
		ac[4] = int8(c.AC[0])
	}
	if a := e.FindAffect(c, affBarkskin); a != nil {
		ac[4] = max(ac[4], barkskinBase(a))
	}
	c.AC[0] = 0
	for i := 0; i <= 4; i++ {
		c.AC[0] += ArmourClass(ac[i])
	}
	c.AC[1] = ArmourClass(int(ac[4]) + int(ac[2]) + int(ac[3]) - 2)
	if FighterKind(c) && int8(c.Race) > 0 {
		c.FightLevel = byte(max(c.Levels[ClsFighter], c.Levels[ClsPaladin], c.Levels[ClsRanger]))
	} else {
		c.FightLevel = 1
	}
}

// weaponStats = 00BA:0000: to-hit and damage of the readied weapon.
func (e *RuleState) weaponStats(c *Character) {
	w := c.Slots[SlotWeapon]
	if w == nil {
		return
	}
	typ := w.Type
	flags := e.Base(typ).Usage
	c.THAC0 = c.Thac0Base
	if flags&2 != 0 { // missile
		c.THAC0 += byte(goldbox.DexMissile(byte(c.Stats[StatDex])))
	}
	c.Bonus[0] = Modifier(e.Base(typ).Bonus)
	if flags&4 != 0 { // strength applies
		c.THAC0 += byte(strHit(c))
		c.Bonus[0] += Modifier(strDamage(c))
	}
	plus := int8(w.Plus)
	if flags > 0x7F && c.Slots[SlotBolts] != nil {
		plus += int8(c.Slots[SlotBolts].Plus)
	}
	if flags&1 != 0 && c.Slots[SlotArrows] != nil && e.isKind(c.Slots[SlotArrows].Type, ItemIDArrow) { // not a missile in the slot
		plus += int8(c.Slots[SlotArrows].Plus)
	}
	c.Bonus[0] += Modifier(plus)
	if c.Race == RaceElf && (e.isBow(w.Type) || e.isKind(w.Type, ItemIDShortSword, ItemIDLongSword)) {
		plus++ // elves with bows and swords
	}
	c.THAC0 += byte(plus)
	c.Dice[0] = Count(e.Base(typ).Dice)
	c.Sides[0] = Count(e.Base(typ).Sides)
}

// armourMove = 00BA:01FB (char, item): body armour sets the movement by
// its weight; magical armour moves 3 faster up to 9.
func (e *RuleState) armourMove(it *Item, c *Character) {
	if e.Base(it.Type).Slot != SlotArmor {
		return
	}
	switch w := it.Weight; {
	case w <= 0x96:
		c.Move = c.BaseMove
	case w <= 0x18F:
		c.Move = 9
	default:
		c.Move = 6
	}
	if it.Plus != 0 && c.Move <= 9 {
		c.Move += 3
	}
}

// itemAC = 00BA:0282 (char, item, ac, ringOnly): ac[1] shield, ac[2]
// protective items, ac[3] best ring, ac[4] best armour; magical body
// armour cancels rings.
func (e *RuleState) itemAC(c *Character, it *Item, ac *[5]int8) (ringOnly bool) {
	b := e.Base(it.Type)
	if !b.givesAC() {
		return
	}
	v, slot, plus := int8(b.AcPoints()), b.Slot, int8(it.Plus)
	switch {
	case slot == SlotShield:
		ac[1] = v + plus
	case v == 0:
		switch {
		case slot == 9:
			if plus > ac[3] {
				ac[3] = plus
			}
		default:
			ac[2] += plus
		}
		c.SaveBonus += it.Plus2 // a ring's or cloak's plus, from fixItemRecord and RandomItem
	case int(v)+int(plus) > int(ac[4]): // armour's scale: the best of armour and bracers
		ac[4] = v + plus
		if plus > 0 && slot == SlotArmor {
			ringOnly = true
		}
	}
	return
}

// strIndex = 00BA:11EA: strength as one scale (18/xx = 18..23, 19+ = 24+).
func strIndex(c *Character) int {
	return StrengthIndex(int(c.Stats[StatStr]), int(c.Str00))
}

// strHit = 00BA:1282.
func strHit(c *Character) int8 {
	if c.StrBonus == 0 {
		return 0
	}
	return int8(StrengthToHit(strIndex(c)))
}

// strDamage = 00BA:1325.
func strDamage(c *Character) int8 {
	if c.StrBonus == 0 {
		return 0
	}
	return int8(StrengthDamage(strIndex(c)))
}

// IsHelpless = 00BA:0B73: any of the four affects in DS:286C.
func (e *RuleState) IsHelpless(c *Character) bool {
	found := false
	for i := 1; i <= 4; i++ {
		if e.FindAffect(c, helplessAffects[i]) != nil {
			found = true
		}
	}
	return found
}

// intSlots and countSlots move one spell class's slots between the record's bytes and ClassStats' ints.
func intSlots(c [MaxSpellLevel]Count) (s LevelSlots) {
	for i, n := range c {
		s[i] = int(n)
	}
	return s
}

func countSlots(s LevelSlots) (c [MaxSpellLevel]Count) {
	for i, n := range s {
		c[i] = Count(n)
	}
	return c
}
