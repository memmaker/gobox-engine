package engine

import (
	"slices"
	"testing"

	"gobox/game"
)

// TestBR1SheetFacts: what a BR1 member's card and sheet publish, as the original's sheet shows it
// (viewchar.pas DisplayChar, PrintCombatStats; menucore.pas PrintArmorClass: Abs(armorClass - 60)).
func TestBR1SheetFacts(t *testing.T) {
	e := br1Begun(t)
	ex, ok := e.build().(game.Exploring)
	if !ok {
		t.Fatalf("state %T", e.build())
	}
	m := ex.Party[0]
	pg := e.sheetPage(e.partyIDs()[0])
	if m.Name != "FLAVIUS" || m.Class != "Warrior" || m.Status != "Okay" || m.AC != 2 || m.Health != 0 {
		t.Errorf("card: %q %q %q AC %d health %d", m.Name, m.Class, m.Status, m.AC, m.Health)
	}
	if len(m.SlotsMax) != 0 || !slices.Equal(m.Levels, []int{2}) {
		t.Errorf("a ruleset without spells: slots %v levels %v", m.SlotsMax, m.Levels)
	}
	if pg.Sex != "Male" || pg.Race != "Desert Runner" || pg.Class != "Warrior" || pg.Alignment != "" || pg.AC != 2 || pg.RearAC != 8 || pg.THAC0 != 16 || pg.ShowMods {
		t.Errorf("sheet: %+v", pg)
	}
	if len(pg.Rules) != 9 || pg.Rules[0].Name != "Tech" || pg.Rules[0].Value != "15" || pg.Rules[3].Name != "Battle Tactics" || pg.Rules[3].Value != "47" {
		t.Errorf("tech and the career's skills: %+v", pg.Rules)
	}
}
