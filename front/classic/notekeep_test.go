package classic

import (
	"testing"
	"time"

	"gobox/front"
	"gobox/game"
)

// a State with no note keeps the View's own toast (the shop's "buys a",
// a refusal): classicGame lays Note over the Game's View
func TestNoteKeepsOwn(t *testing.T) {
	v := front.View{Note: front.Note{Text: "Brann buys a Broad Sword", At: time.Now()}}
	if got := Note(v, game.State{}).Note.Text; got != v.Note.Text {
		t.Fatalf("note %q, want %q", got, v.Note.Text)
	}
	if got := Note(v, game.State{Note: game.Note{Text: "x"}}).Note.Text; got != "x" {
		t.Fatalf("state note %q", got)
	}
}
