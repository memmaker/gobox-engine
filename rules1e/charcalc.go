package rules1e

import (
	"gobox/rules"
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
	levels, old := map[rules.Class]int{}, map[rules.Class]int{}
	for i, class := range PorClass {
		levels[class], old[class] = int(int8(c.Levels[i])), int(int8(c.OldLevels[i]))
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
		Race: c.Race, OldLevels: old, Powers: powers,
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
	skills := e.T().ThiefSkills(int(int8(c.Levels[ClsThief])), c.Race, int(c.Stats[StatDex]), ItemPowers{})
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
	// ReadiedMissiles (extra) is the Readied of a weapon thrown or used in
	// melee (a spear) readied into the Missiles slot, not the weapon hand
	ReadiedMissiles = 5
	SlotArmor       = 2
	slotRing1       = 9
	slotRing2       = 10
	SlotArrows      = 11
	SlotBolts       = 12
	// extra: the Missiles slot (a thrown-only weapon: a dart, a javelin, a
	// flask of oil) and the Launcher slot (a bow, crossbow or sling). Both
	// take no hands; the Ranged action shoots or throws straight from them (Ranged, Wield).
	SlotMissiles = 13
	SlotLauncher = 14
	slotCount    = 15
)

// Base item Usage flags (+0xB) the slots go by.
const (
	usageMeleeToo = 0x04 // with usageThrown: also a melee weapon (a spear, a hand axe)
	usageShoots   = 0x08 // fires (a launcher; with usageThrown the throw itself)
	usageThrown   = 0x10 // thrown
)

// launcher (extra): a bow, crossbow or sling, readied into the Launcher slot.
func (b *ItemBase) launcher() bool {
	return b.Usage&usageShoots != 0 && b.Usage&usageThrown == 0 && b.Range > 1
}

// thrownOnly (extra): a weapon only thrown, readied into the Missiles slot
// (one also used in melee stays in the weapon hand).
func (b *ItemBase) thrownOnly() bool {
	return b.Usage&usageThrown != 0 && b.Usage&usageMeleeToo == 0
}

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
// they stand in for (itemAC: only the better of the two counts), and
// (extra) a launcher and a thrown-only weapon take the Launcher and
// Missiles slots instead of the weapon hand.
func (e *RuleState) readySlot(typ ItemType) int {
	b := e.Base(typ)
	switch {
	case e.isKind(typ, ItemIDBracers):
		return SlotArmor
	case b.launcher():
		return SlotLauncher
	case b.thrownOnly():
		return SlotMissiles
	}
	return int(b.Slot)
}

// handsFor is the hands a readied item of type typ takes: its base
// record's, none in the Launcher and Missiles slots (extra).
func (e *RuleState) handsFor(typ ItemType) byte {
	if s := e.readySlot(typ); s == SlotLauncher || s == SlotMissiles {
		return 0
	}
	return e.Base(typ).Hands
}

// meleeToo (extra): a thrown weapon also used in melee (a spear, a hand
// axe): it goes into the weapon hand or the Missiles slot.
func (b *ItemBase) meleeToo() bool {
	return b.Usage&usageThrown != 0 && b.Usage&usageMeleeToo != 0
}

// asMissile (extra): it, unreadied, would go into c's Missiles slot: a
// thrown weapon also used in melee, where the player's Ready says
// (readyTo), else with the weapon hand holding another and the Missiles
// box free.
func (e *RuleState) asMissile(c *Character, it *Item) bool {
	if it.Readied != 0 || !e.Base(it.Type).meleeToo() {
		return false
	}
	switch e.readyTo {
	case rules.ReadyToMissiles:
		return true
	case rules.ReadyToHand:
		return false
	}
	w := c.Slots[SlotWeapon]
	return w != nil && w != it && e.missileReadied(c) == nil
}

// moveReadied (extra): c's readied weapon it, thrown or used in melee,
// changes between the weapon hand and the Missiles slot (to), what holds
// the place coming off first; any other item stays where it is.
func (e *RuleState) moveReadied(c *Character, it *Item, to rules.ReadyTo) {
	missiles := to == rules.ReadyToMissiles
	if !e.Base(it.Type).meleeToo() || missiles == (it.Readied == ReadiedMissiles) {
		return
	}
	if it.Cursed != 0 {
		e.notice("It's Cursed")
		return
	}
	off := func(o *Item) bool { // o comes off: false when it stays (cursed, no room)
		e.ReadyItem(o)
		e.ApplyItems(c)
		return o.Readied == 0
	}
	if missiles {
		if o := e.missileReadied(c); o != nil && !off(o) {
			return
		}
		it.Readied = ReadiedMissiles
		return
	}
	if o := c.Slots[SlotWeapon]; o != nil && !off(o) {
		return
	}
	if o := c.Slots[SlotShield]; o != nil && int(e.handsFor(it.Type))+int(c.Hands) > 2 && !off(o) { // a two-handed one takes the shield off too
		return
	}
	it.Readied = 1
}

// itemSlot is the slot c's item it fills, readied (or would, unreadied):
// its type's (readySlot), the Missiles slot for a melee weapon thrown from there.
func (e *RuleState) itemSlot(c *Character, it *Item) int {
	if it.Readied == ReadiedMissiles || e.asMissile(c, it) {
		return SlotMissiles
	}
	return e.readySlot(it.Type)
}

// itemHands is the hands c's item it takes there (handsFor; none in the Missiles slot).
func (e *RuleState) itemHands(c *Character, it *Item) byte {
	if e.itemSlot(c, it) == SlotMissiles {
		return 0
	}
	return e.handsFor(it.Type)
}

// missileSlots (extra) are the slots behind the doll's one Missiles box:
// one of them holds a readied stack at a time.
var missileSlots = [3]int{SlotMissiles, SlotArrows, SlotBolts}

// missileReadied (extra) is c's item in the Missiles box, nil for none.
func (e *RuleState) missileReadied(c *Character) *Item {
	for _, s := range missileSlots {
		if c.Slots[s] != nil {
			return c.Slots[s]
		}
	}
	return nil
}

// rangedItem (extra) is what c's Ranged action uses: the launcher when it
// has its ammunition (or needs none), else the Missiles slot's thrown
// weapon; nil for none (a launcher without ammunition is none).
func (e *RuleState) rangedItem(c *Character) *Item {
	l, m := c.Slots[SlotLauncher], c.Slots[SlotMissiles]
	var ammo *Item
	if l != nil && e.ammoWith(c, l, &ammo) { // a launcher without its ammunition shoots nothing: no Ranged
		return l
	}
	return m
}

// Wield (extra) is what a fighter attacks with: the weapon in hand (nil:
// bare hands), the off hand's, or the Launcher or Missiles slot's (the
// Ranged action). Every combat rule takes the Wield of its attack; nothing
// ever stands in the weapon hand for another.
type Wield struct {
	Weapon *Item
	own    bool // the weapon hand's: its figures are the fighter's own (THAC0, Dice, Sides, Bonus, kept by ApplyItems)
}

// Hand is c's attack with the weapon in hand (bare hands without one).
func (e *RuleState) Hand(c *Character) Wield { return Wield{Weapon: c.Slots[SlotWeapon], own: true} }

// Ranged is c's attack with its Launcher or Missiles slot (rangedItem);
// false when it has none to shoot or throw.
func (e *RuleState) Ranged(c *Character) (Wield, bool) {
	it := e.rangedItem(c)
	return Wield{Weapon: it}, it != nil
}

// Wielding is c's Ranged attack when ranged and it has one, else its Hand.
func (e *RuleState) Wielding(c *Character, ranged bool) Wield {
	if w, ok := e.Ranged(c); ranged && ok {
		return w
	}
	return e.Hand(c)
}

// blow is the figures c strikes with (attack slot 1) wielding w, at a
// large target (attackSlots' test) the weapon's dice for one.
type blow struct {
	thac0       byte
	dice, sides Count
	bonus       Modifier
}

// thac0 is c's to-hit figure wielding w.
func (e *RuleState) thac0(c *Character, w Wield) byte {
	if w.own {
		return c.THAC0
	}
	return e.weaponBlow(c, w.Weapon).thac0
}

// ApplyItems = 00BA:0BB8: everything that follows from the items carried
// and readied: slots, weight, hands, attack and damage, armour class,
// movement and saving throw bonus.
func (e *RuleState) ApplyItems(c *Character) {
	if e.partyMember(c) { // extra: unreadied items belong to the party inventory
		e.stow(c)
	}
	c.Slots = [slotCount]*Item{}
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
		switch s := e.itemSlot(c, it); {
		case s == SlotLauncher || s == SlotMissiles: // extra
			c.Slots[s] = it
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
		c.Hands += e.itemHands(c, it)
	}
	c.Dice, c.Sides, c.Bonus = c.BaseDice, c.BaseSides, c.BaseBonus
	var ac [5]int8
	ringOnly := false
	c.SaveBonus = 0
	c.AC[0] = ArmourClass(c.BaseAC)
	c.Move = c.BaseMove
	c.THAC0 = c.Thac0Base
	ac[0] = rules.DexAC(byte(c.Stats[StatDex]))
	if w := c.Slots[SlotWeapon]; w == nil {
		c.THAC0 += byte(strHit(c))
		c.Bonus[0] += Modifier(strDamage(c))
	} else {
		b := e.weaponBlow(c, w)
		c.THAC0, c.Dice[0], c.Sides[0], c.Bonus[0] = b.thac0, b.dice, b.sides, b.bonus
	}
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
	if FighterKind(c) && c.raceNumber() > 0 {
		c.FightLevel = byte(max(c.Levels[ClsFighter], c.Levels[ClsPaladin], c.Levels[ClsRanger]))
	} else {
		c.FightLevel = 1
	}
}

// weaponBlow = 00BA:0000: to-hit and damage of c with the weapon w, by c's
// scores and the ammunition readied (the original worked them out for the
// readied weapon, into the character).
func (e *RuleState) weaponBlow(c *Character, w *Item) blow {
	typ := w.Type
	flags := e.Base(typ).Usage
	out := blow{thac0: c.Thac0Base}
	if flags&2 != 0 { // missile
		out.thac0 += byte(rules.DexMissile(byte(c.Stats[StatDex])))
	}
	out.bonus = Modifier(e.Base(typ).Bonus)
	if flags&4 != 0 { // strength applies
		out.thac0 += byte(strHit(c))
		out.bonus += Modifier(strDamage(c))
	}
	plus := int8(w.Plus)
	if flags > 0x7F && c.Slots[SlotBolts] != nil {
		plus += int8(c.Slots[SlotBolts].Plus)
	}
	if flags&1 != 0 && c.Slots[SlotArrows] != nil && e.isKind(c.Slots[SlotArrows].Type, ItemIDArrow) { // not a missile in the slot
		plus += int8(c.Slots[SlotArrows].Plus)
	}
	out.bonus += Modifier(plus)
	if c.Race == RaceElf && (e.isBow(w.Type) || e.isKind(w.Type, ItemIDShortSword, ItemIDLongSword)) {
		plus++ // elves with bows and swords
	}
	out.thac0 += byte(plus)
	out.dice = Count(e.Base(typ).Dice)
	out.sides = Count(e.Base(typ).Sides)
	return out
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
