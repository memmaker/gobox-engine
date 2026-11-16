package engine

import (
	"os"
	"slices"
	"testing"

	"gobox/game"
	"gobox/rules"
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
	if len(pg.Rules) != 8 || pg.Rules[2].Name != "Battle Tactics" || pg.Rules[2].Value != "47" {
		t.Errorf("the career's skills: %+v", pg.Rules)
	}
	var names []string
	for _, a := range pg.Abilities {
		names = append(names, a.Label)
	}
	if !slices.Equal(names, []string{"STR", "DEX", "CON", "INT", "WIS", "CHA", "TCH"}) || pg.Abilities[6].Name != "Tech" || pg.Abilities[6].Value != 15 {
		t.Errorf("the seven abilities in the original's order: %+v", pg.Abilities)
	}
}

// TestBR1Roster: a new BR1 save folder, seeded from the adventure's saves/ as cmd/gobox homeSaves does, lists the
// shipped characters (the originals' CHRDATAn.SAV, the first of each name), each numbered once, and Add joins them.
func TestBR1Roster(t *testing.T) {
	e := br1Engine(t)
	e.Open("")
	e.titleDone()
	e.Advance()
	list := e.BuildCharList()
	if want := []string{"FLAVIUS", "Rook", "Kade", "Ace", "Vex", "Gizmo", "Doc"}; !slices.Equal(list, want) { // the first CHRDATA of each name, in file order
		t.Fatalf("roster %v, want %v", list, want)
	}
	seen := map[rules.RosterID]bool{}
	for _, n := range list {
		b, err := os.ReadFile(e.rosterFile(n))
		if err != nil || !e.fileFormat().IsRecord(b) {
			t.Fatalf("%s: not a record (%v)", n, err)
		}
		if id := e.fileFormat().RosterOf(b); id == rules.NoRoster || seen[id] {
			t.Fatalf("%s: roster number %d", n, id)
		} else {
			seen[id] = true
		}
	}
	if _, err := e.doNow(game.AddCharacter{Roster: len(list) - 1}); err != nil || e.partySize() != 1 {
		t.Fatalf("Add the last: %v, party %d", err, e.partySize())
	}
}

// TestBR1PurseInCredits: the state's purse counts in the game's own coin (game.Data.Coin, BR's credits), which
// is what the front shows; it showed 0 while the purse was published as gold and copper only.
func TestBR1PurseInCredits(t *testing.T) {
	e := br1Begun(t)
	have := int(e.purse(false).Copper)
	if got := commonOf(e.build()).Purse[game.Credits]; e.gameData.Coin != game.Credits || have == 0 || got != have || commonOf(e.build()).Purse[game.Copper] != 0 {
		t.Fatalf("purse %v under %q, the party holds %d", commonOf(e.build()).Purse, e.gameData.Coin, have)
	}
}
