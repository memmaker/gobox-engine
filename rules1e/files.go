package rules1e

// Character file I/O, ported from ovr18. The records are typed structs in
// the DOS file layout, so the original SAVE files load and save byte for byte.

import "slices"

const (
	ItemRecSize   = 0x3F
	AffectRecSize = 9
)

// Letter is a character of a DOS string buffer; WordID an item name word
// (ItemWords).
type (
	Letter uint8
	WordID uint8
)

// ItemPower is an item's +0x3C..0x3E: a scroll's spells, a wand's spell or
// charges, or (an item that stands in for an affect, +0x3E > 0x7F) the
// affect handler and the affect it grants, as unified affect ids in memory.
// Files hold a byte (ItemFile).
type ItemPower uint16

// Item and Affect are the raw list nodes (next pointers live in the slices).
// Item is the 0x3F-byte DOS item record in memory; ItemFile is its file
// layout (encoding/binary).
type Item struct {
	PName   [0x29]Letter // 0x00 shown name, Pascal string (built by ItemName; the rules read its stale bytes too, itemAffect)
	Pad29   byte         // 0x29
	_       uint32       // 0x2A next pointer (far)
	Type    ItemType     // 0x2E base item type (index into the DS:5582 records)
	Words   [3]WordID    // 0x2F name word ids into DS:10A7
	Plus    byte         // 0x32 signed
	Plus2   byte         // 0x33
	Readied byte         // 0x34
	Hidden  byte         // 0x35 bit 3-i: word i not known yet
	Cursed  byte         // 0x36 non-zero: cursed (TItemRec.cursed)
	Weight  uint16       // 0x37
	Count   byte         // 0x39
	Price   uint16       // 0x3A
	Spells  [3]ItemPower // 0x3C a scroll's spells, +0x80 = being scribed
	// SpellBits is the +0x3D bits outside the game's ItemSpellCode, kept
	// from import to Export (no file field of its own).
	SpellBits ItemSpellBits `json:",omitempty"`
	// SpellShifted: +0x3D held a number up to its ItemSpellCode's Above
	// shifted (COTAB's data has 63 for 40; the game reads both alike),
	// kept so Export writes the byte back.
	SpellShifted bool `json:",omitempty"`
}

// ItemSpellBits is the bits of an item's +0x3D its spell code masks off.
// ItemSpellScribe is SotSB's and POD's bit 7: their record names +0x3C and
// +0x3D scribe1/scribe2 (sotsb-re types.pas), the scroll's +0x80 being
// scribed; on an item casting its own spell the originals mask it off
// (viewchar.pas UseItem, O100:27B6) and nothing reads it.
type ItemSpellBits byte

const ItemSpellScribe ItemSpellBits = 0x80

// ItemFile is Item in its file layout (.itm, ITEM and MON ITM blocks).
type ItemFile struct {
	PName   [0x29]Letter
	Pad29   byte
	_       uint32
	Type    ItemType
	Words   [3]WordID
	Plus    byte
	Plus2   byte
	Readied byte
	Hidden  byte
	Cursed  byte
	Weight  uint16
	Count   byte
	Price   uint16
	Spells  [3]PowerFile
}

// PowerFile is an ItemPower as its file byte.
type PowerFile uint8

// Item is f in memory; File its inverse (affect numbers stay the game's
// bytes both ways: the AffectMap translates them at Import and Export).
func (f ItemFile) Item() Item {
	return Item{PName: f.PName, Pad29: f.Pad29, Type: f.Type, Words: f.Words, Plus: f.Plus, Plus2: f.Plus2, Readied: f.Readied, Hidden: f.Hidden, Cursed: f.Cursed, Weight: f.Weight, Count: f.Count, Price: f.Price,
		Spells: [3]ItemPower{ItemPower(f.Spells[0]), ItemPower(f.Spells[1]), ItemPower(f.Spells[2])}}
}

func (it Item) File() ItemFile {
	return ItemFile{PName: it.PName, Pad29: it.Pad29, Type: it.Type, Words: it.Words, Plus: it.Plus, Plus2: it.Plus2, Readied: it.Readied, Hidden: it.Hidden, Cursed: it.Cursed, Weight: it.Weight, Count: it.Count, Price: it.Price,
		Spells: [3]PowerFile{PowerFile(it.Spells[0]), PowerFile(it.Spells[1]), PowerFile(it.Spells[2])}}
}

// Affect is a spell effect or condition on a character.
type Affect struct {
	AffectRecord
	Item *Item // set when an item stands in for the affect (ovr13 item handlers)
}

// AffectRecord is the 9-byte DOS affect node in its file layout.
type AffectRecord struct {
	Type AffectType // index into the ovr13 handler table
	Time uint16     // remaining duration, 0 = permanent
	Data byte
	Call byte   // non-zero: the handler runs when the affect is removed
	_    uint32 // next pointer (far)
}

// AffectFile is an AffectRecord in its file layout (.spc, MON SPC blocks):
// the affect number is the game's byte there.
type AffectFile struct {
	Type byte
	Time uint16
	Data byte
	Call byte
	_    uint32
}

// Record is f in memory; File its inverse (the number is the game's byte
// both ways: the AffectMap translates it at Import and Export).
func (f AffectFile) Record() AffectRecord {
	return AffectRecord{Type: AffectType(f.Type), Time: f.Time, Data: f.Data, Call: f.Call}
}

func (a AffectRecord) File() AffectFile {
	return AffectFile{Type: byte(a.Type), Time: a.Time, Data: a.Data, Call: a.Call}
}

func (e *RuleState) PartyCount() int { return int(e.Game.PartyCount) }
func (e *RuleState) SetPartyCount(n int) {
	e.Game.PartyCount = uint16(n)
}

// AddToParty = ovr18:1446 (c): append, give the lowest free order 0..7;
// the newcomer's coins go to the party's purse.
func (e *RuleState) AddToParty(c *Character) {
	c.Order = 0xFF
	e.Party = append(e.Party, c)
	var used [8]bool
	for _, m := range e.Party {
		if m.Order < 8 {
			used[m.Order] = true
		}
	}
	e.Who = len(e.Party) - 1
	o := 0
	for o < 8 && used[o] {
		o++
	}
	c.Order = byte(o)
	e.SetPartyCount(e.PartyCount() + 1)
	e.takeRecordCoins(c)
	// extra: old records' affect 0x5A goes, its bonus is in the stored save (ClassStats)
	c.Affects = slices.DeleteFunc(c.Affects, func(a *Affect) bool { return a.Type == affSaveBonusVsType0 })
	e.RecomputeCharacter(c) // ovr24:0000
	e.ApplyItems(c)         // extra: the unreadied items go into the party inventory
}

// ScrollScribing marks a scroll's spell (Item.Spells) being scribed: the
// DOS byte's +0x80, moved above every unified spell id (as SpellPending).
const ScrollScribing ItemPower = 0x8000

// Scribing reports whether scroll spell p is being scribed (p > 0x80 in
// the DOS byte: a spell, flagged).
func (p ItemPower) Scribing() bool { return p&ScrollScribing != 0 && p.Spell() != 0 }

// Spell is scroll spell (or a wand's spell) p without the scribing flag.
func (p ItemPower) Spell() SpellID { return SpellID(p &^ ScrollScribing) }
