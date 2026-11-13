package engine

import (
	"gobox/char"
	"gobox/rules1e"
	"testing"
)

// Paladin and ranger rules ported from CotAB.
func TestPaladinRanger(t *testing.T) {
	e := &Engine{}
	e.initVM()
	pal := &rules1e.Character{}
	pal.Levels[rules1e.ClsPaladin] = 7
	e.state().RecomputeCharacter(pal)
	if pal.BaseAttack[0] != 3 || pal.Saves[0] != 8 || !rules1e.FighterKind(pal) {
		t.Errorf("paladin 7: attacks %d saves %v", pal.BaseAttack[0], pal.Saves)
	}
	rng := &rules1e.Character{}
	rng.Levels[rules1e.ClsRanger] = 7
	e.state().RecomputeCharacter(rng)
	if rng.BaseAttack[0] != 2 {
		t.Errorf("ranger 7: attacks %d", rng.BaseAttack[0])
	}
	if rules1e.XPForLevel(rules1e.CoreClassPaladin, 9) != -1 || rules1e.XPForLevel(rules1e.CoreClassRanger, 8) != -1 || rules1e.XPForLevel(rules1e.CoreClassRanger, 7) != 90001 {
		t.Error("level caps")
	}
	if len(rules1e.RaceCareers(rules1e.CoreRaceHuman)) != 6 || len(rules1e.RaceCareers(rules1e.CoreRaceHalfElf)) != 13 || len(rules1e.CareerAlignments(rules1e.CareerRanger)) != 3 {
		t.Error("class lists")
	}
	if !rules1e.LevelBlocked(4, 0x11, rules1e.ClsRanger, 7) || rules1e.LevelBlocked(7, 0x11, rules1e.ClsRanger, 7) {
		t.Error("half-elf ranger limit")
	}
}

// A paladin saved by the port still reads as CotAB's ConvertPoolRadPlayer
// expects (ovr017): the 0x11D record at the same offsets, 9-byte affect
// records, and none of the new affects among the ones CotAB keeps.
func TestPaladinCotabTransfer(t *testing.T) {
	c := &rules1e.Character{}
	c.Class, c.Alignment, c.Exp = 3, 0, 175001
	c.Levels[rules1e.ClsPaladin] = 8
	b := char.SheetBytes(c.Sheet)
	if len(b) != 0x11D || b[0x2F] != 3 || b[0x96+3] != 8 || b[0xA0] != 0 || b[0xAC] != 0x99 || b[0xAE] != 0x02 {
		t.Fatalf("record % x", b[0x2F:0x30])
	}
	cotabKeeps := map[rules1e.AffectType]bool{18: true, 26: true, 47: true, 48: true, 97: true, 107: true, 124: true} // ovr017 asc_49280
	for _, typ := range []rules1e.AffectType{rules1e.AffProtEvil, rules1e.AffLaidOnHands, rules1e.AffCuredDisease} {
		if cotabKeeps[typ] {
			t.Errorf("CotAB would import affect %#x", typ)
		}
	}
	if n := len(char.WriteRecords([]*rules1e.AffectFile{{Type: byte(rules1e.AffLaidOnHands), Time: 1440}})); n != 9 {
		t.Errorf("affect record %d bytes", n)
	}
}
