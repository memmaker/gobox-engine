package playrules

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// Enter with nothing asked (no Scene pending) must be a quiet no-op, not an error toast.
func TestAnswerNothingAskedIsQuiet(t *testing.T) {
	r := &rules{}
	if res := r.answer(play.Exploring{}, game.Answer{}); res.Err != nil {
		t.Fatalf("answer with nothing asked returned %v", res.Err)
	}
}
