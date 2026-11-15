package ui

import (
	"testing"

	"gobox/game"
)

// TestSpeakReachesGame: a journal page's speech clips go to the game as
// Speak (its voice plays them), not as sound effects.
func TestSpeakReachesGame(t *testing.T) {
	g, f := playGame(game.Exploring{})
	if res := g.r.speak([]string{"Journal_Entry_11"}); res.Err != nil || len(res.Events) != 0 {
		t.Fatalf("speak: %+v", res)
	}
	if s, ok := f.sent[len(f.sent)-1].(game.Speak); len(f.sent) != 1 || !ok || len(s.Keys) != 1 || s.Keys[0] != "Journal_Entry_11" {
		t.Fatalf("sent %v, want one Speak of the clip", f.sent)
	}
}
