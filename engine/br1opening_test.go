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
	"gobox/rules"
	"gobox/rulesxxvc"
)

// br1Engine is the shared engine on BR1's folder, imported from the originals
// into a temp dir (importer.ImportBREngine), with the XXVc ruleset, its save folder seeded with the shipped characters
// as cmd/gobox homeSaves does. It skips
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
	return br1EngineIn(t, adv)
}

// br1EngineIn is a new engine on an imported BR1 folder (the import is read, never written, by an engine).
func br1EngineIn(t testing.TB, adv string) *Engine {
	t.Helper()
	tables, err := adventure.ReadGameTables(adv)
	if err != nil {
		t.Fatal(err)
	}
	saves := t.TempDir()
	if err := os.CopyFS(saves, os.DirFS(filepath.Join(adv, importer.SavesDir))); err != nil { // a new save folder is seeded from the adventure's saves/ (cmd/gobox homeSaves)
		t.Fatal(err)
	}
	e, err := New(rulesxxvc.Builder, tables, adv, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e.Cfg.SaveDir = saves
	seed := uint32(7)
	e.FixedSeed = &seed
	return e
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
	return br1BeginOn(t, br1Engine(t))
}

// br1BeginOn takes a new engine through the title and the party menu to the first area.
func br1BeginOn(t testing.TB, e *Engine) *Engine { return br1BeginWith(t, e, 1) }

// br1BeginWith is br1BeginOn with the first n roster characters added through the party menu before Begin.
func br1BeginWith(t testing.TB, e *Engine, n int) *Engine {
	t.Helper()
	e.Open("")
	e.titleDone()
	e.Advance()
	for i := range n {
		if _, err := e.doNow(game.AddCharacter{Roster: i}); err != nil {
			t.Fatalf("AddCharacter: %v", err)
		}
	}
	if _, err := e.doNow(game.Begin{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	br1Pages(t, e)
	return e
}

// br1AtBaseMenu is whether the base's menu (block 17, Salvation) is up.
func br1AtBaseMenu(e *Engine) bool {
	c := commonOf(e.build())
	return c.Ask != nil && slices.Equal(c.Ask.Replies, []string{"Bar", "Clinic", "Depot", "Hq", "Port"}) && strings.HasPrefix(paneText(c), "WHERE DO YOU WISH TO GO?")
}

// br1ToBaseMenu enters block 17 and answers its greeting's pages up to the base's menu; it says whether
// HIS SPEECH was read in the pager (Log Book entry 41).
func br1ToBaseMenu(t *testing.T, e *Engine) (read bool) {
	t.Helper()
	e.DevAt(17, 0, 0)
	c := commonOf(e.build())
	if e.gs.areaKind != 17 || e.world.Scripted == 0 || c.Ask == nil || paneText(c) != "THE MAN FROM THE HOLOGRAM GREETS YOU." {
		t.Fatalf("block %d scripted %d ask %+v text %q", e.gs.areaKind, e.world.Scripted, c.Ask, paneText(c))
	}
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
	if !br1AtBaseMenu(e) {
		t.Fatalf("menu %+v text %q", commonOf(e.build()).Ask, paneText(commonOf(e.build())))
	}
	return read
}

// TestBR1NewBlock: a new block's start (buckrogers-go TestNewBlock): block 17, Salvation base, has no map and no
// walking; its script greets the party and then asks where to go, again after every answer that ends its run.
func TestBR1NewBlock(t *testing.T) {
	e := br1Begun(t)
	read := br1ToBaseMenu(t, e)
	if !read || !slices.ContainsFunc(e.ext.notes, func(n noteRec) bool { return n.Title == "Log Entry 41" }) {
		t.Fatalf("log entry 41: read %v, notes %+v", read, e.ext.notes)
	}
	if _, ok := e.build().(game.Exploring); ok {
		t.Fatal("a scripted place waits for a move")
	}
	// HQ: the commander's orders, then the team menu.
	if _, err := e.doNow(game.Reply{N: 3}); err != nil {
		t.Fatal(err)
	}
	team := false // HQ ends in the team menu (Program 0), left as it is begun
	for n := 0; !br1AtBaseMenu(e); n++ {
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
func br1Pages(t testing.TB, e *Engine) {
	t.Helper()
	for n := 0; commonOf(e.build()).Ask != nil; n++ {
		if n > 80 {
			t.Fatalf("the pages do not end: %q", paneText(commonOf(e.build())))
		}
		br1Answer(t, e)
	}
}

// br1Answer answers what is up the one way it has: a handout put away, a Delay's pause over, a page's key.
func br1Answer(t testing.TB, e *Engine) { t.Helper(); br1AnswerOut(t, e) }

// br1AnswerOut is br1Answer with what the engine did.
func br1AnswerOut(t testing.TB, e *Engine) game.Outcome {
	t.Helper()
	var a game.Action = game.Reply{}
	if sc, ok := e.build().(game.Scene); ok && sc.Kind == game.SceneHandout {
		a = game.Leave{}
	} else if slices.Contains(commonOf(e.build()).Allowed.Plain, game.Action(game.ContinueAction{})) {
		a = game.ContinueAction{}
	}
	out, err := e.doNow(a)
	if err != nil {
		t.Fatalf("%T on %q: %v (%T, allowed %+v)", a, paneText(commonOf(e.build())), err, e.build(), commonOf(e.build()).Allowed.Plain)
	}
	return out
}

// br1North is gobox's step north with the square's event run.
func br1North(e *Engine) error { return br1NorthVia(e, e.doNow) }

// br1NorthVia is br1North sending its actions through do.
func br1NorthVia(e *Engine, do func(game.Action) (game.Outcome, error)) error {
	if err := stepVia(e, do, 0); err != nil {
		return err
	}
	if e.ext.owed == nil {
		return nil
	}
	_, err := do(game.RunEvent{})
	return err
}

// paneText is the text pane's words, the wrapping's spaces aside.
func paneText(c game.Common) string {
	return strings.Join(strings.Fields(strings.Join(c.Text, " ")), " ")
}

// br1Clinic walks the base's CLINIC (reply 1) from the menu and returns the text up to the menu again.
func br1Clinic(t *testing.T, e *Engine) string {
	t.Helper()
	if _, err := e.doNow(game.Reply{N: 1}); err != nil {
		t.Fatal(err)
	}
	var text []string
	for n := 0; !br1AtBaseMenu(e); n++ {
		if n > 40 {
			t.Fatalf("no menu after the clinic: %q", paneText(commonOf(e.build())))
		}
		text = append(text, paneText(commonOf(e.build())))
		br1Answer(t, e)
	}
	return strings.Join(text, "\n")
}

// TestBR1Clinic: block 17's CLINIC reads and writes the selected member through the typed facts (status, hit
// points, maximum, name) and walks SelectMember past the party (status 0, as BR's noChar).
func TestBR1Clinic(t *testing.T) {
	member := func(e *Engine) rules.CharID { return e.partyIDs()[0] }
	fact := func(e *Engine, f rules.Fact) rules.Value {
		v, err := e.ruleset().Fact(member(e), f)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	hurt := func(e *Engine) {
		if err := e.ruleset().SetFact(member(e), rules.FactHP, rules.Number(1)); err != nil {
			t.Fatal(err)
		}
	}
	t.Run("heals", func(t *testing.T) {
		e := br1Begun(t)
		hurt(e)
		name := string(fact(e, rules.FactName).(rules.Text))
		max := fact(e, rules.FactMaxHP).(rules.Number)
		if max <= 1 {
			t.Fatalf("maximum %d", max)
		}
		br1ToBaseMenu(t, e)
		if got := br1Clinic(t, e); !strings.Contains(got, name+" IS HEALED...") {
			t.Fatalf("first visit: %q", got)
		}
		if hp := fact(e, rules.FactHP); hp != max || fact(e, rules.FactStatus) != rules.StatusOkay {
			t.Fatalf("hp %v of %d, status %v", hp, max, fact(e, rules.FactStatus))
		}
		if got := br1Clinic(t, e); strings.Contains(got, "HEALED") {
			t.Fatalf("second visit heals again: %q", got)
		}
		if hp := fact(e, rules.FactHP); hp != max {
			t.Fatalf("hp %v after the second visit", hp)
		}
	})
	t.Run("dead", func(t *testing.T) {
		e := br1Begun(t)
		name := string(fact(e, rules.FactName).(rules.Text))
		if err := e.ruleset().SetFact(member(e), rules.FactStatus, rules.StatusDead); err != nil {
			t.Fatal(err)
		}
		br1ToBaseMenu(t, e)
		if got := br1Clinic(t, e); !strings.Contains(got, name+" IS DEAD. SORRY.") {
			t.Fatalf("dead member: %q", got)
		}
		if fact(e, rules.FactStatus) != rules.StatusDead {
			t.Fatalf("status %v", fact(e, rules.FactStatus))
		}
	})
}

// TestBR1NoMember: the status word reads 0 while the last SELECT MEMBER matched nobody (BR's noChar), and
// the member's own status (1: okay) again after one that matched.
func TestBR1NoMember(t *testing.T) {
	e := br1Begun(t)
	r, ok := e.layoutField("SelStatus")
	if !ok {
		t.Fatal("no SelStatus in the layout")
	}
	status := func() int64 { return e.host.Get(r) }
	if got := status(); got != 1 {
		t.Fatalf("member 0: status %d", got)
	}
	e.es.memberMissing = true // SelectMember(1) with a party of one
	if got := status(); got != 0 {
		t.Fatalf("no member: status %d", got)
	}
	if got := status(); got != 0 {
		t.Fatalf("no member, read again: status %d", got)
	}
}

// TestBR1MemberFields: BR's member fields have the Pool's names: the hit points, the gender (0 male, 1 female)
// and the class number, which BR numbers by the career's place in the catalog from 1 (BrSelCharClass).
func TestBR1MemberFields(t *testing.T) {
	e := br1Begun(t)
	read := func(name string) int64 {
		r, ok := e.layoutField(name)
		if !ok {
			t.Fatalf("no %s in the layout", name)
		}
		return e.host.Get(r)
	}
	if got, want := read("SelHpCur"), int64(e.selValue(rules.FactHP).(rules.Number)); got != want {
		t.Errorf("SelHpCur %d, want %d", got, want)
	}
	if got, want := read("SelHpMax"), int64(e.selValue(rules.FactMaxHP).(rules.Number)); got != want {
		t.Errorf("SelHpMax %d, want %d", got, want)
	}
	if got, want := read("SelGender"), int64(slices.Index(adventure.SexOrder, e.selValue(rules.FactSex).(rules.Sex))); got != want {
		t.Errorf("SelGender %d, want %d", got, want)
	}
	career := e.selValue(rules.FactClasses).(rules.ClassSet)[0]
	if got, want := read("SelCharClass"), int64(slices.Index(e.ruleset().Catalog().Classes, career)+1); got != want || want < 1 {
		t.Errorf("SelCharClass %d, want %d", got, want)
	}
}
