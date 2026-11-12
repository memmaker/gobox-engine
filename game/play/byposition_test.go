package play

import (
	"testing"

	"gobox/game"
)

// TestByPosition: the party's items and the loot get their places' handles, and Position gives the place back.
func TestByPosition(t *testing.T) {
	s := ByPosition(Looting{Common: Common{Party: []game.Member{{Items: make([]game.Item, 2)}}}, Items: make([]game.Item, 3)}).(Looting)
	if h := s.Party[0].Items[1].Handle; game.Position(h) != 1 {
		t.Fatalf("party item 1: handle %d, position %d", h, game.Position(h))
	}
	if h := s.Items[2].Handle; h != 3 || game.Position(h) != 2 {
		t.Fatalf("loot item 2: handle %d", h)
	}
}
