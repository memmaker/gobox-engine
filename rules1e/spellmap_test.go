package rules1e

import (
	"testing"

	"gobox/rules"
)

// TestGameSpellIDs: each game's map covers every number of its spell table
// (cotab 1..100, sotsb 1..117, pod 1..126) and each id comes back as its
// number (one number per id, so Export is exact).
func TestGameSpellIDs(t *testing.T) {
	for name, g := range map[string]struct {
		m, gone SpellMap
		last    byte
	}{"cotab": {CotabSpellIDs, CotabSpellsGone, 100}, "sotsb": {SotsbSpellIDs, nil, 117}, "pod": {PodSpellIDs, PodSpellsGone, 126}} {
		tb := Tables{SpellIDs: g.m, SpellsGone: g.gone}
		if len(g.m)+len(g.gone) != int(g.last) {
			t.Errorf("%s: %d numbers, want %d", name, len(g.m), g.last)
		}
		for b := byte(1); b <= g.last; b++ {
			id, err := tb.spellID(b)
			if _, ok := tb.SpellsGone[b]; ok {
				continue // read as another id (or none), never written back
			}
			if err != nil || id < 1 || id > LastSpellID {
				t.Fatalf("%s: %d -> %d (%v)", name, b, id, err)
			}
			if back, err := tb.gameSpell(id); err != nil || back != b {
				t.Errorf("%s: %d -> %d -> %d (%v)", name, b, id, back, err)
			}
		}
	}
}

// TestRaceMap: every race byte of every layout comes back as it was, and
// SotSB's byte 3 is a dwarf (0: its placeholder row, kept as the monster's number).
func TestRaceMap(t *testing.T) {
	for _, l := range []rules.RecordLayout{Layout285, Layout422, Layout439, Layout510} {
		for b := range 256 {
			race, monster := RaceOfRecord(l, byte(b))
			if back, ok := recordRaceByte(l, race, monster); !ok || back != byte(b) {
				t.Errorf("layout %d: %d -> %q %d -> %d (%v)", l, b, race, monster, back, ok)
			}
		}
	}
	if race, _ := RaceOfRecord(Layout439, 3); race != RaceDwarf {
		t.Errorf("silver blades byte 3: %q", race)
	}
	if _, ok := recordRaceByte(Layout439, RaceHalfOrc, 0); ok {
		t.Error("a half-orc has a silver blades number")
	}
}

// TestSpellByte: a memory byte splits into the spell and SpellPending and
// back, for every byte.
func TestSpellByte(t *testing.T) {
	for b := range 256 {
		s := SpellOfByte(byte(b))
		if s.Byte() != byte(b) || s.Pending() != (b >= 0x80) || int(s.Spell()) != b&0x7F {
			t.Fatalf("byte %#x: %#x", b, s)
		}
	}
}

// TestGoneSpellsOut: a removed spell (POR's 56) leaves a scroll's slot and a
// wand's own spell blank, as Import and Export make them; an item without
// one is untouched.
func TestGoneSpellsOut(t *testing.T) {
	r := &Rules{S: &RuleState{Items: make([]ItemBase, ItemWand+1)}}
	r.S.Items[ItemMageScroll].Slot = 0xB // scrolls: IsScroll's slots 11..13
	scroll := Item{Type: ItemMageScroll, Spells: [3]ItemPower{56, 5, 0}}
	wand := Item{Type: ItemWand, Spells: [3]ItemPower{0, 56, 0}}
	keep := Item{Type: ItemWand, Spells: [3]ItemPower{0, 5, 0}}
	if !r.S.IsScroll(&scroll) || !castsSpell(r.S, &wand) {
		t.Fatal("fixture: not a scroll and a spell-casting wand")
	}
	items := []Item{scroll, wand, keep}
	if err := r.GoneSpellsOut(items); err != nil {
		t.Fatal(err)
	}
	if items[0].Spells != [3]ItemPower{0, 5, 0} || items[1].Spells[1] != 0 || items[2] != keep {
		t.Errorf("got %v %v %v", items[0].Spells, items[1].Spells, items[2].Spells)
	}
}
