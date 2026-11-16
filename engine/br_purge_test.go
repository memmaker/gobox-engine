package engine

import (
	"slices"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// TestBR1PartyRemoveNPC: Remove on a non-player character at the party menu is the Drop question (chargen.pas MainMenu 'R':
// DropCharacterMenu); a No keeps it, a Yes (twice) purges it. A player character is still saved.
func TestBR1PartyRemoveNPC(t *testing.T) { brPartyRemoveNPC(t, gameBR1) }
func TestBR2PartyRemoveNPC(t *testing.T) { brPartyRemoveNPC(t, gameBR2) }

func brPartyRemoveNPC(t *testing.T, g brGame) {
	e := brAtPartyMenu(t, g)
	who := e.partyIDs()[0]
	name := e.ruleset().Sheet(who).Name
	e.setSelFact(rules.FactControl, 0xB0) // a non-player character
	if _, err := e.doNow(game.RemoveCharacter{Roster: 0}); err != nil {
		t.Fatal(err)
	}
	if q := e.question(); q == nil || q.Text != "Drop "+name+" forever? " {
		t.Fatalf("question %+v", q)
	}
	reply(t, e, 0) // No
	if e.partySize() != 1 {
		t.Fatalf("a No dropped it: party %d", e.partySize())
	}
	if _, err := e.doNow(game.RemoveCharacter{Roster: 0}); err != nil {
		t.Fatal(err)
	}
	reply(t, e, 1)
	reply(t, e, 1)
	if e.partySize() != 0 || slices.Contains(e.BuildCharList(), name) {
		t.Errorf("party %d, listed %v", e.partySize(), slices.Contains(e.BuildCharList(), name))
	}
}

// TestBR1PurgeLastMember: purging the last member from the Change menu does not end the game: the menu stays (the Pascal does
// nothing more) and its Exit ends the map loop, back to the party menu with no one in it.
func TestBR1PurgeLastMember(t *testing.T) { brPurgeLast(t, gameBR1) }
func TestBR2PurgeLastMember(t *testing.T) { brPurgeLast(t, gameBR2) }

func brPurgeLast(t *testing.T, g brGame) {
	e := brResting(t, g)
	reply(t, e, restMenuChange)
	for e.partySize() > 0 {
		reply(t, e, int(changePurge))
		reply(t, e, 0)
		reply(t, e, 1)
		out, err := e.doNow(game.Reply{N: 1})
		if err != nil {
			t.Fatal(err)
		}
		if slices.ContainsFunc(out.Events, func(ev game.Event) bool { _, ok := ev.(game.Ended); return ok }) {
			t.Fatal("the game ended")
		}
	}
	if m := asked(t, e); len(m) != 4 || m[0] != "Order" {
		t.Fatalf("the Change menu stays: %q", m)
	}
	reply(t, e, int(changeExit))
	reply(t, e, restMenuExit)
	if _, ok := e.build().(game.Assembling); !ok {
		t.Errorf("after the party was purged and the menus left: %T, want the party menu", e.build())
	}
}
