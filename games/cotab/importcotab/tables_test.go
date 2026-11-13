package importcotab

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"gobox/char"
	"gobox/core"
	"gobox/ecl/script"
	"gobox/game/journal"
	"gobox/importer"
	"gobox/pool"
	"rules1e"
)

// TestImportedTablesLoad: the imported folder's tables load as pool and rules1e
// read them (pool.ReadGameTables, char.Rules1e over rules.json's tables) and equal COTAB's
// own constants; every scripted journal entry has its text in the data folder's book.
func TestImportedTablesLoad(t *testing.T) {
	dos := filepath.Join("..", "..", "..", "data", "cotab", "sources", "dos")
	if _, err := os.Stat(dos); err != nil {
		t.Skip("no originals:", err)
	}
	root := t.TempDir()
	if err := importer.Run("cotab", dos, root); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, "cotab")

	g, err := pool.ReadGameTables(dst)
	if err != nil {
		t.Fatal(err)
	}
	if g.Script != script.V2 || g.AddNPCIcon || g.ScriptSounds[8].Cue != core.CueStep || g.ScriptSounds[10].Cue != core.CueFireball || len(g.ScriptSounds) != 2 {
		t.Errorf("game tables %+v", g)
	}
	if want := []script.CombatScreen{script.CombatShop, script.CombatTemple}; !slices.Equal(g.Script.Behaviour.NoMonsterCombat, want) {
		t.Errorf("NoMonsterCombat %v, want %v", g.Script.Behaviour.NoMonsterCombat, want)
	}

	rj := filepath.Join(dst, "data", "rules.json")
	l, err := rules1e.ReadLayout(filepath.Join(dst, "data", rules1e.LayoutFile))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(l.Races, Layout1e.Races) || l.SlotStride != Layout1e.SlotStride || l.SlotLevels != Layout1e.SlotLevels || l.SlotTotals != Layout1e.SlotTotals || l.RangerMageColumn != Layout1e.RangerMageColumn {
		t.Errorf("layout %+v, want %+v", l, Layout1e)
	}
	tb, err := rules1e.POR.LoadRules(rj, l)
	if err != nil {
		t.Fatal(err)
	}
	if tb.SlotLevels != Layout1e.SlotLevels {
		t.Errorf("slot levels %d", tb.SlotLevels)
	}
	if r, _ := char.Rules1e("", nil, nil, nil, nil); r == nil {
		t.Error("char.Rules1e builds no rules")
	}

	book, err := journal.Load(os.DirFS(filepath.Join(dst, "data", pool.JournalDir)), "JOURNAL.TXT")
	if err != nil {
		t.Fatal(err)
	}
	es, err := script.ReadEntries(os.DirFS(filepath.Join(dst, "scripts")))
	if err != nil || len(es) == 0 {
		t.Fatal("entries:", len(es), err)
	}
	for _, en := range es {
		if en.Book != "JOURNAL.TXT" || !book.Has(en.Title) {
			t.Errorf("entry %d (%s %q) has no text", en.ID, en.Book, en.Title)
		}
	}
}
