package rules1e

// Items: names, lists and the START helpers around them.

import (
	"gobox/rules"
	"slices"
	"strconv"
)

// itemWord = DS:10A7 + 0x15*n: word n of the game's table.
func (e *RuleState) itemWord(n WordID) string {
	w := ItemWords[:]
	if e.T().Words != nil {
		w = e.T().Words
	}
	return w[n]
}

func (it *Item) Name() string {
	n := int(it.PName[0])
	if n > itemNameMax {
		n = itemNameMax
	}
	return string(it.PName[1 : n+1])
}

// itemNameMax is the name field's capacity (string[40]): the record's, not
// the page's width (textCols grows with the window; a wider cut wrote a
// length byte past the 40 bytes copied).
const itemNameMax = 0x28

func (it *Item) setName(s string) {
	s = rules.Cut(s, itemNameMax)
	it.PName[0] = Letter(len(s))
	copy(it.PName[1:itemNameMax+1], []Letter(s))
}

// ItemName = 00BA:0442 (char, item, x, y, readied, display): builds the
// item's shown name into its name field ("Yes"/"No" readied column, "*"
// for magic when someone detects magic, count, the known words with the
// plural "s" on the right one) and optionally prints it.
func (e *RuleState) ItemName(c *Character, it *Item, readied bool) {
	name := ""
	if readied {
		if it.Readied != 0 {
			name = " Yes  "
		} else {
			name = " No   "
		}
	}
	detect := false
	for _, p := range e.Party {
		if detect {
			break
		}
		if e.FindAffect(p, affDetectMagic) != nil {
			detect = true
		}
	}
	if detect && (int8(it.Plus) > 0 || int8(it.Plus2) > 0 || it.Cursed != 0) {
		name = rules.Cut(name+"* ", itemNameMax)
	}
	if it.Count > 0 {
		name = rules.Cut(name+strconv.Itoa(int(it.Count))+" ", itemNameMax)
	}
	mask := 0
	for i := 1; i <= 3; i++ {
		if it.Words[i-1] != 0 && (it.Hidden>>(3-i))&1 == 0 {
			mask += 1 << (i - 1)
		}
	}
	plural := false
	for i := 3; i >= 1; i-- {
		if (mask>>(i-1))&1 == 0 {
			continue
		}
		name = rules.Cut(name+e.itemWord(it.Words[i-1]), itemNameMax)
		s := " "
		if it.Count >= 2 && !plural {
			typ := it.Type
			switch {
			case 1<<(i-1) == mask,
				i == 1 && mask > 4 && !e.isKind(typ, ItemIDOil),
				i == 2 && mask&1 == 0,
				i == 3 && e.isKind(typ, ItemIDOil),
				e.isKind(typ, ItemIDArrow, ItemIDQuarrel) && it.Words[2] != 0xB1:
				s = "s "
				plural = true
			}
		}
		name = rules.Cut(name+s, itemNameMax)
	}
	it.setName(name + e.ItemStats(it))
}

// ItemStats (extra): Angband-style stats after the name, "(1d8) " for a
// weapon's base damage (its magic plus stays in the name), "[6] " for the
// armour class points armour, a shield or bracers give, the known plus
// counted in (Bracers +4 are <4>).
func (e *RuleState) ItemStats(it *Item) string {
	typ := it.Type
	s := ""
	if n, d := e.Base(typ).Dice, e.Base(typ).Sides; n > 0 && d > 0 {
		s = "(" + strconv.Itoa(int(n)) + "d" + strconv.Itoa(int(d))
		if b := int8(e.Base(typ).Bonus); b > 0 {
			s += "+" + strconv.Itoa(int(b))
		} else if b < 0 {
			s += strconv.Itoa(int(b))
		}
		s += ") "
	}
	if b := e.Base(typ); b.AcPoints() > 0 {
		s += "<" + strconv.Itoa(b.AcBetter()+int(e.KnownPlus(it))) + "> "
	}
	return s
}

// RemoveItem = 00BA:1509 (char, item): unlinks the item and frees it.
func (e *RuleState) RemoveItem(c *Character, it *Item) {
	if it == nil {
		e.notice("Nil Item pointer...")
		return
	}
	for i, x := range c.Items {
		if x == it {
			c.Items = append(c.Items[:i:i], c.Items[i+1:]...)
			return
		}
	}
	if i := slices.Index(e.Pool, it); i >= 0 && e.partyMember(c) { // extra: the party inventory
		e.Pool = slices.Delete(e.Pool, i, i+1)
		return
	}
	e.notice("Tried to Lose item & couldn't find it!")
}

// DestroyItems = pod O040:2066 for one member: every carried item of id
// (with a plus when magicalOnly) goes, a readied one put away first (as c
// acting), then c's items are applied again.
func (e *RuleState) DestroyItems(c *Character, id rules.ItemID, magicalOnly bool) {
	who := e.Who
	if i := e.PartyIndex(c); i < len(e.Party) {
		e.Who = i
	}
	for _, it := range e.Carried(c) {
		if e.kind(it.Type) != id || magicalOnly && it.Plus == 0 {
			continue
		}
		if it.Readied != 0 {
			e.ReadyItem(it)
		}
		e.RemoveItem(c, it)
	}
	e.ApplyItems(c)
	e.Who = who
}

// AddItemCopy = 00BA:15F7 (char, item): a copy of the item at the end of
// the character's list.
func (e *RuleState) AddItemCopy(c *Character, it *Item) *Item {
	n := *it
	if n.Readied == 0 && e.partyMember(c) { // extra: into the party inventory
		return e.addToPool(&n)
	}
	c.Items = append(c.Items, &n)
	// extra: a stack joins the first like stack the character has, so no
	// two exist below the 255 a stack holds (the byte count's limit)
	if n.Count == 0 || n.Spells[0] >= 2 {
		return &n
	}
	for _, p := range c.Items {
		if sameStack(p, &n) {
			e.joinItems(c, p)
			return p
		}
	}
	return &n
}

// sameStack: joinItems's test for two stacks of the same thing.
func sameStack(p, it *Item) bool {
	return p.Count > 0 && p.Spells[0] < 2 && p.Words == it.Words && p.Type == it.Type &&
		p.Plus == it.Plus && p.Plus2 == it.Plus2 && p.Cursed == it.Cursed && p.Weight == it.Weight
}

// ItemMend is one of a game's data fixes (Tables.ItemMends): the items of
// Type (with Word in word slot WordSlot when Word is not 0) get Fix.
type ItemMend struct {
	Type     ItemType `json:"type"`
	WordSlot uint8    `json:"wordSlot"`
	Word     WordID   `json:"word"` // 0: any words
	Fix      ItemFix  `json:"fix"`
}

// ItemFix names what an ItemMend changes; each checks the record is still
// broken, so a mended item is left as it is.
type ItemFix string

const (
	FixFlameTongue   ItemFix = "flameTongue"   // affect 0 -> bonus by target race and fire damage
	FixFeatherFall   ItemFix = "featherFall"   // no affect -> Feather Fall
	FixBracersAC2    ItemFix = "bracersAC2"    // the plus of Bracers AC 4 -> AC 2's
	FixCompositeBow  ItemFix = "compositeBow"  // a heavy crossbow in the data -> the long bow it is named
	FixPlus2FromPlus ItemFix = "plus2FromPlus" // no save bonus -> the plus counts for saves too
)

// PorItemMends are POR's broken data records, as the Pascal reference
// (por-re BUGS.md, FixItemRecord) mends them:
//   - the Flame Tongue (Valjevo Castle's well) had its affect byte 0 and
//     0x16 for data, so it was a usable item and never a Flame Tongue: it
//     now grants 0x06, the flame tongue's bonus by target and fire damage;
//   - the Ring of Feather Falling (under the Wealthy Area's trap door) had
//     no affect: it now grants Feather Fall (0x1B), as the DS:0970 record;
//   - Bracers AC 2 (the Nomad Camp shaman's) had the plus of Bracers AC 4;
//   - the Fine Composite Long Bow was a heavy crossbow;
//   - rings of protection and cloaks of displacement gave no save bonus.
var PorItemMends = []ItemMend{
	{Type: ItemLongSword, WordSlot: 0, Word: 0xF1, Fix: FixFlameTongue},
	{Type: ItemFeatherFallRing, WordSlot: 0, Word: 0xEE, Fix: FixFeatherFall},
	{Type: ItemBracers, WordSlot: 1, Word: 0xDC, Fix: FixBracersAC2},
	{Type: ItemHeavyCrossbow, WordSlot: 1, Word: WordID(ItemCompositeLongBow), Fix: FixCompositeBow},
	{Type: ItemProtectionRing, Fix: FixPlus2FromPlus},
	{Type: ItemCloakDisplacement, Fix: FixPlus2FromPlus},
}

// MendItem applies the game's ItemMends to it as it is read from an ITEM
// or MONnITM block or a save; a game without a list mends nothing.
func (t *Tables) MendItem(it *Item) {
	for _, m := range t.ItemMends {
		if it.Type != m.Type || m.Word != 0 && it.Words[m.WordSlot] != m.Word {
			continue
		}
		switch m.Fix {
		case FixFlameTongue:
			if it.Spells[2] == 0 {
				it.Spells[1], it.Spells[2] = ItemPower(AffBonusByTargetRace), ItemPower(AffItem0)
			}
		case FixFeatherFall:
			if it.Spells[2] == 0 {
				it.Spells[1], it.Spells[2] = ItemPower(AffFeatherFall), ItemPower(AffItem0)
			}
		case FixBracersAC2:
			if it.Plus == 6 {
				it.Plus = 8
			}
		case FixCompositeBow:
			it.Type = ItemCompositeLongBow
		case FixPlus2FromPlus:
			if it.Plus2 == 0 {
				it.Plus2 = it.Plus
			}
		}
		return // one mend per item, as the reference's switch
	}
}
