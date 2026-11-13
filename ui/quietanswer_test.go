package ui

import (
	"image"
	"testing"

	"gobox/game/play"
)

// Enter with nothing asked (no Scene pending) is refused, but with no text,
// so no toast is raised (classic shows only a refusal that says why).
func TestAnswerNothingAskedIsQuiet(t *testing.T) {
	r := &frontState{}
	res := r.answer(play.Exploring{}, play.Reply{}, image.Point{})
	if res.Err == nil {
		t.Fatal("answer with nothing asked was not refused")
	}
	if res.Err.Error() != "" {
		t.Fatalf("refusal says %q: it would be toasted", res.Err.Error())
	}
}
