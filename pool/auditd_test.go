package pool

import (
	"rules1e"
	"testing"
)

// The broken data records are mended as they are read.
func TestFixItemRecord(t *testing.T) {
	ft := &rules1e.Item{Type: rules1e.ItemLongSword, Words: [3]rules1e.WordID{0xF1, 0xA3, 0x24}, Spells: [3]rules1e.ItemPower{0, 0x16, 0}}
	ring := &rules1e.Item{Type: rules1e.ItemFeatherFallRing, Words: [3]rules1e.WordID{0xEE, 0xA7, 0x42}}
	br := &rules1e.Item{Type: rules1e.ItemBracers, Plus: 6, Words: [3]rules1e.WordID{0, 0xDC, 0x4F}}
	for _, it := range []*rules1e.Item{ft, ring, br} {
		rules1e.POR.MendItem(it)
	}
	if ft.Spells[2] != rules1e.ItemPower(rules1e.AffItem0) || ft.Spells[1] != rules1e.ItemPower(rules1e.AffBonusByTargetRace) || ring.Spells[1] != rules1e.ItemPower(rules1e.AffFeatherFall) || br.Plus != 8 {
		t.Errorf("flame tongue %v, ring %v, bracers %d", ft.Spells, ring.Spells, br.Plus)
	}
}

// The thieves' well finds Feather Fall on its climber in the effect table.
func TestEffectTable(t *testing.T) {
	e := &Engine{}
	e.initVM()
	a, b := &rules1e.Character{}, &rules1e.Character{}
	e.state().Party = []*rules1e.Character{a, b}
	e.state().AddAffect(a, rules1e.AffBless, 0, 0xFF, 0)
	e.state().AddAffect(b, rules1e.AffFeatherFall, 0, 0xFF, 0)
	if v, _ := e.state().EffectTable(1); v != uint16(rules1e.AffFeatherFall) {
		t.Fatalf("entry 1 = %d", v)
	}
	if v, _ := e.state().EffectTable(65); v != 1 {
		t.Errorf("owner of entry 1 = %d", v)
	}
}

// Spells held past the slots a drain left are forgotten.
func TestTrimMemorized(t *testing.T) {
	e := &Engine{}
	e.initVM()
	c := &rules1e.Character{}
	c.Levels[rules1e.ClsMage], c.SpellSlots = 1, rules1e.SpellSlots{rules1e.SpellClassMage: {1}}
	c.Memorized[0], c.Memorized[1] = rules1e.SpMagicMissile, rules1e.SpMagicMissile|rules1e.SpellPending
	e.state().TrimMemorized(c)
	if c.Memorized[0] != rules1e.SpMagicMissile || c.Memorized[1] != 0 {
		t.Errorf("memorized %v", c.Memorized[:2])
	}
}
