package main

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"gobox/adventure"
	"gobox/asset/orig"
	"gobox/game"
	"gobox/importer"
	"gobox/ui"

	"buckrogers/presentdata"
	"buckrogers/rules"
)

// newGame is the game as main hands it to gobox's front: the rules under
// classic with the game's data.
func br1NewGame(t *testing.T) *ui.Game {
	t.Helper()
	e, err := rules.New(br1Adv(), t.TempDir(), 1)
	if err != nil {
		t.Skip(err)
	}
	ui.SetScripts(br1Start(t), nil)
	return ui.NewGame(e, presentdata.Data(br1Adv()))
}

// br1Start is the imported adventure's Start script (its manifest's).
func br1Start(t *testing.T) *adventure.Sequence {
	t.Helper()
	f, err := importer.Read(br1Adv())
	if err != nil {
		t.Skip(err)
	}
	return f.Manifest.Start
}

// TestTitle plays the title as gobox's front does (front.IntroStep's
// calls): each picture step shows its TITLE pictures, decodable from the
// game's files; the credits are placed; TitleDone brings up the team menu.
func TestBR1Title(t *testing.T) {
	g := br1NewGame(t)
	intro := br1TitleIntro(g)
	if intro == nil || len(intro.Pictures) != 3 { // TITLE 1, TITLE 2 and 3, the credits page
		t.Fatalf("no title to play: %+v", intro)
	}
	l, err := orig.NewLoader(br1Dos)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range intro.Pictures[:2] {
		g.TitlePicture(n)
		ps := g.Page().Layers
		if len(ps) == 0 {
			t.Fatalf("step %d: no pictures", n)
		}
		for _, p := range ps {
			im := orig.ParseVGA(l.BlockShared(p.File, p.Block), false, 0)
			if im == nil || im.Width*8 != p.At.Dx() || im.Height != p.At.Dy() {
				t.Errorf("step %d: %s %d at %v, decodes to %+v", n, p.File, p.Block, p.At, im)
			}
		}
	}
	if len(br1Start(t).Credits) == 0 {
		t.Error("no credits")
	}
	g.TitleCredits(0)
	if !g.Flags().Credits {
		t.Fatal("the credits page is not up")
	}
	g.TitleDone()
	if br1TitleIntro(g) != nil || g.Mode() != game.ModeParty {
		t.Fatalf("after the title: intro %v, mode %v", br1TitleIntro(g), g.Mode())
	}
}

// titleIntro is the title the game waits to have played (nil: none).
func br1TitleIntro(g *ui.Game) *game.Intro {
	if s, ok := g.G.State().(game.Scene); ok && s.Kind == game.SceneTitle {
		return s.Intro
	}
	return nil
}

// TestCreate creates a character the way gobox's front does: Create on
// gobox's party menu (front.CreateAct: CreateCharacter, then the sheet
// while CreationView is up), the sheet's fields, Roll, the name, Keep; the
// original's questions after the name (skills, icon, save) follow.
func TestBR1Create(t *testing.T) {
	g := br1NewGame(t)
	g.TitleDone()
	// present.PartyMenu maps the label "Create Character" to Create.
	if p := g.Asked(); p == nil || !slices.ContainsFunc(p.Options, func(o game.Option) bool { return o.Label == "Create Character" }) {
		t.Fatalf("party menu %+v: no Create Character", p)
	}
	if err := g.CreateCharacter(); err != nil {
		t.Fatal(err)
	}
	cr := g.CreationView()
	if cr == nil {
		t.Fatal("no creation sheet")
	}
	// find is the sheet field offering text, and that option's index.
	find := func(text string) (f, o int) {
		for f, fd := range g.CreationView().Fields {
			for o, op := range fd.Options {
				if op.Text == text {
					return f, o
				}
			}
		}
		t.Fatalf("no field offers %q: %+v", text, g.CreationView().Fields)
		return
	}
	// The adnd2-br sheet: seven abilities ending in TECH, the figures with
	// an age, no alignment row, and each field's box over its word.
	dr := g.Draft()
	if dr == nil || len(dr.Stats) != 7 || dr.Stats[6].Label != "TECH" {
		t.Fatalf("sheet abilities %+v", dr)
	}
	if len(dr.Derived) == 0 || dr.Derived[0].Label != "Age" {
		t.Fatalf("sheet figures %+v", dr.Derived)
	}
	if age, _ := strconv.Atoi(dr.Derived[0].Value); age <= 0 {
		t.Errorf("age %q", dr.Derived[0].Value)
	}
	if len(cr.Fields) != 3 {
		t.Fatalf("fields %+v: want gender, race, career", cr.Fields)
	}
	for i, f := range cr.Fields {
		if strings.EqualFold(f.Name, "Alignment") {
			t.Errorf("an alignment row: %+v", f)
		}
		if word := f.Options[f.Cur].Text; f.At.X1 != 1 || f.At.Y1 != 3+i || f.At.X2 != len(word) {
			t.Errorf("field %d box %+v, its word %q at row %d", i, f.At, word, 3+i)
		}
	}
	race, tinker := find("Tinker")
	class, jock := find("Rocket Jock")
	if err := g.SetField(race, tinker); err != nil {
		t.Fatal(err)
	}
	cl := g.CreationView().Fields[class]
	if cl.Options[cl.Cur].Text != "Medic" || cl.Options[jock].Fits {
		t.Fatalf("a Tinker: %+v", cl)
	}
	if err := g.SetField(class, jock); err == nil {
		t.Error("a Tinker Rocket Jock taken")
	}
	if err := g.Roll(false, ""); err != nil {
		t.Fatal(err)
	}
	if err := g.Rename("Wilma"); err != nil {
		t.Fatal(err)
	}
	if err := g.EndCreation(true); err != nil {
		t.Fatal(err)
	}
	c := g.Draft()
	if c == nil || c.Draft.Name != "Wilma" || c.Draft.Race != "Tinker" || g.Asked() == nil {
		t.Fatalf("after Keep: %+v", c)
	}
}

// TestFont: gobox's front draws with the game's own font, 8X8D1 block 201,
// its 'A' the original's.
func TestBR1Font(t *testing.T) {
	e, err := rules.New(br1Adv(), t.TempDir(), 1)
	if err != nil {
		t.Skip(err)
	}
	d := presentdata.Data(br1Adv())
	f := ui.NewFront(ui.NewGame(e, d), d)
	l, err := orig.NewLoader(br1Dos)
	if err != nil {
		t.Skip(err)
	}
	want := [8]byte(l.BlockShared("8X8D1", 201)[8:16]) // 'A', glyph 1
	if f.Glyph == nil || f.Glyph('A') != want {
		t.Fatalf("the front's 'A' is not the original's %x", want)
	}
}

// TestMenuTitle: TITLE 4 is gobox's backdrop under the team menu
// (the Start's Backdrop, O030:05BE), not an intro step.
func TestBR1MenuTitle(t *testing.T) {
	g := br1NewGame(t)
	if b := br1Start(t).Backdrop; len(b) != 1 || b[0].File != "TITLE" || b[0].Block != 4 || slices.Contains(br1TitleIntro(g).Pictures, 4) {
		t.Fatalf("backdrop %v, intro %v", b, br1TitleIntro(g).Pictures)
	}
}

// TestPartyMenuKeys drives gobox's main menu as the player does, by keys
// on present's Front: with no team it offers no Begin; Party, Create opens
// the creation sheet; Party, Add, a pick adds a character; then the menu
// offers Begin and B leaves the team menu.
func TestBR1PartyMenuKeys(t *testing.T) {
	var (
		e *rules.Engine
		g *ui.Game
		f *ui.Front
	)
	start := func() { // a fresh game on gobox's main menu
		var err error
		if e, err = rules.New(br1Adv(), t.TempDir(), 1); err != nil {
			t.Skip(err)
		}
		d := presentdata.Data(br1Adv())
		g = ui.NewGame(e, d)
		f = ui.NewFront(g, d)
		f.Frame(640, 400)
		g.TitleDone()
		f.Frame(640, 400)
	}
	// offered is the team menu's entries the game hands gobox's front.
	offered := func() map[string]bool {
		m := map[string]bool{}
		if p := g.Asked(); p != nil {
			for _, o := range p.Options {
				m[o.Label] = true
			}
		}
		return m
	}
	keys := func(ks ...uint16) {
		for _, k := range ks {
			f.Key(k)
			f.Frame(640, 400)
		}
	}
	start()
	if open := offered(); open["Begin Adventuring"] || !open["Create Character"] || !open["Add Character"] {
		t.Fatalf("no team: menu %v", open)
	}
	keys('P', 'C')
	if g.CreationView() == nil {
		t.Fatalf("Party, Create: no creation sheet (%s)", f.Screen())
	}
	start()
	keys('P', 'A', '\r', 27, 27, 27)
	if n := len(e.Facts().Party); n != 1 {
		t.Fatalf("Party, Add: team of %d (%s)", n, f.Screen())
	}
	if open := offered(); !open["Begin Adventuring"] {
		t.Fatalf("a team: menu %v", open)
	}
	f.Key('B')
	f.Act() // no Draw: the walls are not drawn headless
	if m := e.Facts().Mode; m == game.ModeParty {
		t.Fatalf("Begin: still on the team menu (%s)", f.Screen())
	}
}

// freshFront is gobox's front, as main builds it, on a copy of the game's
// files with no saved characters (a fresh roster), past the title and its
// fade: the main menu takes keys.
func br1FreshFront(t *testing.T) (*rules.Engine, *ui.Game, *ui.Front, func(...uint16)) {
	t.Helper()
	e, err := rules.New(br1BareAdv(t), t.TempDir(), 1) // no saved characters: a fresh roster
	if err != nil {
		t.Fatal(err)
	}
	d := presentdata.Data(br1Adv())
	g := ui.NewGame(e, d)
	f := ui.NewFront(g, d)
	keys := func(ks ...uint16) {
		for _, k := range ks {
			f.Key(k)
			f.Frame(640, 400)
		}
	}
	f.Frame(640, 400)
	for range 8 { // Esc through the title, a key past the menu's fade
		if f.P.Shown().MainMenu {
			break
		}
		keys(27)
	}
	keys('x')
	if !f.P.Shown().MainMenu {
		t.Fatalf("no main menu after the title (%s)", f.Screen())
	}
	return e, g, f, keys
}

// prompt is the game's question now ("": none).
func br1Prompt(g *ui.Game) string {
	if p := g.Asked(); p != nil {
		return p.Prompt
	}
	return ""
}

// in3D begins from the team menu (B), pages past the start's texts by
// Enter and draws frames in the 3D view.
func br1In3D(t *testing.T, g *ui.Game, f *ui.Front, keys func(...uint16)) {
	t.Helper()
	keys('B')
	for i := 0; i < 40 && !g.Flags().Exploring; i++ {
		keys('\r')
	}
	if !g.Flags().Exploring || g.Walls() == nil {
		t.Fatalf("Begin: not in the 3D view (%q, %s)", br1Prompt(g), f.Screen())
	}
	for range 5 {
		f.Frame(640, 400)
	}
}

// TestCreateByKeys plays a fresh roster to the 3D view by keys only: Party,
// Create, the sheet (career Warrior, the name, Roll, Keep), the skill
// screen, the Warrior's weapon, the icon, Save; Party, Add; Begin.
func TestBR1CreateByKeys(t *testing.T) {
	e, g, f, keys := br1FreshFront(t)
	const tab, down = 9, 0x5000
	keys('P', 'C')
	if g.CreationView() == nil {
		t.Fatalf("Party, Create: no sheet (%s)", f.Screen())
	}
	keys('R')  // Roll (not Point Buy)
	keys('\r') // the name (the first part)
	for _, c := range "Wilma" {
		keys(uint16(c))
	}
	keys('\r', tab, tab, tab, '\r', down, down, '\r', 'R', 'K') // Class: Warrior, Roll, Keep
	if c := g.Draft(); c == nil || c.Draft.Name != "Wilma" || c.Draft.Class != "Warrior" {
		t.Fatalf("after Keep: %+v", c)
	}
	steps := []struct {
		ask  string
		keys []uint16
	}{
		{"Select a weapon", []uint16{'\r', 'E'}},                // the pick, then Exit (0AA5's list stays)
		{"Career skill points", []uint16{'\r', '\r', 'D', 'Y'}}, // two points, Done, leave the rest
		{"General skill points", []uint16{'D', 'Y'}},
		{"Select a body icon", []uint16{'\r'}},
		{"Is this icon ok", []uint16{'\r'}},
		{"Save Wilma", []uint16{'\r'}},
	}
	for _, s := range steps {
		if !strings.HasPrefix(br1Prompt(g), s.ask) {
			t.Fatalf("want %q, the game asks %q (%s)", s.ask, br1Prompt(g), f.Screen())
		}
		keys(s.keys...)
	}
	if g.Mode() != game.ModeParty || g.CreationView() != nil {
		t.Fatalf("after Save: mode %v, %q", g.Mode(), br1Prompt(g))
	}
	keys('A', '\r', 27, 27) // the Party submenu is still up
	if n := len(e.Facts().Party); n != 1 {
		t.Fatalf("Party, Add: team of %d (%q, %s)", n, br1Prompt(g), f.Screen())
	}
	br1In3D(t, g, f, keys)
}

// TestQuickStart: the empty team menu offers Quick Start; Q makes the team
// of six (every career, each saved to the roster) and Begin goes into 3D.
func TestBR1QuickStart(t *testing.T) {
	e, g, f, keys := br1FreshFront(t)
	if p := g.Asked(); p == nil || !slices.ContainsFunc(p.Options, func(o game.Option) bool { return o.Label == "Quick Start" }) {
		t.Fatalf("no Quick Start on the empty team menu: %+v", p)
	}
	keys('Q')
	ps := e.Facts().Party
	if len(ps) != 6 {
		t.Fatalf("Quick Start: team of %d (%q)", len(ps), br1Prompt(g))
	}
	got := map[string]bool{}
	for _, m := range ps {
		got[m.Class] = true
	}
	if len(got) != 5 {
		t.Errorf("careers %v", got)
	}
	for _, o := range g.Asked().Options {
		if o.Label == "Quick Start" {
			t.Error("Quick Start offered with a team")
		}
	}
	br1In3D(t, g, f, keys)
}

// TestItemsTabSheet: Tab from the items (CloseItems, not All) shows the
// selected member's sheet on the play path.
func TestBR1ItemsTabSheet(t *testing.T) {
	e, err := rules.New(br1Adv(), t.TempDir(), 1)
	if err != nil {
		t.Skip(err)
	}
	r := ui.NewGame(e, presentdata.Data(br1Adv()))
	r.Send(game.Reply{})
	r.Send(game.QuickStart{})
	r.Send(game.Begin{})
	for i := 0; i < 15; i++ {
		r.Send(game.Reply{})
	}
	r.Send(game.OpenItems{})
	r.CloseItems(false)
	if r.SheetUp() == nil {
		t.Fatal("Tab closed the items without the sheet")
	}
}

// TestDraftFigures: the play path's creation sheet shows AC and THAC0.
func TestBR1DraftFigures(t *testing.T) {
	e, err := rules.New(br1Adv(), t.TempDir(), 1)
	if err != nil {
		t.Skip(err)
	}
	r := ui.NewGame(e, presentdata.Data(br1Adv()))
	r.Send(game.Reply{})
	r.Send(game.NewCharacter{})
	r.SetField(1, 0)
	r.SetField(2, 0)
	if c := r.Draft(); c == nil || c.Draft.AC == 0 || c.Draft.THAC0 == 0 {
		t.Fatalf("draft figures %+v", c)
	}
}

// TestItemsPaneActs: from gobox's item pane (no 0F75 under it) Ready
// toggles, the AC follows the armour, and Drop asks then drops.
func TestBR1ItemsPaneActs(t *testing.T) {
	e, err := rules.New(br1Adv(), t.TempDir(), 1)
	if err != nil {
		t.Skip(err)
	}
	r := ui.NewGame(e, presentdata.Data(br1Adv()))
	r.Send(game.Reply{})
	r.Send(game.QuickStart{})
	r.Send(game.Begin{})
	for i := 0; i < 15; i++ {
		r.Send(game.Reply{})
	}
	r.Send(game.OpenItems{})
	n := r.SelectedMember()
	m := func() game.Member { return r.Party()[n] }
	ac := m().AC
	// class 76: an unreadied item goes to the party pack (the list's end), so find items by type
	at := func(ty int) int {
		for i, it := range m().Items {
			if it.Type == ty {
				return i
			}
		}
		t.Fatalf("no item of type %d", ty)
		return -1
	}
	suit, gun := m().Items[0].Type, m().Items[1].Type
	if res := r.Send(game.Ready{Member: n, Item: game.PositionHandle(at(suit))}); res.Err != nil || m().Items[at(suit)].Readied || m().AC <= ac {
		t.Fatalf("unready the suit: %v AC %d from %d", res.Err, m().AC, ac)
	}
	if r.Send(game.Ready{Member: n, Item: game.PositionHandle(at(suit))}); !m().Items[at(suit)].Readied || m().AC != ac {
		t.Fatalf("ready the suit again: AC %d want %d", m().AC, ac)
	}
	if n := r.Common().PartyCap; n != 60 {
		t.Fatalf("class 82: PartyCap %d", n)
	}
	used := r.Common().PoolItems
	r.Send(game.Ready{Member: n, Item: game.PositionHandle(at(gun))}) // the gun off
	if r.Common().PoolItems != used+1 {
		t.Fatalf("class 82: unready: PoolItems %d want %d", r.Common().PoolItems, used+1)
	}
	r.Send(game.Ready{Member: n, Item: game.PositionHandle(at(gun))})
	if r.Common().PoolItems != used {
		t.Fatalf("class 82: ready: PoolItems %d want %d", r.Common().PoolItems, used)
	}
	r.Send(game.Ready{Member: n, Item: game.PositionHandle(at(gun))})
	if r.Send(game.Drop{Member: n, Item: game.PositionHandle(at(gun))}); r.Asked() == nil {
		t.Fatal("Drop did not ask")
	}
	if r.Send(game.Reply{N: 1}); len(m().Items) != 1 || r.Asked() != nil {
		t.Fatalf("Drop yes: %d items", len(m().Items))
	}
}

// TestLootListTakes (class 83): on the loot screen's item list, the list's
// All goes to the rules as TakeLoot (no gate menu swallows it as OK).
func TestBR1LootListTakes(t *testing.T) {
	e, _, f, keys := br1FreshFront(t)
	keys('Q')
	e.Debug = "fight"
	keys('B')
	for i := 0; i < 200 && e.Facts().Combat == nil; i++ {
		keys('\r')
	}
	for i := 0; i < 200 && e.Facts().Mode != game.ModeLoot; i++ {
		if e.Facts().Combat != nil && e.Pending() == nil {
			e.Do(game.Quick{All: true})
		} else {
			if _, err := e.Do(game.ContinueAction{}); err != nil {
				e.Do(game.Reply{})
			}
		}
		f.Frame(640, 400)
	}
	for i := 0; i < 40 && e.Facts().Mode == game.ModeLoot && !f.P.Shown().Inventory; i++ {
		keys('\r')
	}
	if e.Facts().Mode != game.ModeLoot {
		t.Fatalf("no loot screen (mode %v, %s)", e.Facts().Mode, f.Screen())
	}
	if sh := f.P.Shown(); !sh.Inventory || sh.Treasure == nil {
		t.Fatalf("no loot item list (%s)", f.Screen())
	}
	keys('A')
	if l := e.Facts().Loot; l != nil && len(l.Items) > 0 {
		t.Fatalf("All on the loot list left %v (%s)", l.Items, f.Screen())
	}
}
