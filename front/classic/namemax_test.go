package classic

import (
	"testing"

	"gobox/game"
)

// TestNameMaxDefault: a name asked with no length (Max 0) still takes
// letters: the Gold Box name's 15.
func TestNameMaxDefault(t *testing.T) {
	q, _ := Ask(game.State{Mode: game.ModeAsk, Pending: &game.Ask{Kind: game.Name, Prompt: "Character name: "}})
	if q == nil || q.Max != 15 {
		t.Fatalf("ask %+v", q)
	}
}
