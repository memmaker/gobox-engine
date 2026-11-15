package importer

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"gobox/adventure"
	"gobox/ecl"
	"gobox/game"
	"gobox/rules"
	"gobox/rules1e"
)

// TestImportedTablesLoad: the imported folder's tables load as pool and rules1e
// read them (pool.ReadGameTables, char.Rules1e over rules.json's tables) and equal COTAB's
// own constants; every scripted journal entry has its text in the data folder's book.
func TestImportedTablesLoad(t *testing.T) {
	dos := filepath.Join("..", "data", "cotab", "sources", "dos")
	if _, err := os.Stat(dos); err != nil {
		t.Skip("no originals:", err)
	}
	root := t.TempDir()
	if err := Run("cotab", dos, root); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, "cotab")

	g, err := adventure.ReadGameTables(dst)
	if err != nil {
		t.Fatal(err)
	}
	if g.Script != ecl.V2 || g.AddNPCIcon || g.ScriptSounds[8].Cue != rules.CueStep || g.ScriptSounds[10].Cue != rules.CueFireball || len(g.ScriptSounds) != 2 {
		t.Errorf("game tables %+v", g)
	}
	if want := []ecl.CombatScreen{ecl.CombatShop, ecl.CombatTemple}; !slices.Equal(g.Script.Behaviour.NoMonsterCombat, want) {
		t.Errorf("NoMonsterCombat %v, want %v", g.Script.Behaviour.NoMonsterCombat, want)
	}

	rj := filepath.Join(dst, "data", "rules.json")
	l, err := rules1e.ReadLayout(filepath.Join(dst, "data", rules1e.LayoutFile))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(l.Races, CotabLayout1e.Races) || l.SlotStride != CotabLayout1e.SlotStride || l.SlotLevels != CotabLayout1e.SlotLevels || l.SlotTotals != CotabLayout1e.SlotTotals || l.RangerMageColumn != CotabLayout1e.RangerMageColumn {
		t.Errorf("layout %+v, want %+v", l, CotabLayout1e)
	}
	tb, err := rules1e.POR.LoadRules(rj, l)
	if err != nil {
		t.Fatal(err)
	}
	if tb.SlotLevels != CotabLayout1e.SlotLevels {
		t.Errorf("slot levels %d", tb.SlotLevels)
	}
	if r, _ := rules1e.Rules1e(noSource{}, nil); r == nil {
		t.Error("char.Rules1e builds no rules")
	}

	book, err := game.LoadJournal(os.DirFS(filepath.Join(dst, adventure.JournalDir)), "JOURNAL.TXT")
	if err != nil {
		t.Fatal(err)
	}
	es, err := ecl.ReadEntries(os.DirFS(filepath.Join(dst, "scripts")))
	if err != nil || len(es) == 0 {
		t.Fatal("entries:", len(es), err)
	}
	for _, en := range es {
		if en.Book != "JOURNAL.TXT" || !book.Has(en.Title) {
			t.Errorf("entry %d (%s %q) has no text", en.ID, en.Book, en.Title)
		}
	}
}
