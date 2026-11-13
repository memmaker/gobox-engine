package rules1e

import "testing"

// TestLearnablePOR: Scribe and PlanSpell take POR's spells 1..56 only, as
// its 56-entry book did (57..67 are powers); the book lists 1..55.
func TestLearnablePOR(t *testing.T) {
	s := RuleState{Tables: &POR}
	c := &Character{}
	for i := 1; i <= int(LastSpellID); i++ {
		if s.learnable(c, SpellID(i)) != (i <= 56) {
			t.Errorf("spell %d learnable %v", i, !(i <= 56))
		}
		if s.inBook(c, SpellID(i)) != (i <= 0x37) { // POR's book lists 1..55 (Restoration, 56, has no slot level)
			t.Errorf("spell %d in book %v", i, !(i <= 0x37))
		}
	}
}

// TestItemSpellCode: POR's shifted bytes and SotSB/POD's seven bits decode
// to the game's number and encode back byte-equal, SotSB/POD's bit 7 kept
// as ItemSpellScribe; COTAB's 0x3F for 40 (stored shifted) back as 0x3F.
func TestItemSpellCode(t *testing.T) {
	for _, c := range []struct {
		code      *ItemSpellCode
		b, n, out ItemPower
		bits      ItemSpellBits
	}{{nil, 0x52, 59, 0x52, 0}, {nil, 0x2F, 47, 0x2F, 0}, {&SevenBitItemSpell, 0x85, 5, 0x85, ItemSpellScribe}, {&SevenBitItemSpell, 0x60, 0x60, 0x60, 0}, {nil, 0x3F, 40, 0x3F, 0}} {
		tb := Tables{ItemSpell: c.code}
		id, bits, shifted, err := tb.itemSpellID(c.b)
		if err != nil || id != c.n || bits != c.bits {
			t.Errorf("%#x -> %d %#x (%v), want %d %#x", c.b, id, bits, err, c.n, c.bits)
		}
		if back, err := tb.gameItemSpell(id, bits, shifted); err != nil || back != c.out {
			t.Errorf("%d -> %#x (%v), want %#x", id, back, err, c.out)
		}
	}
}

// TestItemMendsPerGame: POR mends its broken ring; a game without
// itemMends (cotab, sotsb, pod) leaves the same record as it is.
func TestItemMendsPerGame(t *testing.T) {
	ring := Item{Type: ItemProtectionRing, Plus: 2}
	por, other := ring, ring
	POR.MendItem(&por)
	(&Tables{}).MendItem(&other)
	if por.Plus2 != 2 || other.Plus2 != 0 {
		t.Errorf("POR plus2 %d, other %d", por.Plus2, other.Plus2)
	}
}
