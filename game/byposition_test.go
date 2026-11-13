package game

import (
	"testing"
)

// TestByPosition: the party's items and the loot get their places' handles, and Position gives the place back.
func TestByPosition(t *testing.T) {
	s := ByPositionState(Looting{Common: Common{Party: []Member{{Items: make([]Item, 2)}}}, Items: make([]Item, 3)}).(Looting)
	if h := s.Party[0].Items[1].Handle; Position(h) != 1 {
		t.Fatalf("party item 1: handle %d, position %d", h, Position(h))
	}
	if h := s.Items[2].Handle; h != 3 || Position(h) != 2 {
		t.Fatalf("loot item 2: handle %d", h)
	}
}
