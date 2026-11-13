package pool

import (
	"gobox/game/play"
	"rules1e"
	"strings"
	"testing"
)

// A party wipe shows the original's death page and waits for a key before
// the game ends (ovr06:1527's PromptRow24), never straight to the menu.
func TestWipeWaitsOnDeathPage(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	for _, c := range e.state().Party {
		c.Status, c.Okay = rules1e.StatusDead, 0
	}
	after := func() bool { return len(e.k) == 0 } // the cont is done: the script goes on
	e.push(cont{kind: kNop})
	e.pushAfterCombat() // as the script's combat op
	if r := e.snap.results; r == nil || !strings.Contains(r.End, "destroyed") {
		t.Fatalf("death page: %+v", r)
	}
	if e.question() == nil {
		t.Fatal("the death page does not wait for a key")
	}
	if after() {
		t.Fatal("the game went on under the death page")
	}
	press(e)
	if !after() {
		t.Fatal("the key did not end the death page")
	}
}

// A won fight's results page waits for its key (ovr06:0916) before what
// follows; a hired member's share page waits for its own key first
// (ovr06:12E8). BUG-CLASSES 11.
func TestResultPagesWait(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	if len(e.state().Party) < 2 {
		t.Skip("needs two members")
	}
	h := e.state().Party[1]
	h.Control, h.Share, h.Status, h.Okay = 0x80, 1, rules1e.StatusOkay, 1
	e.state().Loot.Gold = 100
	after := func() bool { return len(e.k) == 0 }
	e.push(cont{kind: kNop})
	e.pushAfterCombat()
	if r := e.snap.results; r == nil || len(r.Shares) == 0 || e.question() == nil || after() {
		t.Fatalf("share page: results %+v, question %v, after %v", r, e.question() != nil, after())
	}
	press(e)
	if r := e.snap.results; r == nil || r.Outcome == "" || e.question() == nil || after() {
		t.Fatalf("results page: results %+v, question %v, after %v", r, e.question() != nil, after())
	}
	sc, _ := e.build().(play.Scene)
	if sc.Results == nil {
		t.Fatal("results page pending but not published (BUG-CLASSES 14)")
	}
	if g := sc.Results; strings.Contains(g.Exp, "experience") || !strings.ContainsAny(g.Exp[len(g.Exp)-1:], "0123456789") { // BUG-CLASSES 24: gobox adds " experience points."
		t.Fatalf("Exp %q carries gobox's wording", g.Exp)
	}
	press(e)
	if !after() {
		t.Fatal("the key did not end the results page")
	}
}

// TestAskAllowsAnswersOnly (BUG-CLASSES 18): while a question is published
// (Ask), Allowed holds its answer and no item verb.
func TestAskAllowsAnswersOnly(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	e.keyWait(cont{kind: kNop})
	c := commonOf(e.build())
	if c.Ask == nil {
		t.Fatal("no question up")
	}
	for _, a := range c.Allowed.Plain {
		switch a.(type) {
		case play.Continue, play.Reply:
		default:
			t.Errorf("%T allowed while asking", a)
		}
	}
	if len(c.Allowed.Plain) == 0 || len(c.Allowed.Aimed) != 0 || len(c.Allowed.Moves) != 0 {
		t.Errorf("allowed while asking: %+v", c.Allowed)
	}
}
