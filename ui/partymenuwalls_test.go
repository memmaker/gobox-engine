package ui

import (
	"testing"

	"gobox/game"
)

// TestPartyMenuKeepsWalls: the party menu never drops the 3D view: it
// shows the engine's (a loaded game's hall), else the one last explored.
func TestPartyMenuKeepsWalls(t *testing.T) {
	kept := &game.Walls{Key: "explored"}
	if shown, keep := wallsOf(game.Assembling{}, kept); shown != kept || keep != kept {
		t.Errorf("party menu dropped the explored view: %v %v", shown, keep)
	}
	w := game.MapWalls{Side: [][][4]int{{{1}}}}
	if shown, _ := wallsOf(game.Assembling{Walls: &w}, nil); shown == nil {
		t.Error("party menu ignores the engine's 3D view")
	}
}
