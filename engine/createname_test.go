package engine

import (
	"slices"
	"testing"

	"gobox/game"
)

// TestCreateNameReply: Keep on an unnamed draft asks the name; the typed
// Reply names and keeps the character.
func TestCreateNameReply(t *testing.T) {
	e := testEngine(t)
	e.viewBit0 = 1
	e.Init()
	e.Start(e.Run)
	for range 5 { // past the title
		if sc, ok := e.build().(game.Scene); !ok || sc.Kind != game.SceneTitle {
			break
		}
		if _, err := e.doNow(game.ContinueAction{}); err != nil {
			t.Fatal("Continue:", err)
		}
	}
	if _, err := e.doNow(game.NewCharacter{}); err != nil {
		t.Fatal("NewCharacter:", err)
	}
	if _, err := e.doNow(game.KeepCharacter{}); err != nil {
		t.Fatal("KeepCharacter:", err)
	}
	as, _ := e.build().(game.Assembling)
	if as.Draft == nil || as.Draft.Ask == nil || !as.Draft.Ask.Typed {
		t.Fatalf("no name asked: %+v", as.Draft)
	}
	if !slices.ContainsFunc(as.Allowed.Plain, func(a game.Action) bool { _, ok := a.(game.Reply); return ok }) {
		t.Fatalf("no Reply allowed: %v", as.Allowed.Plain)
	}
	if _, err := e.doNow(game.Reply{Text: "BOB"}); err != nil {
		t.Fatal("Reply:", err)
	}
	if e.creation != nil {
		t.Fatal("draft still up after the name")
	}
	if r := e.build().(game.Assembling).Roster; !slices.Contains(r, "BOB") {
		t.Fatalf("BOB not in the roster %v", r)
	}
}

// TestCreateRename: the creation sheet's name (Rename) is taken, and Keep
// then keeps without asking it.
func TestCreateRename(t *testing.T) {
	e := testEngine(t)
	e.viewBit0 = 1
	e.Init()
	e.Start(e.Run)
	for range 5 {
		if sc, ok := e.build().(game.Scene); !ok || sc.Kind != game.SceneTitle {
			break
		}
		if _, err := e.doNow(game.ContinueAction{}); err != nil {
			t.Fatal("Continue:", err)
		}
	}
	if _, err := e.doNow(game.NewCharacter{}); err != nil {
		t.Fatal("NewCharacter:", err)
	}
	if _, err := e.doNow(game.Rename{Name: "ANN"}); err != nil {
		t.Fatal("Rename:", err)
	}
	if _, err := e.doNow(game.KeepCharacter{}); err != nil {
		t.Fatal("KeepCharacter:", err)
	}
	if e.creation != nil {
		t.Fatal("Keep asked a name though one was set")
	}
	if r := e.build().(game.Assembling).Roster; !slices.Contains(r, "ANN") {
		t.Fatalf("ANN not in the roster %v", r)
	}
}
