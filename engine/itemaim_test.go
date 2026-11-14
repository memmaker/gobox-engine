package engine

import (
	"slices"
	"testing"

	"gobox/game"
	"gobox/rules"
	"gobox/rules1e"
)

// TestItemAimOffered: an item whose spell takes a square (a wand) opens its
// aim in the turn; the state offers it (the Aimed Cast, Leave to give it
// up). Nothing was offered: the bar's Items did nothing.
func TestItemAimOffered(t *testing.T) {
	e := partyTurnScene(t)
	c := e.member(e.cs.actor)
	s := e.state()
	s.AddItemCopy(c, s.RandomItem(rules1e.ItemWand))
	e.turnChoices(e.charID(c))
	e.rebuild()
	f := fighting(e)
	i := slices.IndexFunc(f.Allowed.Plain, func(a game.Action) bool { u, ok := a.(game.UseItem); return ok && u.Item != rules.NoItem })
	if i < 0 {
		t.Fatal("the wand is not offered")
	}
	if _, err := e.Do(f.Allowed.Plain[i]); err != nil {
		t.Fatal(err)
	}
	if e.cs.aim == nil {
		t.Skip("the wand's spell takes no aim")
	}
	al := fighting(e).Allowed
	if len(al.Aimed) != 1 || !playAllows(fighting(e), game.Leave{}) {
		t.Fatalf("aim up, offered %+v", al.Plain)
	}
	if _, ok := al.Aimed[0].Action.(game.Cast); !ok {
		t.Fatalf("aimed %T, want the spell's Cast", al.Aimed[0].Action)
	}
	if _, err := e.Do(game.Leave{}); err != nil {
		t.Fatal(err)
	}
	if e.cs.aim != nil || !playAllows(fighting(e), game.EndTurn{}) {
		t.Error("Leave did not give the aim up for the turn's menu")
	}
}
