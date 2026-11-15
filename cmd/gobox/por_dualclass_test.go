package main

import (
	"testing"

	"gobox/game"
)

// TestHallChangeClassAsked: the hall's Change Class puts up the pick of the
// classes, then the yes/no with what the change costs; nothing changes
// before Yes (a key never answers a question that is not drawn).
func TestHallChangeClassAsked(t *testing.T) {
	f, eng := frontAndEngine(t, "dual")
	step := func(key uint16, rows int, class string) {
		t.Helper()
		f.Key(key)
		for range 30 {
			f.Frame(1280, 800)
			f.Tick()
		}
		if got := game.CommonOf(eng.State()).Party[0].Class; len(f.P.WinRects()) != rows || got != class {
			t.Fatalf("after %q: %d rows drawn, class %s; want %d, %s", rune(key), len(f.P.WinRects()), got, rows, class)
		}
	}
	step(uint16(game.ChangeClassFunction.Key), 4, "Fighter") // Cleric, Paladin, Ranger, Magic-User
	step('M', 2, "Fighter")                                  // Yes, No
	step('Y', 2, "Magic-User")
}
