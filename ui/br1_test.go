package ui

import (
	"encoding/json"
	"gobox/adventure"
	"gobox/game"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"gobox/buck"
	"gobox/engine"
	"gobox/importer"
)

// br1Originals are BR1's DOS files, beside the repo's data (skipped without).
const br1Originals = "../data/br1/sources/dos"

// br1Folder imports BR1 for the shared engine (gobox import -shared) into a
// temp dir, once per test.
func br1Folder(t *testing.T) importer.Folder {
	t.Helper()
	if _, err := os.Stat(filepath.Join(br1Originals, "START.EXE")); err != nil {
		t.Skip("no BR1 originals:", err)
	}
	root := t.TempDir()
	if err := importer.RunWith("br1", br1Originals, root, true); err != nil {
		t.Fatal(err)
	}
	f, err := importer.Read(filepath.Join(root, "br1"))
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// br1Front is the presenter on BR1's shared engine as cmd/gobox's playPool
// builds it (title, saves, no window).
func br1Front(t *testing.T, f importer.Folder) (*Front, *engine.Engine) {
	t.Helper()
	eng, err := buck.New(f.Dir, filepath.Join(f.Dir, "sources"))
	if err != nil {
		t.Fatal(err)
	}
	if err := eng.SetStart(f.Manifest.Start); err != nil {
		t.Fatal(err)
	}
	if err := eng.SetEnd(f.Manifest.End); err != nil {
		t.Fatal(err)
	}
	eng.Cfg.SaveDir = t.TempDir()
	seed := uint32(7)
	eng.FixedSeed = &seed
	data := eng.Data()
	SetScripts(f.Manifest.Start, f.Manifest.End)
	g := NewFront(NewGame(eng, data), data)
	g.Start("")
	return g, eng
}

// br1Frames draws n frames.
func br1Frames(f *Front, n int) {
	for range n {
		f.Frame(flowW, flowH)
		f.Tick()
	}
}

// br1Roster seeds the engine's roster with the originals' shipped characters
// (the import's data/characters), as .cha files and charlist.txt are.
// ponytail: the test seeds the roster; who seeds a new save folder in play is the team menu's slice (REFACTOR.md B1f).
func br1Roster(t *testing.T, f importer.Folder, eng *engine.Engine) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(f.Dir, adventure.DataDir, importer.CharactersDir, "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatal("no imported characters", err)
	}
	var list string
	for _, p := range files[:4] {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		var c struct{ Name string }
		if err := json.Unmarshal(b, &c); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(eng.Cfg.SaveDir, adventure.MakeFileName(c.Name)+".CHA"), b, 0o644); err != nil {
			t.Fatal(err)
		}
		list += c.Name + "\r\n"
	}
	if err := os.WriteFile(filepath.Join(eng.Cfg.SaveDir, "CHARLIST.TXT"), []byte(list), 0o644); err != nil {
		t.Fatal(err)
	}
}

// br1Begin takes the party menu from the title to Begin Adventuring with
// one member added, as the player does (Party, Add, Enter, Esc, Exit, Begin).
func br1Begin(t *testing.T, g *Front, eng *engine.Engine) {
	t.Helper()
	br1Check(t, g, eng, "title", game.Scene{})
	for range 3 {
		g.Key(keyEnter)
		br1Frames(g, 5)
	}
	br1Check(t, g, eng, "party menu", game.Assembling{})
	flowKey(g, 'p')
	flowKey(g, 'a')
	flowKey(g, keyEnter)
	if as, _ := eng.State().(game.Assembling); len(as.Party) != 1 {
		t.Fatalf("member added: party %d", len(as.Party))
	}
	flowKey(g, keyEsc) // the picker
	flowKey(g, 'e')
	flowKey(g, 'b')
}

// br1Check draws frames and checks the state is of kind want's type, the
// presenter draws it (no panic) and a question the game asks is on a screen:
// the frame is not stuck, and with nothing of the presenter's own over the
// game's screen it is not unseen either.
func br1Check(t *testing.T, g *Front, eng *engine.Engine, what string, want game.State) {
	t.Helper()
	br1Frames(g, 6)
	st := eng.State()
	if want != nil && reflect.TypeOf(st) != reflect.TypeOf(want) { // nil: whatever the engine published
		t.Fatalf("%s: state %T, want %T", what, st, want)
	}
	if got := g.P.state(); reflect.TypeOf(got) != reflect.TypeOf(st) {
		t.Errorf("%s: the presenter holds %T, the engine published %T", what, got, st)
	}
	if g.P.stuck() {
		t.Errorf("%s: asks %s on no screen", what, g.P.asksID())
	}
	if _, menu := st.(game.Assembling); !menu && g.P.unseen() { // the party menu's submenus are the presenter's own, over the game's question
		t.Errorf("%s: %T asks %s, not drawn", what, st, g.P.asksID())
	}
}

// br1Enter presses Enter until the state is of want's type (at most max keys).
func br1Enter(t *testing.T, g *Front, eng *engine.Engine, what string, want game.State, max int) {
	t.Helper()
	for range max {
		if reflect.TypeOf(eng.State()) == reflect.TypeOf(want) {
			br1Check(t, g, eng, what, want)
			return
		}
		g.Key(keyEnter)
		br1Frames(g, 6)
		if g.P.stuck() {
			t.Fatalf("%s: stuck on %T asking %s", what, eng.State(), g.P.asksID())
		}
	}
	t.Fatalf("%s: still %T after %d keys", what, eng.State(), max)
}

// TestBR1Explore: the title, the party menu, a member added, Begin, the opening's pages to the tarmac and a step,
// each state drawn by the presenter with its question seen, and the 3D view's walls published (the engine reads
// GameTables.Walls, the import writes importer/walls/br1.txt).
func TestBR1Explore(t *testing.T) {
	f := br1Folder(t)
	g, eng := br1Front(t, f)
	br1Roster(t, f, eng)
	br1Begin(t, g, eng)
	br1Check(t, g, eng, "opening", game.Scene{})
	br1Enter(t, g, eng, "tarmac", game.Exploring{}, 60)
	w := eng.State().(game.Exploring).Map.Walls
	if len(w.Side) == 0 || len(w.Pool) == 0 {
		t.Errorf("the tarmac's 3D view has %d rows of walls, textures %v", len(w.Side), w.Pool)
	}
	g.Key(keyUp)
	br1Check(t, g, eng, "a step", nil) // the step's script may open a page (a Scene) or leave the party exploring
}

// TestBR1BaseMenu: block 17's menu (Salvation base: WHERE DO YOU WISH TO GO?) is a Scene whose question the
// presenter draws. The dev start is the engine's -at (Capture.Dev.At).
func TestBR1BaseMenu(t *testing.T) {
	f := br1Folder(t)
	g, eng := br1Front(t, f)
	br1Roster(t, f, eng)
	eng.DevAction = func() { eng.DevAt(17, 0, 0) }
	br1Begin(t, g, eng)
	const baseReplies = 5 // Bar, Clinic, Depot, Hq, Port
	for range 60 {
		if sc, ok := eng.State().(game.Scene); ok && sc.Ask != nil && len(sc.Ask.Replies) == baseReplies {
			br1Check(t, g, eng, "base menu", game.Scene{})
			return
		}
		g.Key(keyEnter)
		br1Frames(g, 6)
		if g.P.stuck() {
			t.Fatalf("greeting: stuck on %T asking %s", eng.State(), g.P.asksID())
		}
	}
	t.Fatalf("no base menu: %T", eng.State())
}

// TestBR1Fight: a fight (the dev scene -fight 1) and its first turns, each frame drawn and each question seen.
func TestBR1Fight(t *testing.T) {
	f := br1Folder(t)
	g, eng := br1Front(t, f)
	br1Roster(t, f, eng)
	eng.DevMonster = 1
	eng.DevAction = eng.DevScene("combat")
	br1Begin(t, g, eng)
	br1Enter(t, g, eng, "fight", game.Fighting{}, 80)
	br1FightPictures(t, eng)
	for range 10 {
		g.Key(keyEnter)
		br1Check(t, g, eng, "fight turn", nil)
	}
}

// br1FightPictures: the fight publishes a picture for every unit and every square's tile (typed pictures: pixels in
// the art's own palette, the mask colour Clear), and a shot flown with a unit's picture as its sprite is a Report
// the presenter draws (eventsAt) with that sprite. The shot of a real weapon is engine/br1aim_test.go's.
func br1FightPictures(t *testing.T, eng *engine.Engine) {
	t.Helper()
	f := eng.State().(game.Fighting)
	art := eng.Art()
	for _, u := range f.Field.Units {
		if u.Foe { // the dev scene's foe loads CPIC block 1, which BR's CPIC1 does not hold: a script's foes are engine/br1pictures_test.go's
			continue
		}
		if u.Art.Pixels == nil || len(u.Art.Palette) == 0 || u.Art.Clear < 0 {
			t.Errorf("unit %d %s has no picture (%d pixels, %d colours, clear %d)", u.ID, u.Name, len(u.Art.Pixels), len(u.Art.Palette), u.Art.Clear)
		}
	}
	for _, s := range f.Field.Squares {
		if name := f.Field.Tiles[s.Tile]; name == "" || art[name].Pixels == nil {
			t.Errorf("square %v: tile %d (%q) has no picture", s.At, s.Tile, name)
		}
	}
	shot := game.Shot{From: f.Field.Units[0].ID, To: f.Field.Units[len(f.Field.Units)-1].At, Art: f.Field.Units[0].Art, StepMs: 10}
	if es := eventsAt([]game.Event{shot}, 0, nil); len(es) != 1 || es[0].Kind != game.EvMissile || len(es[0].Sprite) != 1 || es[0].Sprite[0].Pixels == nil {
		t.Errorf("the shot is %+v, want one missile with its sprite", es)
	}
}
