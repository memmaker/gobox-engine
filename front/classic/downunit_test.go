package classic

import (
	"testing"

	"gobox/game"
)

// TestDeadUnitDown: a dead unit the game leaves on the map (POD's bulettes)
// is a body (Down), not a standing figure; a gone one is off the board.
func TestDeadUnitDown(t *testing.T) {
	s := game.State{Combat: &game.Combat{Units: []game.Unit{{ID: 1, Member: -1, Foe: true}, {ID: 2, Member: -1, Foe: true, Dead: true}, {ID: 3, Member: -1, Dead: true, Gone: true}}}}
	b := BattleOf(s.Combat, s.Party, nil)
	if len(b.Units) != 2 || b.Units[0].Down || !b.Units[1].Down {
		t.Fatalf("units %+v", b.Units)
	}
}
