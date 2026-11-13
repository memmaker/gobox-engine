package rules1e

import (
	"fmt"
	"slices"
)

// SpellMap is a game's spell numbers as the unified spell ids (POR's numbers,
// and new ids for the spells only another game has, spellids_games.go):
// rules.json spellIds, written by the game's importer. Numbers enter as the
// game's once, at the boundaries, and leave as them on Export; inside
// rules1e only ids exist.
type SpellMap map[byte]SpellID

// spellID is the game's spell number b as an id; nil (POR): b is the id.
func (t *Tables) spellID(b byte) (SpellID, error) {
	if id, gone := t.SpellsGone[b]; gone { // a removed spell: its stand-in, or NoSpell
		return id, nil
	}
	if t.SpellIDs == nil {
		return SpellID(b), nil
	}
	id, ok := t.SpellIDs[b]
	if !ok {
		return 0, fmt.Errorf("rules1e: spell number %d is not in the game's spellIds", b)
	}
	return id, nil
}

// gameSpell is spellID's inverse.
func (t *Tables) gameSpell(id SpellID) (byte, error) {
	if t.SpellIDs == nil {
		return byte(id), nil
	}
	for b, v := range t.SpellIDs {
		if v == id {
			return b, nil
		}
	}
	return 0, fmt.Errorf("rules1e: spell id %d has no number in the game", id)
}

// spellsByID is a game's spell table (by its numbers) by unified id: an id
// the game lacks has an empty record, as POR's record 0.
// A removed spell's number (gone) has no record.
func spellsByID(table [][]SpellProp, m, gone SpellMap) ([][]SpellProp, error) {
	out := make([][]SpellProp, LastSpellID+1)
	for i := range out {
		out[i] = make([]SpellProp, spellRecord)
	}
	for n, rec := range table {
		if n == 0 {
			continue
		}
		id, ok := m[byte(n)]
		if _, g := gone[byte(n)]; g {
			continue
		}
		if !ok {
			return nil, fmt.Errorf("rules1e: spell number %d of the spell table is not in spellIds", n)
		}
		out[id] = rec
	}
	return out, nil
}

// spellRecord is a spell table record's length.
const spellRecord = 16

// raceID is the game's race number n as rules1e's RaceID (raceRow's order);
// nil Races (POR): n is the id.
func (t *Tables) raceID(n RaceID) (RaceID, error) {
	if t.Races == nil {
		return n, nil
	}
	if int(n) < len(t.Races) {
		if row, ok := raceRow[t.Races[n]]; ok {
			return RaceID(row), nil
		}
	}
	return 0, fmt.Errorf("rules1e: race number %d has no race in the game", n)
}

// gameRace is raceID's inverse.
func (t *Tables) gameRace(id RaceID) (RaceID, error) {
	if t.Races == nil {
		return id, nil
	}
	for n, race := range t.Races {
		if row, ok := raceRow[race]; ok && row == int(id) {
			return RaceID(n), nil
		}
	}
	return 0, fmt.Errorf("rules1e: race %d has no number in the game", id)
}

// scrollsIn rewrites the spells of items between the game's file bytes
// and unified ids: a scroll's (number, +0x80 being scribed) with
// ScrollScribing, and the spell its own use casts (castsSpell) from the
// item-power byte (the game's ItemSpellCode).
func (e *RuleState) scrollsIn(items []Item, toIDs bool) error {
	t := e.T()
	for i := range items {
		it := &items[i]
		if castsSpell(e, it) {
			var err error
			if toIDs {
				it.Spells[1], it.SpellBits, it.SpellShifted, err = t.itemSpellID(it.Spells[1])
			} else {
				it.Spells[1], err = t.gameItemSpell(it.Spells[1], it.SpellBits, it.SpellShifted)
				it.SpellBits, it.SpellShifted = 0, false
			}
			if err != nil {
				return err
			}
			continue
		}
		if !e.IsScroll(it) {
			continue
		}
		for k, p := range it.Spells {
			if toIDs {
				if p > 0xFF {
					return fmt.Errorf("rules1e: scroll spell %#x is not a byte", p)
				}
				s := SpellOfByte(byte(p))
				id, err := t.spellNumber(s.Spell())
				if err != nil {
					return err
				}
				it.Spells[k] = ItemPower(id)
				if s.Pending() {
					it.Spells[k] |= ScrollScribing
				}
				continue
			}
			n, err := t.gameSpellNumber(p.Spell())
			if err != nil {
				return err
			}
			s := n
			if p&ScrollScribing != 0 {
				s |= SpellPending
			}
			it.Spells[k] = ItemPower(s.Byte())
		}
	}
	return nil
}

// castsSpell tells whether it is no scroll and its own use casts the spell
// its +0x3D names (wands, potions: +0x3E below the affect powers).
func castsSpell(e *RuleState, it *Item) bool {
	return !e.IsScroll(it) && it.Spells[1] > 0 && it.Spells[2] < 0x80
}

// ItemSpellCode is how a game's item +0x3D byte holds its spell number:
// the bits of Mask, those above Above stored plus Shift. POR and COTAB
// (ovr20:1A8A, aiUseItem) shift those past 0x38 by 0x17; SotSB and POD
// (O100:27B6) take the low seven bits.
type ItemSpellCode struct {
	Mask  byte `json:"mask"`
	Above byte `json:"above"`
	Shift byte `json:"shift"`
}

// PorItemSpell is POR's and COTAB's code; SevenBitItemSpell SotSB's and POD's.
var (
	PorItemSpell      = ItemSpellCode{Mask: 0xFF, Above: 0x38, Shift: 0x17}
	SevenBitItemSpell = ItemSpellCode{Mask: 0x7F, Above: 0x7F}
)

func (t *Tables) itemSpellCode() ItemSpellCode {
	if t.ItemSpell == nil {
		return PorItemSpell
	}
	return *t.ItemSpell
}

// number is the game's spell number byte b holds.
func (k ItemSpellCode) number(b byte) byte {
	b &= k.Mask
	if b > k.Above {
		return b - k.Shift
	}
	return b
}

// itemSpellID is an item's +0x3D (a byte in the game's code) as a unified
// id, the bits outside the code's Mask and whether a number up to Above was
// stored shifted (both kept for Export).
func (t *Tables) itemSpellID(p ItemPower) (ItemPower, ItemSpellBits, bool, error) {
	if p > 0xFF {
		return 0, 0, false, fmt.Errorf("rules1e: item spell %#x is not a byte", p)
	}
	k := t.itemSpellCode()
	n := k.number(byte(p))
	id, err := t.spellID(n)
	return ItemPower(id), ItemSpellBits(byte(p) &^ k.Mask), byte(p)&k.Mask > k.Above && n <= k.Above, err
}

// gameItemSpell is itemSpellID's inverse: numbers up to Above are stored
// as they are (as POR's data has 34..56), later ones shifted.
func (t *Tables) gameItemSpell(p ItemPower, bits ItemSpellBits, shifted bool) (ItemPower, error) {
	n, err := t.gameSpell(SpellID(p))
	if err != nil {
		return 0, err
	}
	k := t.itemSpellCode()
	b := int(n)
	if n > k.Above || shifted {
		b += int(k.Shift)
	}
	if b > 0xFF || k.number(byte(b)) != n || byte(bits)&k.Mask != 0 {
		return 0, fmt.Errorf("rules1e: item spell number %d (bits %#x) cannot be stored", n, bits)
	}
	return ItemPower(byte(b) | byte(bits)), nil
}

// spellNumber and gameSpellNumber are spellID and gameSpell for a
// record's spell field: 0 (none) stays 0.
func (t *Tables) spellNumber(n SpellID) (SpellID, error) {
	if n == 0 {
		return 0, nil
	}
	if n > 0xFF {
		return 0, fmt.Errorf("rules1e: spell number %d is not a byte", n)
	}
	return t.spellID(byte(n))
}

func (t *Tables) gameSpellNumber(id SpellID) (SpellID, error) {
	if id == 0 {
		return 0, nil
	}
	b, err := t.gameSpell(id)
	return SpellID(b), err
}

// sheetSpells rewrites c's memory, plan and spell book between the game's
// numbers and unified ids (conv: spellNumber or gameSpellNumber).
func sheetSpells(s *Sheet, plan *[MemorySlots]SpellID, conv func(SpellID) (SpellID, error)) error {
	for i, v := range s.Memorized {
		id, err := conv(v.Spell())
		if err != nil {
			return fmt.Errorf("memory slot %d: %w", i, err)
		}
		s.Memorized[i] = id | v&SpellPending
	}
	if plan != nil {
		for i, v := range plan {
			id, err := conv(v)
			if err != nil {
				return fmt.Errorf("plan entry %d: %w", i, err)
			}
			plan[i] = id
		}
	}
	var book [LastSpellID]Flag
	for i, f := range s.SpellBook {
		if f == 0 {
			continue
		}
		id, err := conv(SpellID(i + 1))
		if err != nil {
			return fmt.Errorf("spell book: %w", err)
		}
		if id == NoSpell { // a removed spell (Tables.SpellsGone) is not known
			continue
		}
		if int(id) > len(book) {
			return fmt.Errorf("spell book: spell %d out of range", id)
		}
		book[id-1] = f
	}
	s.SpellBook = book
	return nil
}

// porLastSpellID is POR's highest spell number (57..67 the item and monster
// powers): a game without spellIds has the ids 1..porLastSpellID.
const porLastSpellID SpellID = 67

// hasSpell reports whether id is one of the game's spells.
func (t *Tables) hasSpell(id SpellID) bool {
	if t.SpellIDs == nil {
		return id >= 1 && id <= porLastSpellID && SpellIDs[id] != "" // a removed spell (PorSpellsGone) has no unified id
	}
	for _, v := range t.SpellIDs {
		if v == id {
			return true
		}
	}
	return false
}

// SpellClassMap is a game's spell class bytes (its spell table's field 0)
// as class names: rules.json spellClasses, written by the game's importer.
type SpellClassMap map[byte]SpellClassName

// SpellClassName is a SpellClass as rules.json names it.
type SpellClassName string

const (
	ClassNameCleric SpellClassName = "cleric"
	ClassNameMage   SpellClassName = "mage"
	ClassNameDruid  SpellClassName = "druid"
	ClassNamePower  SpellClassName = "power"
)

var spellClassByName = map[SpellClassName]SpellClass{ClassNameCleric: SpellClassCleric, ClassNameMage: SpellClassMage, ClassNameDruid: SpellClassDruid, ClassNamePower: SpellClassPower}

// Class is the SpellClass n names (false: none).
func (n SpellClassName) Class() (SpellClass, bool) {
	c, ok := spellClassByName[n]
	return c, ok
}

// Each later game's class bytes (fold T3d ruling; POR's 0 cleric, 1 mage,
// 2 power are applied in its built-in table, tables1e.go): CotAB and PoD 0 cleric, 1 druid, 2 mage, 3 power; SotSB 0 cleric,
// 1 druid, 3 mage, 4 power.
var (
	CotabSpellClasses = SpellClassMap{0: ClassNameCleric, 1: ClassNameDruid, 2: ClassNameMage, 3: ClassNamePower}
	SotsbSpellClasses = SpellClassMap{0: ClassNameCleric, 1: ClassNameDruid, 3: ClassNameMage, 4: ClassNamePower}
	PodSpellClasses   = SpellClassMap{0: ClassNameCleric, 1: ClassNameDruid, 2: ClassNameMage, 3: ClassNamePower}
)

// classesOf is table (by the game's numbers) with each record's class byte
// as its SpellClass (m); the empty record 0 stays as it is.
func classesOf(table [][]SpellProp, m SpellClassMap) ([][]SpellProp, error) {
	out := make([][]SpellProp, len(table))
	for n, rec := range table {
		out[n] = slices.Clone(rec)
		if n == 0 || len(rec) == 0 {
			continue
		}
		c, ok := m[byte(rec[0])].Class()
		if !ok {
			return nil, fmt.Errorf("rules1e: spell %d's class byte %d is not in spellClasses", n, rec[0])
		}
		out[n][0] = SpellProp(c)
	}
	return out, nil
}

// GoneSpellsOut rewrites file items to what Import and Export make of them
// when an item holds a removed spell (Tables.SpellsGone): a scroll's spell or
// the spell a wand's, potion's or any item's own use casts (castsSpell) goes
// through scrollsIn there and back: its stand-in's number, or 0 (a blank
// slot that casts nothing). Other items are left as they are, so the round
// trip stays checked against the file. TestOriginals expects it.
func (r *Rules) GoneSpellsOut(items []Item) error {
	for i := range items {
		if !r.S.holdsGoneSpell(&items[i]) {
			continue
		}
		one := items[i : i+1]
		if err := r.S.scrollsIn(one, true); err != nil {
			return err
		}
		if err := r.S.scrollsIn(one, false); err != nil {
			return err
		}
	}
	return nil
}

// holdsGoneSpell tells whether the file item it holds a removed spell
// (Tables.SpellsGone) in a slot scrollsIn converts.
func (e *RuleState) holdsGoneSpell(it *Item) bool {
	t := e.T()
	gone := func(n byte) bool { _, ok := t.SpellsGone[n]; return ok }
	if castsSpell(e, it) {
		return it.Spells[1] <= 0xFF && gone(t.itemSpellCode().number(byte(it.Spells[1])))
	}
	if !e.IsScroll(it) {
		return false
	}
	for _, p := range it.Spells {
		if p != 0 && p <= 0xFF && gone(byte(SpellOfByte(byte(p)).Spell())) {
			return true
		}
	}
	return false
}
