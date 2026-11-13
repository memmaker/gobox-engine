package present

import (
	"testing"
	"time"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// With no refusal's note the front's own toast shows (the shop's "buys a");
// a refusal's note goes over it.
func TestNoteKeepsOwn(t *testing.T) {
	p, _ := presenterOn(play.Exploring{})
	own := front.FrontNote{Text: "Brann buys a Broad Sword", At: time.Now()}
	p.eng.r.toast = own
	if got := p.note().Text; got != own.Text {
		t.Fatalf("note %q, want %q", got, own.Text)
	}
	p.eng.r.note = game.Note{Text: "x"}
	if got := p.note().Text; got != "x" {
		t.Fatalf("refusal note %q", got)
	}
}
