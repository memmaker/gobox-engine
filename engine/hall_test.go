package engine

import (
	"slices"
	"strings"
	"testing"

	"gobox/game"
	"gobox/rules1e"
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
	e.world.PartyExp = devHallTeaches
	e.state().Party[0].Exp, e.state().Purse.Gold = 1<<20, 1<<20 // Train is offered only to a member who would rise
	e.push(cont{kind: kDrawMain})
	e.partyMenu(true, true)
	e.Advance()
	if _, ok := e.build().(game.Assembling); !ok || !e.ext.hall {
		t.Fatalf("hall's menu: %T hall %v", e.build(), e.ext.hall)
	}
	training := func() bool { tr, ok := e.build().(game.Trading); return ok && tr.Kind == game.TradeTraining }
	if err := e.openTraining(0); err != nil || !training() || commonOf(e.build()).Adventure {
		t.Fatalf("Train: %v, training %v", err, training())
	}
	if err := e.closeTraining(); err != nil || training() {
		t.Fatalf("CloseTraining: %v", err)
	}
	e.begin()
	e.Advance()
	_, exploring := e.build().(game.Exploring)
	adventure := commonOf(e.build()).Adventure
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
	for _, a := range []game.Action{nil, game.MakeCamp{}, game.BreakCamp{}, game.MakeCamp{}, game.SelectMember{Member: 0}, game.BreakCamp{}} {
		if a != nil {
			if _, err := e.doNow(a); err != nil {
				t.Fatalf("%T: %v", a, err)
			}
		}
		c := commonOf(e.build())
		pics := c.Frames
		if c.Picture != nil {
			pics = append(pics, *c.Picture)
		}
		if pt := c.Portrait; pt != nil {
			pics = append(pics, pt.Head, pt.Body)
		}
		for _, p := range pics {
			if p.Pixels == nil {
				t.Fatalf("after %T: an empty picture %+v", a, p)
			}
		}
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
	adventure := func() bool { e.Advance(); return commonOf(e.build()).Adventure }
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
	trading := func(k game.TradeKind) bool {
		e.Advance()
		tr, ok := e.build().(game.Trading)
		return ok && tr.Kind == k
	}
	exploring := func() bool { _, ok := e.build().(game.Exploring); return ok }
	left := func() bool { return len(e.k) == 0 } // the cont is done: the script goes on
	e.push(cont{kind: kNop})
	e.Shop()
	if e.pending != PendShop || !trading(game.TradeShop) || exploring() {
		t.Fatalf("shop: pending %v, %T", e.pending, e.build())
	}
	if e.cure(0, 2) == nil {
		t.Error("a temple's cure in the shop")
	}
	if err := e.browse(0); err != nil || !e.invUp() || !e.snap.shop {
		t.Fatalf("Browse: %v", err)
	}
	if e.Leave() == nil {
		t.Error("left the shop with the goods up")
	}
	if e.CloseItems(true); !trading(game.TradeShop) {
		t.Fatal("back from the goods: no shop")
	}
	if err := e.Leave(); err != nil || !left() {
		t.Fatalf("Leave: %v, script on %v", err, left())
	}
	if e.pending != PendNone || trading(game.TradeShop) || !exploring() {
		t.Fatalf("after the shop: pending %v, %T", e.pending, e.build())
	}
	e.push(cont{kind: kNop})
	e.Temple()
	e.state().Purse.Gold = 0
	if !trading(game.TradeTemple) || e.cure(0, 2) == nil {
		t.Fatal("temple: a cure without money")
	}
	if err := e.Leave(); err != nil || !left() || trading(game.TradeTemple) {
		t.Fatalf("Leave the temple: %v, script on %v", err, left())
	}
}

// TestTempleWords: a game's temple data reaches the cure: what it says of a member it helped (Done).
func TestTempleWords(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	tab := *e.state().T()
	tab.Town.Temple = slices.Clone(tab.Town.Temple)
	tab.Town.Temple[2].Done = "is healed"
	e.state().Tables = &tab
	e.push(cont{kind: kNop})
	e.Temple()
	e.Advance()
	tr, ok := e.build().(game.Trading)
	if !ok || len(tr.Services) < 3 {
		t.Fatalf("services %+v (%T)", tr.Services, e.build())
	}
	e.state().Purse.Gold = 10000
	e.state().Party[0].HP = 1 // a wound cure is given only to the hurt
	if err := e.cure(0, 2); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(e.ext.note, "is healed") {
		t.Errorf("note %q: want the service's own words", e.ext.note)
	}
}

// invUp: the inventory screen is up (not under a modal).
func (e *Engine) invUp() bool { return e.snap.invOn }

// Change Class is the training hall's only, offered for the selected
// member when it qualifies (classOffers), absent on the title's party menu.
func TestHallChangeClass(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	if len(e.state().Party) < 2 {
		t.Fatalf("party of %d, want two", len(e.state().Party))
	}
	dual, plain := e.state().Party[0], e.state().Party[1]
	dual.Race, dual.Class, dual.Alignment = rules1e.RaceHuman, rules1e.ClsFighter, 0
	dual.Stats = [6]rules1e.Score{17, 17, 17, 17, 17, 17}
	plain.Race = rules1e.RaceDwarf
	if len(rules1e.NewClasses(dual)) == 0 || len(rules1e.NewClasses(plain)) != 0 {
		t.Fatalf("fixture: dual %v, plain %v", rules1e.NewClasses(dual), rules1e.NewClasses(plain))
	}
	offered := func() bool {
		e.Advance()
		as, ok := e.build().(game.Assembling)
		if !ok {
			t.Fatalf("party menu: %T", e.build())
		}
		in := slices.ContainsFunc(as.Allowed.Plain, func(a game.Action) bool { _, ok := a.(game.ChangeClass); return ok })
		q := e.pendingAsk()
		if q != nil && slices.Contains(q.Replies, game.ChangeClassFunction.Label) != in {
			t.Errorf("menu %v disagrees with Allowed %v", q.Replies, in)
		}
		return in
	}
	e.push(cont{kind: kDrawMain})
	e.partyMenu(false, true)
	e.selectMember(0)
	if offered() {
		t.Error("Change Class on the title's party menu")
	}
	if _, err := e.doNow(game.ChangeClass{Member: 0}); err == nil {
		t.Error("Change Class done on the title's party menu")
	}
	e.push(cont{kind: kDrawMain})
	e.partyMenu(true, true)
	e.selectMember(0)
	if !offered() {
		t.Error("hall: a qualifying member is not offered Change Class")
	}
	e.selectMember(1)
	if offered() {
		t.Error("hall: a non-qualifying member is offered Change Class")
	}
	e.selectMember(0)
	offered()
	if _, err := e.doNow(game.ChangeClass{Member: 0}); err != nil {
		t.Fatalf("Change Class: %v", err)
	}
	if pq := e.ext.pq; pq == nil || len(pq.calls) == 0 || pq.calls[0].kind != callClass {
		t.Fatalf("Change Class asks %+v, want the class", pq)
	}
	// the class picked: what it costs is asked, No (where it starts) changes nothing, Yes changes the class
	for _, yes := range []int{0, 1} {
		if _, err := e.doNow(game.Reply{N: 0}); err != nil {
			t.Fatal(err)
		}
		pq := e.ext.pq
		if pq == nil || !pq.q.YesNo || pq.q.Default != 0 || !strings.Contains(pq.q.Text, "This is final.") {
			t.Fatalf("the class picked asks %+v, want the yes/no with what it costs, on No", pq)
		}
		if _, err := e.doNow(game.Reply{N: yes}); err != nil {
			t.Fatal(err)
		}
		if changed := rules1e.DualClassed(dual); changed != (yes == 1) || e.ext.pq != nil {
			t.Fatalf("reply %d: changed %v, question %+v", yes, changed, e.ext.pq)
		}
		if yes == 0 {
			if _, err := e.doNow(game.ChangeClass{Member: 0}); err != nil {
				t.Fatal(err)
			}
		}
	}
	if got := e.ruleset().Sheet(e.charID(dual)).DualClass; !strings.Contains(got, "dormant until") {
		t.Errorf("the sheet's dual class line %q", got)
	}
}
