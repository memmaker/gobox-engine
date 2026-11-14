package rules1e

import (
	"fmt"
	"gobox/rules"
	"slices"
	"strconv"
	"strings"
)

// The view side of core.Rules (fold step E-R6): what the Engine's sheets,
// menus and shops read of a character, as words and numbers.

// word is c's sheet word for w (por sheetPage).
func word(c *Character, w rules.WordFact) (rules.Value, error) {
	switch w.Of {
	case rules.FactRace:
		return rules.Text(RaceName(c.Race)), nil
	case rules.FactSex:
		return rules.Text(GenderNames[c.Sex]), nil
	case rules.FactAlignment:
		return rules.Text(AlignmentNames[int8(c.Alignment)]), nil
	case rules.FactClasses:
		return rules.Text(ClassName(c.Class)), nil
	case rules.FactExceptional: // shown after an 18: "(57)", "(00)"
		if c.Stats[0] != 0x12 || c.Str00 == 0 {
			return rules.Text(""), nil
		}
		return rules.Text(fmt.Sprintf("(%02d)", c.Str00%100)), nil
	}
	return nil, errNoFact
}

// PorStatus is core's status of each record status (the DS:1047 order).
var PorStatus = [...]rules.Status{StatusOkay: rules.StatusOkay, StatusAnimated: rules.StatusAnimated, statusTempGone: rules.StatusTempGone,
	StatusRunning: rules.StatusRunning, StatusUnconscious: rules.StatusUnconscious, StatusDying: rules.StatusDying,
	StatusDead: rules.StatusDead, StatusStoned: rules.StatusStoned, StatusGone: rules.StatusGone}

// levelSpan (extra): the experience c's level took and its next level
// takes, for the class with the nearest next level; to is 0 when it can
// rise no further.
func levelSpan(c *Character) (from, to int64) {
	for i := 0; i <= 7; i++ {
		if lv := int8(c.Levels[i]); lv > 0 && lv < 9 && !Dormant(c, i) { // columns 2..9 hold levels 2..9; the rest overlap other tables
			if n := int64(XPForLevel(PorClass[i], int(lv+1))); n > 0 && (to == 0 || n < to) {
				from, to = 0, n
				if lv > 1 {
					from = int64(XPForLevel(PorClass[i], int(lv)))
				}
			}
		}
	}
	return from, to
}

// ItemLines is the Look text for an item, at most 0x1E wide. With vs (the
// item readied in the slot it would take) the damage and armour class lines
// end with the difference to it, " (+1)" better, " (-2)" worse.
// The damage dice count the known magic plus (Mace +2 is 1d6+3).
func (e *RuleState) ItemLines(it, vs *Item) []string {
	typ := it.Type
	b := e.Base(typ)
	delta := func(v, o float64) string {
		if vs == nil || v == o {
			return ""
		}
		sign := ""
		if v > o {
			sign = "+"
		}
		return " (" + sign + strings.TrimSuffix(strconv.FormatFloat(v-o, 'f', 1, 64), ".0") + ")" // one decimal at most
	}
	dice := func(n, d byte, bonus int8) string {
		s := strconv.Itoa(int(n)) + "d" + strconv.Itoa(int(d))
		if bonus > 0 {
			s += "+"
		}
		if bonus != 0 {
			s += strconv.Itoa(int(bonus))
		}
		return s
	}
	var l []string
	add := func(s string) { l = append(l, s) }
	slot, flags := b.Slot, b.Usage
	words := func(w WordID) bool { return it.Words[0] == w || it.Words[1] == w || it.Words[2] == w }
	if b.Dice > 0 && b.Sides > 0 {
		hands := "one hand"
		if b.Hands >= 2 {
			hands = "two hands"
		}
		add("Weapon, " + hands)
		plus := int8(e.KnownPlus(it)) // the Damage lines count the known magic plus, as their comparison does
		add("Damage " + dice(b.Dice, b.Sides, int8(b.Bonus)+plus) + delta(e.DamageAvg(it), e.DamageAvg(vs)))
		if b.LargeDice > 0 && (b.LargeDice != b.Dice || b.LargeSides != b.Sides || b.LargeBonus != b.Bonus) {
			add(" vs large " + dice(b.LargeDice, b.LargeSides, int8(b.LargeBonus)+plus) + delta(e.LargeAvg(it), e.LargeAvg(vs)))
		}
		if r := int(b.Range); r > 1 {
			add("Range " + strconv.Itoa(r-1) + ", worse past " + strconv.Itoa((r-1)/3))
			switch n := max(int(b.Speed), 2); {
			case n == 2:
				add("1 shot per round")
			case n%2 == 0:
				add(strconv.Itoa(n/2) + " shots per round")
			default:
				add(strconv.Itoa(n) + " shots per 2 rounds")
			}
		}
		switch {
		case e.isKind(typ, ItemIDArrow, ItemIDQuarrel): // before the flags: quarrels carry the crossbows' 0x80
			add("Ammunition")
		case flags&0x14 == 0x14:
			add("Thrown, or used in melee")
		case flags&0x10 != 0:
			add("Thrown, used up; not with")
			add("a foe adjacent")
		case flags&0x80 != 0:
			add("Needs quarrels")
		case flags&1 != 0 && flags&8 != 0:
			add("Needs arrows")
		}
		if flags&4 != 0 {
			add("Strength adds to hit, damage")
		} else if flags&2 != 0 {
			add("Dexterity adds to hit")
		}
	}
	if v := b.AcBetter(); b.AcPoints() > 0 {
		switch {
		case slot == SlotArmor:
			add("Armor, AC " + strconv.Itoa(10-v) + delta(e.Protection(it), e.Protection(vs)))
			switch w := it.Weight; {
			case w > 0x18F:
				add("Movement 6")
			case w > 0x96:
				add("Movement 9")
			}
		case slot == SlotShield:
			add("Shield, AC " + strconv.Itoa(v) + " better" + delta(e.Protection(it), e.Protection(vs)))
		default: // bracers: on the armour's scale and in its place, the better one counts (itemAC's ac[4])
			add("AC " + strconv.Itoa(10-v-int(e.KnownPlus(it))) + ", as armor" + delta(e.Protection(it), e.Protection(vs)))
			add("Worn in place of armor")
		}
	}
	switch {
	case e.isKind(typ, ItemIDHolyWater):
		add("Hurts undead: 1d6+1 a hit")
	case e.isKind(typ, ItemIDOil):
		add("Burns some monsters: 3d8")
	case words(0x76):
		add("Readied: turns gaze attacks")
		add("back, wards some monsters")
	case words(0x98) && words(0xFC):
		add("Readied: wards some monsters")
	case slot > 0xA && slot < 0xE:
		add("Scroll: read to cast its")
		add("spells, once each")
	case slot == 0xA && b.Hands == 0:
		add("Treasure: worth selling")
	}
	if words(0xB1) && b.Dice > 0 {
		add("Silver: hurts monsters that")
		add("ignore plain weapons")
	}
	if it.Hidden == 0 && int8(it.Plus) > 0 && (b.Dice > 0 || b.AcPoints() > 0) {
		add("Magic +" + strconv.Itoa(int(int8(it.Plus))))
	}
	add("Price " + strconv.Itoa(int(it.Price))) // extra: no weight (the party inventory counts slots)
	if slot != 0xA || b.Hands != 0 {
		var who []string
		for k, m := range classMask { // the DS:05D6 masks, in the DS:0D38 order
			if ClassBits(b.Classes)&m != 0 {
				who = append(who, strings.TrimSpace(ClassNames[k]))
			}
		}
		switch {
		case len(who) == 0:
			add("No class can use it")
		case len(who) >= 7:
			add("Most classes can use it")
		default:
			s := "For"
			for _, w := range who {
				if len(s)+1+len(w) > 0x1E {
					add(s)
					s = "   "
				}
				s += " " + w
			}
			add(s)
		}
	}
	return l
}

// Unknown (extra): Id would reveal a word of the item's name.
func (it *Item) Unknown() bool {
	for i := 1; i <= 3; i++ {
		if it.Words[i-1] != 0 && (it.Hidden>>(3-i))&1 != 0 {
			return true
		}
	}
	return false
}

// loseFightMorale is the lose-fight fixture's foes' morale (Control's low
// seven bits): its roll, twice it, is 100, so aiFleeCheck holds whatever the
// hit points (POR's cap reads above 102 as 0).
const loseFightMorale = 50

// fixture is a dev scene's set-up (por DevScene, unchanged).
func (r *Rules) fixture(f rules.Fixture) error {
	s := r.S
	switch f.ID {
	case rules.FixtureShop:
		s.Treasure = nil
		s.Treasure = s.devItems(ItemGauntlets, ItemJavelin, ItemArrow, ItemMageScroll, ItemLeatherArmor, ItemBillGuisarme)
		if len(s.Party) > 2 { // unknown items for Id
			s.Party[0].Items = append(s.Party[0].Items, s.devItems(ItemJavelin)...)
			s.Party[2].Items = append(s.Party[2].Items, s.devItems(ItemLeatherArmor, ItemBillGuisarme)...)
		}
	case rules.FixtureHall, rules.FixtureHallPoor:
		s.Game.PartyExp = 0xFF
		s.Party[0].Exp, s.Purse.Gold = 1<<20, 1<<20
		if f.ID == rules.FixtureHallPoor {
			s.Purse.Gold = 0
		}
	case rules.FixtureTemple:
		c := r.chars[f.Who]
		if c == nil {
			return errNoChar
		}
		s.Purse.Gold = 10000
		c.HP = 1
	case rules.FixtureAppraise:
		s.Purse.Gems, s.Purse.Jewelry = 3, 2
	case rules.FixtureMissile: // items for everyone: a sling, a bow and arrows, javelins
		launchers := []ItemType{ItemSling, ItemShortBow} // the first two who may: a sling, a bow with arrows (the Launcher slot: hands kept)
		fits := func(t ItemType, c *Character) bool {    // the game's type of t's kind is c's
			g, ok := s.devType(t)
			return ok && s.Base(g).Classes&c.ClassMask != 0
		}
		for i, c := range s.Party {
			s.Who = i
			s.ApplyItems(c)
			if c.Status != StatusOkay { // unarmed members too (a save's party may hold no weapon)
				continue
			}
			if len(launchers) > 0 && fits(launchers[0], c) {
				for _, it := range s.devItems(launchers[0]) {
					s.ReadyItem(s.AddItemCopy(c, it))
				}
				s.ApplyItems(c)
				if launchers[0] == ItemShortBow {
					for _, it := range s.devItems(ItemArrow) {
						s.ReadyItem(s.AddItemCopy(c, it))
					}
				}
				launchers = launchers[1:]
			}
			if c.Slots[SlotLauncher] == nil { // else javelins in the Missiles slot
				for _, it := range s.devItems(ItemJavelin) {
					s.ReadyItem(s.AddItemCopy(c, it))
				}
			}
			s.ApplyItems(c)
		}
	case rules.FixtureCastFight: // every member has Cure Light Wounds and Stinking Cloud (an area aim) memorized
		for _, c := range s.Party {
			c.Memorized[0], c.Memorized[1] = SpCureLightWounds, spStinkingCloud
		}
	case rules.FixtureAutoCombat: // as Quick for everyone
		for _, c := range s.Party {
			c.Auto = 1
		}
	case rules.FixtureWinFight, rules.FixtureLoseFight: // the placed foes at 1 hp (win), or at 250 against a party of 1 hp (lose)
		for _, c := range s.CombatantSlots[:s.Ncomb] {
			switch {
			case c == nil:
			case f.ID == rules.FixtureWinFight && c.Side != s.Party[0].Side:
				c.HP = 1
			case f.ID == rules.FixtureLoseFight && c.Side != s.Party[0].Side:
				c.HP = 250
				if c.Control > 0x7F { // a monster of morale 0 flees (aiFleeCheck: PoD's bulettes got away), so the foes hold: 50, a roll of 100
					c.Control = 0x80 | loseFightMorale
				}
			case f.ID == rules.FixtureLoseFight:
				c.HP = 1
			}
		}
	case rules.FixtureCast, rules.FixtureCastKO: // Cure Light Wounds memorized by every able member and Who; castko: Who unconscious
		c := r.chars[f.Who]
		if c == nil {
			return errNoChar
		}
		if f.ID == rules.FixtureCastKO {
			c.Status, c.Okay = StatusUnconscious, 0
		}
		for _, m := range s.Party {
			if m.Status == StatusOkay {
				m.Memorized[0] = SpCureLightWounds
			}
		}
		c.Memorized[0] = SpCureLightWounds
	case rules.FixtureSheetCast, rules.FixtureLootDetect: // Who has Cure Light Wounds (sheetcast) or Detect Magic (lootdetect) memorized
		c := r.chars[f.Who]
		if c == nil {
			return errNoChar
		}
		c.Memorized[0] = SpCureLightWounds
		if f.ID == rules.FixtureLootDetect {
			c.Memorized[0] = SpDetectMagicCl
		}
	case rules.FixtureEffects: // spell effects for the View page: many on member 1, one on member 2, items for the inventory
		for _, t := range []AffectType{AffBless, AffCurse, AffMonsterInvisible, AffPoisoned, AffRegenerating, AffFireResistance} {
			s.AddAffectNow(s.Party[0], t, 0, 100)
		}
		s.AddAffectNow(s.Party[1], AffMonsterInvisible, 0, 100)
		for i, it := range s.devItems(ItemGauntlets, ItemJavelin, ItemArrow) {
			s.AddItemCopy(s.Party[i%2], it)
		}
		s.ApplyItems(s.Party[0])
		s.ApplyItems(s.Party[1])
	case rules.FixtureInventory: // member 1 fully equipped, a party inventory to scroll
		c := s.Party[0]
		for _, it := range c.Items { // the save's own gear joins the party inventory
			it.Readied = 0
		}
		for _, it := range s.devItems(ItemLongSword, ItemShield, ItemPlateMail, ItemGauntlets, 0x40, 0x41, 0x42, ItemCloak, 0x44, ItemRing, ItemProtectionRing, ItemArrow) {
			it.Readied = 1
			c.Items = append(c.Items, it)
		}
		for _, it := range s.devItems(ItemDagger, ItemShortBow, ItemPotion, ItemMageScroll, ItemWand, ItemValuable, ItemHolyWater, ItemOil, ItemJavelin, ItemQuarrel, ItemChainMail, ItemDart, ItemHammer, 0x48, ItemClericScroll, ItemStrengthPotion, 0x4B, ItemBracers) {
			s.AddItemCopy(c, it)
		}
		s.ApplyItems(c)
	case rules.FixtureNoPicker: // no one can pick a lock
		for _, c := range s.Party {
			c.Thief[1] = 0
		}
	case rules.FixtureDual: // member 1 a human fighter who may change class
		c := s.Party[0]
		c.Race, c.Class, c.Alignment = RaceHuman, ClsFighter, 0
		c.Stats = [6]Score{17, 17, 17, 17, 17, 17}
	case rules.FixtureHurt: // member 1 at 1 hp
		s.Party[0].HP = 1
	case rules.FixtureLearn: // member 5 ready to rise
		s.Party[4].Exp = 1 << 20
	case rules.FixturePaladin, rules.FixtureDrop, rules.FixtureUsable: // Who a paladin of level 3; with a dagger to drop; with a potion to drink
		c := r.chars[f.Who]
		if c == nil {
			return errNoChar
		}
		switch f.ID {
		case rules.FixturePaladin:
			c.Class, c.Levels[ClsPaladin] = ClsPaladin, 3
		case rules.FixtureDrop:
			for _, it := range s.devItems(ItemDagger) {
				s.AddItemCopy(c, it)
			}
		case rules.FixtureUsable:
			for _, it := range s.devItems(ItemStrengthPotion) {
				s.AddItemCopy(c, it)
			}
		}
	case rules.FixtureLoot: // a fight's leavings to take
		s.Treasure = s.devItems(ItemGauntlets, ItemJavelin, ItemArrow)
		s.Loot = Purse{Gold: 156, Gems: 7}
	default:
		return errBadValue
	}
	return nil
}

// learn adds a learnable spell to the book (por Learn, the hall's pick).
func (r *Rules) learn(a rules.Learn) error {
	c := r.chars[a.Who]
	if c == nil {
		return errNoChar
	}
	n := slices.Index(SpellIDs[:], a.Spell)
	if n <= 0 || !slices.Contains(r.S.Learnable(c), SpellID(n)) {
		return ErrNotNow
	}
	c.SpellBook[n-1] = 1
	return nil
}

// Handle is it's ItemHandle, given at its first use and kept while it
// lives; nil is core.NoItem. Until the Engine holds handles itself, it names
// its items through Handle.
func (r *Rules) Handle(it *Item) rules.ItemHandle {
	if it == nil {
		return rules.NoItem
	}
	if h, ok := r.handles[it]; ok {
		return h
	}
	if r.handles == nil {
		r.handles, r.items = map[*Item]rules.ItemHandle{}, map[rules.ItemHandle]*Item{}
	}
	r.nextItem++
	r.handles[it], r.items[r.nextItem] = r.nextItem, it
	return r.nextItem
}

// Held is the item of handle h, nil for core.NoItem (goes when the Engine
// holds handles).
func (r *Rules) Held(h rules.ItemHandle) *Item { return r.items[h] }

// SellValue = ovr20:1CEE (item): what selling it pays, in copper (200 a
// gold piece, so half the value; a twentieth per piece of a stack).
// extra: the halves are kept as copper where the original dropped them.
func (it *Item) SellValue() uint32 {
	v := uint32(it.Price) * 100
	if it.Count > 0 { // a stackable item's Price is a bundle of 20's; the original's Count > 1 sold a lone one at the whole bundle's price
		// (the arrows/quarrels branch of the original can never be taken)
		v = uint32(it.Count) * v / 0x14
	}
	return v
}

// sheetView fills the sheet's view figures of c (por showCombatStats,
// ovr20:0608): THAC0 as shown, movement, the effects, whether it can cast.
func (r *Rules) sheetView(c *Character, s *rules.Sheet) {
	s.Age, s.THAC0 = int32(c.Age), int32(byte(0x3C-c.THAC0))
	m := c.Move
	// ruling U5, a display fix: the original (ovr20:086A) doubled under Slow and halved under Haste; the Pascal fix lands with viewchar in src/pure
	if r.S.FindAffect(c, AffHaste) != nil {
		m <<= 1
	}
	if r.S.FindAffect(c, AffSlow) != nil {
		m >>= 1
	}
	s.Move = int32(m)
	s.Effects = r.S.AffectList(c)
	s.CanCast = slices.ContainsFunc(c.Memorized[:0x15], SpellID.Held) // memorized, not still to be learned
}

// TrainsTo is what training in the classes of mask makes c: "a level 4
// Fighter" for each class it trains in (the hall's Becomes).
func (e *RuleState) TrainsTo(c *Character, mask byte) []string {
	var to []string
	for k := 0; k <= 7; k++ {
		if int8(c.Levels[k]) > 0 && !Dormant(c, k) && byte(HpClassMask[k])&mask > 0 {
			to = append(to, "a level "+strconv.Itoa(int(int8(c.Levels[k]))+1)+" "+ClassNames[k])
		}
	}
	return to
}

// dollSlotName is the shared Item.Slot of it in doll slot d: in the
// Missiles box arrows, bolts or (extra) missiles, a thrown weapon.
func (e *RuleState) dollSlotName(it *Item, d int) string {
	switch {
	case d == rules.DollArrows && e.isKind(it.Type, ItemIDArrow):
		return "arrows"
	case d == rules.DollArrows && e.isKind(it.Type, ItemIDQuarrel):
		return "bolts"
	case d == rules.DollArrows:
		return "missiles"
	}
	return rules.DollNames[d]
}

// ItemKindOf is the item's tab: what the weapon hand holds (and its
// ammunition), what is worn, else the rest (rings among them).
func (e *RuleState) ItemKindOf(c *Character, it *Item) rules.InvTab {
	switch d := e.DollSlot(c, it); {
	case d == rules.DollWeapon || d == rules.DollArrows || d == rules.DollLauncher:
		return rules.TabWeapon
	case d >= 0 && d != rules.DollRing1 && d != rules.DollRing2:
		return rules.TabArmor
	}
	return rules.TabMisc
}

// ItemTab is it's inventory tab as c sees it ("weapon", "armor", "ring",
// else "misc") and, readied, the shared Item.Slot it is readied in.
func (e *RuleState) ItemTab(c *Character, it *Item) (tab, slot string) {
	d := e.DollSlot(c, it)
	tab = [...]string{rules.TabMisc: "misc", rules.TabWeapon: "weapon", rules.TabArmor: "armor"}[e.ItemKindOf(c, it)]
	if d == rules.DollRing1 || d == rules.DollRing2 {
		tab = "ring"
	}
	if it.Readied != 0 && d >= 0 {
		slot = e.dollSlotName(it, d)
	}
	return tab, slot
}

// wordNames are the words of it's name, known or not.
func (e *RuleState) wordNames(it *Item) []string {
	var l []string
	for _, w := range it.Words {
		if w != 0 {
			l = append(l, e.itemWord(w))
		}
	}
	return l
}

// BookSpell is spell number id of c's book as the sheet lists it: its
// class and level (Tables.Spells) and how many of it c plans to memorize.
func (e *RuleState) BookSpell(c *Character, id int) (class SpellClass, level SpellProp, planned int) {
	sp := SpellID(id)
	return e.SpellClass(sp), e.SpellLevel(sp), SpellCount(c.Plan[:], sp)
}
