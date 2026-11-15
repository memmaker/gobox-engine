package engine

import (
	"os"
	"path/filepath"
	"testing"

	"gobox/adventure"
	"gobox/game"
	"gobox/importer"
	"gobox/rulesxxvc"
)

// br1Engine is the shared engine on BR1's folder, imported from the originals
// into a temp dir (importer.ImportBREngine), with the XXVc ruleset. It skips
// without the originals.
func br1Engine(t *testing.T) *Engine {
	t.Helper()
	dos := filepath.Join("..", "data", "br1", "sources", "dos")
	if _, err := os.Stat(filepath.Join(dos, "START.EXE")); err != nil {
		t.Skip("no BR1 originals:", err)
	}
	adv := t.TempDir()
	if err := importer.ImportBREngine("br1", importer.BR1EXE, dos, adv); err != nil {
		t.Fatal(err)
	}
	tables, err := adventure.ReadGameTables(adv)
	if err != nil {
		t.Fatal(err)
	}
	e, err := New(rulesxxvc.Builder, tables, adv, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	seed := uint32(7)
	e.FixedSeed = &seed
	return e
}

// br1Roster saves the originals' shipped characters (the import's
// data/characters) into the engine's roster, as a character saved from the
// party menu is.
// ponytail: the test seeds the roster; who seeds a new save folder in play is the team menu's slice (REFACTOR.md B1f).
func br1Roster(t *testing.T, e *Engine) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(e.adv, adventure.DataDir, importer.CharactersDir, "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatal("no imported characters", err)
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		id, err := e.ruleset().Import(b)
		if err != nil {
			t.Fatal(f, err)
		}
		if err := e.SaveCharacter(id, ""); err != nil {
			t.Fatal(f, err)
		}
	}
}

// TestBR1Opening: BR1's opening on the shared engine (REFACTOR.md B1f1; the
// oracle is buckrogers-go rules/adv_test.go TestOpening).
func TestBR1Opening(t *testing.T) {
	e := br1Engine(t)
	e.Open("")
	e.titleDone()
	e.Advance()
	if _, ok := e.build().(game.Assembling); !ok {
		t.Fatalf("first wait is %T, want the party menu", e.build())
	}
	br1Roster(t, e)
	if _, err := e.doNow(game.AddCharacter{Roster: 0}); err != nil {
		t.Fatal(err)
	}
	if n := e.partySize(); n != 1 {
		t.Fatalf("party of %d after Add", n)
	}
	if _, err := e.doNow(game.Begin{}); err != nil {
		t.Fatal(err)
	}
	// As far as the engine gets today: the party is out, block 16's script started and stops at its first
	// command the engine lacks (V4 76, the area kind). The oracle's assertions follow as the commands land.
	if _, ok := e.build().(game.Exploring); !ok {
		t.Fatalf("after Begin: %T, want game.Exploring", e.build())
	}
}
