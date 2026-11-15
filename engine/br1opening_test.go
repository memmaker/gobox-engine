package engine

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
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
	sc, ok := e.build().(game.Scene)
	if !ok || sc.Ask == nil || len(sc.Ask.Replies) != 1 || !strings.HasPrefix(paneText(sc.Common), "AFTER HEARING OF THE BRAVERY OF BUCK ROGERS") {
		t.Fatalf("start: %T %q ask %+v", e.build(), paneText(sc.Common), sc.Ask)
	}
	if e.gs.areaKind != 16 {
		t.Fatalf("block %d, want 16", e.gs.areaKind)
	}
	if p := sc.Picture; p == nil || p.File != "PIC9" || p.Block != 62 { // the window's PIC while its page is up (O124:02D0's PIC9 for 62)
		t.Fatalf("start picture: %+v", p)
	}
	br1Pages(t, e) // the intro
	x, ok := e.build().(game.Exploring)
	if !ok || x.Where.X != 5 || x.Where.Y != 6 || x.Where.Facing != 0 || x.Picture != nil || paneText(x.Common) != "" {
		t.Fatalf("after the intro: %T %+v %+v %q", e.build(), x.Where, x.Picture, paneText(x.Common))
	}
	if err := br1North(e); err != nil {
		t.Fatal(err)
	}
	if c := commonOf(e.build()); c.Where.Y != 5 || !strings.HasPrefix(paneText(c), "AN OFFICER IS COLLAPSED") {
		t.Fatalf("north: %+v %q", c.Where, paneText(c))
	}
	if err := br1North(e); err == nil {
		t.Fatal("walked while a question waits")
	}
	br1Pages(t, e)
	if err := br1North(e); err != nil {
		t.Fatal(err)
	}
	br1Pages(t, e)
	if err := br1North(e); err != nil {
		t.Fatal(err)
	}
	if c := commonOf(e.build()); !strings.HasPrefix(paneText(c), "THE HEAT FROM THE EXPLOSIONS") {
		t.Fatalf("tarmac: %+v %q", c.Where, paneText(c))
	}
	if _, err := e.doNow(game.Reply{}); err != nil {
		t.Fatal(err)
	}
	if c := commonOf(e.build()); c.Where.Y != 4 {
		t.Fatalf("not driven back: %+v", c.Where)
	}
}

// br1Pages answers the pages and pauses up until the game waits for a move.
func br1Pages(t *testing.T, e *Engine) {
	t.Helper()
	for n := 0; commonOf(e.build()).Ask != nil; n++ {
		if n > 80 {
			t.Fatalf("the pages do not end: %q", paneText(commonOf(e.build())))
		}
		var a game.Action = game.Reply{}
		if slices.Contains(commonOf(e.build()).Allowed.Plain, game.Action(game.ContinueAction{})) { // a Delay's pause
			a = game.ContinueAction{}
		}
		if _, err := e.doNow(a); err != nil {
			t.Fatalf("page %d %q: %v", n, paneText(commonOf(e.build())), err)
		}
	}
}

// br1North is gobox's step north with the square's event run.
func br1North(e *Engine) error {
	if err := stepBy(e, 0); err != nil {
		return err
	}
	if e.ext.owed == nil {
		return nil
	}
	_, err := e.doNow(game.RunEvent{})
	return err
}

// paneText is the text pane's words, the wrapping's spaces aside.
func paneText(c game.Common) string {
	return strings.Join(strings.Fields(strings.Join(c.Text, " ")), " ")
}
