package present

import (
	"testing"

	"gobox/front"
)

// A Blink run shows the highlight at first and its own ink once the swaps are done.
func TestBlink(t *testing.T) {
	p := &Presenter{}
	view := func() front.View {
		return front.View{ViewText: [][]front.PaneRun{{{Text: "x", Blink: true, Fg: col(inkGreen)}}}}
	}
	if got := p.blinked(view()).ViewText[0][0].Fg; got != col(inkWhite) {
		t.Fatalf("first %v", got)
	}
	p.blinkAt = now().Add(-(blinkSwaps + 1) * blinkMs)
	if got := p.blinked(view()).ViewText[0][0].Fg; got != col(inkGreen) {
		t.Fatalf("last %v", got)
	}
}
