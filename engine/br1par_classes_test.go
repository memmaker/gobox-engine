package engine

import (
	"slices"
	"testing"

	"gobox/adventure"
	"gobox/game"
)

// brParStates opens the states the old engine's bug-class sweeps ran over (rules/sweep_test.go, routed_test.go,
// quiet_test.go): exploring, a party turn of a fight and a shop, on game g.
func brParStates(g brGame) map[string]func(t *testing.T) *Engine {
	return map[string]func(t *testing.T) *Engine{
		"exploring": func(t *testing.T) *Engine { return brBegun(t, g) },
		"fight": func(t *testing.T) *Engine {
			e := brBegun(t, g)
			br1Join(t, e)
			br1OwnBlock(t, e, 201, br1AdjacentFight)
			for range br1FightCap {
				if myTurn(e.build()) {
					return e
				}
				if e.waiting(PendFight) {
					e.Step()
				} else {
					press(e)
				}
			}
			t.Fatal("no party turn")
			return e
		},
		"shop": func(t *testing.T) *Engine {
			return brRun(t, g, 201, "    PriceMultiplier = 16\n    Shop(1)")
		},
	}
}

// TestBR1NoSilentDo ports rules TestClassNoSilentAccept, TestBugClass45Routed (an offered action never panics or
// falls through unrouted) and TestClassStateStable (two builds are equal, nothing offered twice): every action
// the state offers, done on a fresh copy, changes the state, tells an event or is refused. Also runs for BR2.
func TestBR1NoSilentDo(t *testing.T) { noSilentDo(t, brParStates(gameBR1)) }

func TestBR2NoSilentDo(t *testing.T) { noSilentDo(t, brParStates(gameBR2)) }

// brStartsAtTitle is a new engine on game g's folder at its title, the state before any key.
func brStartsAtTitle(t *testing.T, g brGame) *Engine {
	t.Helper()
	e := brEngine(t, g)
	e.Open("")
	return e
}

// TestBR1TitleThenPartyMenu ports rules TestTitleThenMenu and TestStartIsBR1Title: the state at the start is the
// title scene whose intro is the manifest's Start sequence (adventure.TitleIntro, which the engine composes);
// the title's end is the party menu (game.Assembling), whose Quit ends the run with game.Ended (rules TestQuit).
func TestBR1TitleThenPartyMenu(t *testing.T) { brTitleThenPartyMenu(t, gameBR1) }

func TestBR2TitleThenPartyMenu(t *testing.T) { brTitleThenPartyMenu(t, gameBR2) }

func brTitleThenPartyMenu(t *testing.T, g brGame) {
	e := brStartsAtTitle(t, g)
	sc, ok := e.build().(game.Scene)
	if !ok || sc.Kind != game.SceneTitle || sc.Intro == nil {
		t.Fatalf("start: %T, kind %v", e.build(), sc.Kind)
	}
	want, _, err := adventure.TitleIntro(e.title.start, e.FullIntro.Load(), e.graphics())
	if err != nil || !slices.Equal(sc.Intro.Pictures, want.Pictures) {
		t.Errorf("the title's pictures are not the Start sequence's: %v", err)
	}
	e.titleDone()
	e.Advance()
	if _, ok := e.build().(game.Assembling); !ok {
		t.Fatalf("after the title %T, want the party menu", e.build())
	}
	quitEnds(t, e)
}

// TestBR1AimKeepsFighting ports rules TestClassAimedSquares and TestClassSubAskKeepsState: with the aim cursor up
// the state is still Fighting, it carries no Ask (a square ask is answered by the aimed actions, not a question),
// and every square the aimed actions offer is one the ruleset's aim check accepts.
func TestBR1AimKeepsFighting(t *testing.T) {
	e, _, _, _ := br1AimTurn(t)
	f, ok := e.build().(game.Fighting)
	if !ok || f.Ask != nil {
		t.Fatalf("state %T, ask %+v", e.build(), f.Ask)
	}
	offered := 0
	for _, a := range f.Allowed.Aimed {
		for _, sq := range a.Squares {
			offered++
			if why := e.cs.aim.check(sq).Why; why != "" {
				t.Errorf("%T offered at %v is refused: %q", a.Action, sq, why)
			}
		}
	}
	if offered == 0 {
		t.Error("no square offered with the aim up")
	}
}

// TestBR1SheetEverywhere ports rules TestClassSheetEverywhere: every member's sheet is published in a fight and in a
// shop (the party list of the state carries it), as the game's own.
func TestBR1SheetEverywhere(t *testing.T) { brSheetEverywhere(t, gameBR1) }

func TestBR2SheetEverywhere(t *testing.T) { brSheetEverywhere(t, gameBR2) }

func brSheetEverywhere(t *testing.T, g brGame) {
	states := brParStates(g)
	for _, name := range []string{"fight", "shop"} {
		var party []game.Member
		switch st := states[name](t).build().(type) {
		case game.Fighting:
			party = st.Party
		case game.Trading:
			party = st.Party
		default:
			t.Fatalf("%s: state %T", name, st)
		}
		if len(party) == 0 {
			t.Fatalf("%s: no party", name)
		}
		for i, m := range party {
			if m.Page == nil || m.Page.Name != m.Name {
				t.Errorf("%s: member %d (%s) has no sheet", name, i, m.Name)
			}
		}
	}
}

// TestBR1QuietAnswer ports rules TestClassQuietAnswer (bug class 79): with nothing asked the state offers no Reply
// or Continue (no Enter, Yes or No).
func TestBR1QuietAnswer(t *testing.T) {
	e := brBegun(t, gameBR1)
	c := commonOf(e.build())
	if c.Ask != nil {
		t.Skipf("still asked: %+v", c.Ask)
	}
	for _, a := range c.Allowed.Plain {
		switch a.(type) {
		case game.Reply, game.ContinueAction:
			t.Errorf("%T offered with nothing asked", a)
		}
	}
}
