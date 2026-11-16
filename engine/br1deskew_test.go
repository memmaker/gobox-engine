package engine

import (
	"testing"

	"gobox/game"
)

// TestBR1DeskewFight: with the engine's Deskew option a BR1 dungeon fight runs on the straight arena (rulesxxvc
// buildDungeonArena, rules.StraightX): everyone the original places is placed, on a square that is no wall, facing
// 0 2 4 6. (The wall columns themselves: rulesxxvc TestDeskewWestColumn.)
func TestBR1DeskewFight(t *testing.T) {
	e := brBegun(t, gameBR1)
	br1Join(t, e)
	e.Deskew.Store(true)
	br1OwnBlock(t, e, 201, br1AdjacentFight)
	f, ok := e.build().(game.Fighting)
	if !ok {
		t.Fatalf("%T, want the fight", e.build())
	}
	if !f.Field.Dungeon {
		t.Fatal("not a dungeon fight")
	}
	if len(f.Field.Units) < 3 {
		t.Fatalf("%d units placed, want the party and two foes", len(f.Field.Units))
	}
	blocked := map[[2]int]bool{}
	for _, s := range f.Field.Squares {
		if s.Block {
			blocked[[2]int{s.At.X, s.At.Y}] = true
		}
	}
	for _, u := range f.Field.Units {
		if blocked[[2]int{u.At.X, u.At.Y}] {
			t.Errorf("%s stands on a wall at %v", u.Name, u.At)
		}
		if u.Facing%2 != 0 {
			t.Errorf("%s faces %d, want a quarter direction 0 2 4 6 on the straight board", u.Name, u.Facing)
		}
	}
}
