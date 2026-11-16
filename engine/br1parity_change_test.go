package engine

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"gobox/game"
	"gobox/rules"
	"gobox/rulesxxvc"
)

// The Change menu (explore.pas CampAlterMenu) and the party menu's Remove and Purge (chargen.pas DropCharacterMenu): ports rules
// TestRemove and TestPurge. Messages and Anim are the front's settings (the menu has no such rows).

func asked(t *testing.T, e *Engine) []string {
	t.Helper()
	q := commonOf(e.build()).Ask
	if q == nil {
		t.Fatalf("no question on %T", e.build())
	}
	return q.Replies
}

func reply(t *testing.T, e *Engine, n int) {
	t.Helper()
	if _, err := e.doNow(game.Reply{N: n}); err != nil {
		t.Fatal(err)
	}
}

// brResting is a game of two members with the rest menu up in a block of the test's own.
func brResting(t *testing.T, g brGame) *Engine {
	t.Helper()
	e := br1BeginWith(t, brEngine(t, g), 2)
	br1OwnBlock(t, e, 202, "    Program(9)\n    Text(\"AFTER THE REST\")\n    PressKey()")
	if m := asked(t, e); !slices.Equal(m, []string{"Rest", "Save", "Change", "Exit"}) {
		t.Fatalf("the rest menu: %q", m)
	}
	return e
}

func TestBR1ChangeLevel(t *testing.T) { brChangeLevel(t, gameBR1) }
func TestBR2ChangeLevel(t *testing.T) { brChangeLevel(t, gameBR2) }

func brChangeLevel(t *testing.T, g brGame) {
	e := brResting(t, g)
	reply(t, e, restMenuChange)
	if m := asked(t, e); !slices.Equal(m, []string{"Order", "Purge", "Level", "Exit"}) {
		t.Fatalf("the Change menu: %q", m)
	}
	reply(t, e, int(changeLevel))
	if m := asked(t, e); !slices.Equal(m, []string{"Novice", "Cadet", "Trooper", "Veteran", "Ace", "Exit"}) {
		t.Fatalf("the level menu: %q", m)
	}
	reply(t, e, 3) // Veteran
	level := func() rules.Difficulty {
		b, err := e.ruleset().(*rulesxxvc.Rules).Save()
		if err != nil {
			t.Fatal(err)
		}
		var s struct{ Difficulty rules.Difficulty }
		if err := json.Unmarshal(b, &s); err != nil {
			t.Fatal(err)
		}
		return s.Difficulty
	}
	if level() != rules.DifficultyHarder {
		t.Errorf("level %d, want Veteran (%d)", level(), rules.DifficultyHarder)
	}
	if m := asked(t, e); len(m) != 4 || m[0] != "Order" { // the Change menu again
		t.Fatalf("after the level: %q", m)
	}
	reply(t, e, int(changeExit))
	if m := asked(t, e); !slices.Equal(m, []string{"Rest", "Save", "Change", "Exit"}) {
		t.Fatalf("Exit goes back to the rest menu: %q", m)
	}
}

func TestBR1ChangeOrder(t *testing.T) { brChangeOrder(t, gameBR1) }
func TestBR2ChangeOrder(t *testing.T) { brChangeOrder(t, gameBR2) }

// brChangeOrder: Order picks a member and moves it down a place (OrderMoveDown), Exit goes back.
func brChangeOrder(t *testing.T, g brGame) {
	e := brResting(t, g)
	was := slices.Clone(e.partyIDs())
	reply(t, e, restMenuChange)
	reply(t, e, int(changeOrder))
	reply(t, e, 0) // the first member
	if m := asked(t, e); !slices.Equal(m, []string{"Up", "Down", "Exit"}) {
		t.Fatalf("the place menu: %q", m)
	}
	reply(t, e, orderDown)
	if got := e.partyIDs(); got[0] != was[1] || got[1] != was[0] {
		t.Errorf("order %v, was %v: the first member did not move down", got, was)
	}
	reply(t, e, orderUp) // and back up
	if got := e.partyIDs(); !slices.Equal(got, was) {
		t.Errorf("order %v, want %v", got, was)
	}
	reply(t, e, 2) // Exit
	if m := asked(t, e); len(m) != 4 || m[0] != "Order" {
		t.Fatalf("after Exit: %q", m)
	}
}

func TestBR1ChangePurge(t *testing.T) { brChangePurge(t, gameBR1) }
func TestBR2ChangePurge(t *testing.T) { brChangePurge(t, gameBR2) }

// brChangePurge ports rules TestPurge: two questions (default No), then the saved files and the member go; a No keeps it.
func brChangePurge(t *testing.T, g brGame) {
	e := brResting(t, g)
	who := e.partyIDs()[1]
	name := e.ruleset().Sheet(who).Name
	if err := e.SaveCharacter(who, ""); err != nil {
		t.Fatal(err)
	}
	file := e.savePath(e.charBase(who) + ".cha")
	if !e.fileExists(file) {
		t.Fatalf("no saved file %s", file)
	}
	reply(t, e, restMenuChange)
	reply(t, e, int(changePurge))
	reply(t, e, 1) // the second member
	if q := e.question(); q == nil || q.Text != "Drop "+name+" forever? " || !q.YesNo || q.Default != 0 {
		t.Fatalf("first question: %+v", q)
	}
	reply(t, e, 0) // No
	if e.partySize() != 2 || !strings.Contains(e.ext.note, "sigh of relief") {
		t.Errorf("a No keeps the member: party %d note %q", e.partySize(), e.ext.note)
	}
	reply(t, e, int(changePurge))
	reply(t, e, 1)
	reply(t, e, 1) // Yes
	if q := e.question(); q == nil || q.Text != "Are you sure? " || q.Default != 0 {
		t.Fatalf("second question: %+v", q)
	}
	reply(t, e, 1) // Yes
	if e.partySize() != 1 || e.fileExists(file) || slices.Contains(e.BuildCharList(), name) {
		t.Errorf("purged: party %d, file kept %v, listed %v", e.partySize(), e.fileExists(file), slices.Contains(e.BuildCharList(), name))
	}
	if m := asked(t, e); len(m) != 4 || m[0] != "Order" {
		t.Fatalf("the Change menu again: %q", m)
	}
}

// TestBR1ChangeOnOrbit: on the orbit map the Change menu has no Order and no Purge (CampAlterMenu: ORBIT_MAP).
func TestBR1ChangeOnOrbit(t *testing.T) { brChangeOnOrbit(t, gameBR1) }
func TestBR2ChangeOnOrbit(t *testing.T) { brChangeOnOrbit(t, gameBR2) }

func brChangeOnOrbit(t *testing.T, g brGame) {
	e := brFlying(t, g)
	reply(t, e, rocketMenuChange)
	if m := asked(t, e); !slices.Equal(m, []string{"Level", "Exit"}) {
		t.Fatalf("the orbit map's Change menu: %q", m)
	}
	reply(t, e, 1) // Exit
	if m := asked(t, e); len(m) != 4 || m[0] != "Move" {
		t.Fatalf("Exit goes back to the rocket menu: %q", m)
	}
}

func TestBR1PartyRemove(t *testing.T) { brPartyRemove(t, gameBR1) }
func TestBR2PartyRemove(t *testing.T) { brPartyRemove(t, gameBR2) }

// brPartyRemove ports rules TestRemove: Remove saves the member (its record keeps its items, the stash too) and takes it
// out of the party; with no team left Begin is refused.
func brPartyRemove(t *testing.T, g brGame) {
	e := brAtPartyMenu(t, g)
	who := e.partyIDs()[0]
	name := e.ruleset().Sheet(who).Name
	file := e.savePath(e.charBase(who) + ".cha")
	if _, err := e.doNow(game.RemoveCharacter{Roster: 0}); err != nil {
		t.Fatal(err)
	}
	if e.partySize() != 0 || !e.fileExists(file) || !slices.Contains(e.BuildCharList(), name) {
		t.Errorf("removed: party %d, file %v, listed %v", e.partySize(), e.fileExists(file), slices.Contains(e.BuildCharList(), name))
	}
	if _, err := e.doNow(game.Begin{}); err == nil {
		t.Error("began with no team")
	}
}

func TestBR1PartyPurge(t *testing.T) { brPartyPurge(t, gameBR1) }
func TestBR2PartyPurge(t *testing.T) { brPartyPurge(t, gameBR2) }

// brPartyPurge ports rules TestPurge on the party menu: Purge deletes the saved files and takes the member out.
func brPartyPurge(t *testing.T, g brGame) {
	e := brAtPartyMenu(t, g)
	who := e.partyIDs()[0]
	name := e.ruleset().Sheet(who).Name
	if err := e.SaveCharacter(who, ""); err != nil {
		t.Fatal(err)
	}
	file := e.savePath(e.charBase(who) + ".cha")
	if _, err := e.doNow(game.DropMember{Member: 0}); err != nil {
		t.Fatal(err)
	}
	if e.partySize() != 0 || e.fileExists(file) || slices.Contains(e.BuildCharList(), name) {
		t.Errorf("purged: party %d, file kept %v, listed %v", e.partySize(), e.fileExists(file), slices.Contains(e.BuildCharList(), name))
	}
}
