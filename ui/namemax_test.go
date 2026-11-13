package ui

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestNameMaxDefault: a name asked with no length (Max 0) still takes
// letters: the Gold Box name's 15.
func TestNameMaxDefault(t *testing.T) {
	q, _ := Ask(game.ModeAsk, &play.Ask{Kind: game.Name, Prompt: "Character name: "})
	if q == nil || q.Max != 15 {
		t.Fatalf("ask %+v", q)
	}
}
