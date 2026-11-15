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
	if r, ok := e.layoutField(fieldScriptBlock); !ok || e.host.Get(r) != 16 { // the start script ran with 0: the intro; a loaded game's finds 16
		t.Fatalf("ScriptBlock %d, want 16", e.host.Get(r))
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

// br1Begun is BR1 begun with one roster character, the intro's pages answered: the party stands on the tarmac.
func br1Begun(t *testing.T) *Engine {
	t.Helper()
	e := br1Engine(t)
	e.Open("")
	e.titleDone()
	e.Advance()
	br1Roster(t, e)
	for _, a := range []game.Action{game.AddCharacter{Roster: 0}, game.Begin{}} {
		if _, err := e.doNow(a); err != nil {
			t.Fatalf("%T: %v", a, err)
		}
	}
	br1Pages(t, e)
	return e
}

// TestBR1NewBlock: a new block's start (buckrogers-go TestNewBlock): block 17, Salvation base, has no map and no
// walking; its script greets the party and then asks where to go, again after every answer that ends its run.
func TestBR1NewBlock(t *testing.T) {
	e := br1Begun(t)
	e.DevAt(17, 0, 0)
	c := commonOf(e.build())
	if e.gs.areaKind != 17 || e.world.Scripted == 0 || c.Ask == nil || paneText(c) != "THE MAN FROM THE HOLOGRAM GREETS YOU." {
		t.Fatalf("block %d scripted %d ask %+v text %q", e.gs.areaKind, e.world.Scripted, c.Ask, paneText(c))
	}
	read := false       // HIS SPEECH, entry 41 of the Log Book, in the pager
	for n := 0; ; n++ { // the greeting's pages, up to the base's menu
		sc, ok := e.build().(game.Scene)
		if ok && sc.Kind == game.SceneHandout {
			read = len(sc.Texts) == 1 && sc.Texts[0].Title == "Log Entry 41" && sc.Handout == "Log Book"
		} else if a := commonOf(e.build()).Ask; a != nil && len(a.Replies) == 5 {
			break
		}
		if n > 20 {
			t.Fatalf("no menu: %q %+v", paneText(commonOf(e.build())), commonOf(e.build()).Ask)
		}
		br1Answer(t, e)
	}
	if !read || !slices.ContainsFunc(e.ext.notes, func(n noteRec) bool { return n.Title == "Log Entry 41" }) {
		t.Fatalf("log entry 41: read %v, notes %+v", read, e.ext.notes)
	}
	atMenu := func() bool {
		c := commonOf(e.build())
		return c.Ask != nil && slices.Equal(c.Ask.Replies, []string{"Bar", "Clinic", "Depot", "Hq", "Port"}) && strings.HasPrefix(paneText(c), "WHERE DO YOU WISH TO GO?")
	}
	if !atMenu() {
		t.Fatalf("menu %+v text %q", commonOf(e.build()).Ask, paneText(commonOf(e.build())))
	}
	if _, ok := e.build().(game.Exploring); ok {
		t.Fatal("a scripted place waits for a move")
	}
	// HQ: the commander's orders, then the team menu.
	if _, err := e.doNow(game.Reply{N: 3}); err != nil {
		t.Fatal(err)
	}
	team := false // HQ ends in the team menu (Program 0), left as it is begun
	for n := 0; !atMenu(); n++ {
		if n > 20 {
			t.Fatalf("no menu after HQ: %q %+v", paneText(commonOf(e.build())), commonOf(e.build()).Ask)
		}
		if _, ok := e.build().(game.Assembling); ok {
			team = true
			if _, err := e.doNow(game.Begin{}); err != nil {
				t.Fatal(err)
			}
			continue
		}
		br1Answer(t, e)
	}
	if !team {
		t.Fatal("HQ did not open the team menu")
	}
}

// TestBR1TurnAndWall: turning, and a wall stops a step (buckrogers-go TestTurnAndWall; O09E:14E8 moves only
// through door state 1).
func TestBR1TurnAndWall(t *testing.T) {
	e := br1Begun(t)
	if _, err := e.doNow(game.Face{Facing: 6}); err != nil || e.facing != 6 {
		t.Fatalf("facing %d: %v", e.facing, err)
	}
	for y := range 16 {
		for x := range 16 {
			for d := 0; d < 8; d += 2 {
				if e.DoorAt(x, y, d) != 0 {
					continue
				}
				e.posX, e.posY, e.facing = x, y, d
				if err := stepBy(e, 0); err == nil && (e.posX != x || e.posY != y) {
					t.Fatalf("walked through the wall at %d,%d side %d", x, y, d)
				}
				return
			}
		}
	}
	t.Fatal("no wall on the map")
}

// TestScriptedPlace: in a place run by its script (view kind 3) a script's run that ends does not wait for a
// move: the move entry runs again and asks (BR1 pO0CA input's default; buckrogers-go input). Block 17 never
// ends a run, so a block of the test's own, written into the imported folder, does.
func TestScriptedPlace(t *testing.T) {
	e := br1Begun(t)
	const block = 200
	head, _, _ := strings.Cut(readFile(t, filepath.Join(e.scriptDir(), "area16.txt")), "\n") // the version line of this import's layout
	text := head + `
on load {
    AreaKind(3, 255)
    exit
}

on move {
    PartyX += 1
    Text("WHERE DO YOU WISH TO GO?")
    PressKey()
    exit
}

on step {
    exit
}

on camp {
    exit
}

on rest {
    exit
}
`
	if err := os.WriteFile(filepath.Join(e.scriptDir(), "area200.txt"), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	e.push(cont{kind: kStepEnd}) // as after a step whose script changed the block: the next command follows the new area's start
	e.DevAt(block, 0, 0)
	for want := 1; want <= 3; want++ { // each key ends a run; the next one asks again, one run each
		c := commonOf(e.build())
		if _, walking := e.build().(game.Exploring); walking || c.Ask == nil || paneText(c) != "WHERE DO YOU WISH TO GO?" || e.posX != want {
			t.Fatalf("run %d: %T ask %+v text %q, %d runs of the move entry", want, e.build(), c.Ask, paneText(c), e.posX)
		}
		br1Answer(t, e)
	}
}

func readFile(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// br1Pages answers the pages and pauses up until the game waits for a move.
func br1Pages(t *testing.T, e *Engine) {
	t.Helper()
	for n := 0; commonOf(e.build()).Ask != nil; n++ {
		if n > 80 {
			t.Fatalf("the pages do not end: %q", paneText(commonOf(e.build())))
		}
		br1Answer(t, e)
	}
}

// br1Answer answers what is up the one way it has: a handout put away, a Delay's pause over, a page's key.
func br1Answer(t *testing.T, e *Engine) {
	t.Helper()
	var a game.Action = game.Reply{}
	if sc, ok := e.build().(game.Scene); ok && sc.Kind == game.SceneHandout {
		a = game.Leave{}
	} else if slices.Contains(commonOf(e.build()).Allowed.Plain, game.Action(game.ContinueAction{})) {
		a = game.ContinueAction{}
	}
	if _, err := e.doNow(a); err != nil {
		t.Fatalf("%T on %q: %v (%T, allowed %+v)", a, paneText(commonOf(e.build())), err, e.build(), commonOf(e.build()).Allowed.Plain)
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
