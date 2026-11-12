package pool

import (
	"image"
	"slices"
	"strings"
	"testing"

	"gobox/game/play"
)

// The training hall (opProgram 0) puts the party menu up; Train shows the
// hall's offer, and Begin goes back out where the party stood.
func TestHallTrain(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	x, y, n := e.posX, e.posY, e.top()
	e.state().Game.PartyExp = 0xFF
	e.state().Party[0].Exp, e.state().Purse.Gold = 1<<20, 1<<20 // Train is offered only to a member who would rise
	e.push(cont{kind: kDrawMain})
	e.partyMenu(true, true)
	e.Advance()
	if _, ok := e.State().(play.Assembling); !ok || !e.ext.hall {
		t.Fatalf("hall's menu: %T hall %v", e.State(), e.ext.hall)
	}
	training := func() bool { tr, ok := e.State().(play.Trading); return ok && tr.Kind == play.TradeTraining }
	if err := e.OpenTraining(0); err != nil || !training() || commonOf(e.State()).Adventure {
		t.Fatalf("Train: %v, training %v", err, training())
	}
	if err := e.CloseTraining(); err != nil || training() {
		t.Fatalf("CloseTraining: %v", err)
	}
	e.Begin()
	e.Advance()
	_, exploring := e.State().(play.Exploring)
	adventure := commonOf(e.State()).Adventure
	if e.ext.hall {
		t.Error("still the hall's menu after Begin")
	}
	if e.top() != n || e.posX != x || e.posY != y || !adventure || !exploring {
		t.Errorf("after Begin: on %d (was %d), at %d,%d (was %d,%d), adventure %v", e.top(), n, e.posX, e.posY, x, y, adventure)
	}
}

// The engine draws nothing of the original's screens: what it publishes
// as pictures is art, never a fill or a frame.
func TestPicturesAreArt(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	for _, a := range []play.Action{nil, play.MakeCamp{}, play.BreakCamp{}, play.MakeCamp{}, play.SelectMember{Member: 0}, play.BreakCamp{}} {
		if a != nil {
			if _, err := e.Do(a); err != nil {
				t.Fatalf("%T: %v", a, err)
			}
		}
		for _, p := range e.pagePictures() {
			if p.Fill || p.Img == nil && p.Pix == nil {
				t.Fatalf("after %T: a fill or empty picture at %v", a, p.At)
			}
		}
	}
}

// The title's pictures are whole-screen pictures (Flags.Full), 320x200.
func TestTitleFull(t *testing.T) {
	e := testEngine(t)
	e.Init()
	e.ext.intro = true // as Run
	e.fullPicture(e.LoadPic("TITLE", 1))
	e.Advance()
	pics := e.pagePictures()
	if !e.snap.full || len(pics) == 0 {
		t.Fatalf("title: full %v, %d pictures", e.snap.full, len(pics))
	}
	var box image.Rectangle
	for _, p := range pics {
		box = box.Union(p.At)
	}
	if box.Dx() != 320 || box.Dy() != 200 {
		t.Errorf("title picture %v, want 320x200", box)
	}
}

// The sheet opened while exploring is a page of its own: the world's 3D
// view (Flags.Adventure) does not show over it.
func TestSheetOverWorld(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	adventure := func() bool { e.Advance(); return commonOf(e.State()).Adventure }
	if !adventure() {
		t.Fatal("not out in the world")
	}
	if e.OpenSheet(-1); !e.sheetUp() || adventure() {
		t.Fatalf("State: sheet %v, adventure %v", e.sheetUp(), adventure())
	}
	if e.CloseSheet(); !adventure() {
		t.Error("back from the sheet: not out in the world")
	}
}

// A shop or temple the script opened is its wait (PendShop): the
// rule calls act in it, Leave answers it and the script goes on.
func TestShopTemplePending(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	trading := func(k play.TradeKind) bool {
		e.Advance()
		tr, ok := e.State().(play.Trading)
		return ok && tr.Kind == k
	}
	exploring := func() bool { _, ok := e.State().(play.Exploring); return ok }
	left := func() bool { return len(e.k) == 0 } // the cont is done: the script goes on
	e.push(cont{kind: kNop})
	e.Shop()
	if e.pending != PendShop || !trading(play.TradeShop) || exploring() {
		t.Fatalf("shop: pending %v, %T", e.pending, e.State())
	}
	if e.Cure(0, 2) == nil {
		t.Error("a temple's cure in the shop")
	}
	if err := e.Browse(0); err != nil || !e.invUp() || !e.snap.shop {
		t.Fatalf("Browse: %v", err)
	}
	if e.Leave() == nil {
		t.Error("left the shop with the goods up")
	}
	if e.CloseItems(true); !trading(play.TradeShop) {
		t.Fatal("back from the goods: no shop")
	}
	if err := e.Leave(); err != nil || !left() {
		t.Fatalf("Leave: %v, script on %v", err, left())
	}
	if e.pending != PendNone || trading(play.TradeShop) || !exploring() {
		t.Fatalf("after the shop: pending %v, %T", e.pending, e.State())
	}
	e.push(cont{kind: kNop})
	e.Temple()
	e.state().Purse.Gold = 0
	if !trading(play.TradeTemple) || e.Cure(0, 2) == nil {
		t.Fatal("temple: a cure without money")
	}
	if err := e.Leave(); err != nil || !left() || trading(play.TradeTemple) {
		t.Fatalf("Leave the temple: %v, script on %v", err, left())
	}
}

// TestTempleWords: a game's temple data reaches the screen and the cure: the service the temple
// declines to cast anyway (game.Service.Declines) and what it says of a member it helped (Done).
func TestTempleWords(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	tab := *e.state().T()
	tab.Town.Temple = slices.Clone(tab.Town.Temple)
	tab.Town.Temple[2].Declines, tab.Town.Temple[2].Done = true, "is healed"
	e.state().Tables = &tab
	e.push(cont{kind: kNop})
	e.Temple()
	e.Advance()
	tr, ok := e.State().(play.Trading)
	if !ok || len(tr.Services) < 3 || !tr.Services[2].Declines || tr.Services[1].Declines {
		t.Fatalf("services %+v (%T)", tr.Services, e.State())
	}
	e.state().Purse.Gold = 10000
	if err := e.Cure(0, 2); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(e.ext.note, "is healed") {
		t.Errorf("note %q: want the service's own words", e.ext.note)
	}
}
