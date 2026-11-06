package present

import (
	"testing"

	"gobox/front"
)

// TestNoFrontScreenOverAsk: F10 opens no slot list while the game's yes/no
// is up (BR2's "Quit to DOS?").
func TestNoFrontScreenOverAsk(t *testing.T) {
	v := front.View{Party: []front.Member{{Name: "A"}}, Ask: &front.Ask{Kind: front.AskConfirm, Menu: front.Menu{Prompt: "Quit to DOS?"}}}
	p := &Presenter{}
	p.SetEngine(&slotsGame{})
	p.Key(KeyF10, v)
	if p.slots != nil || p.fullMap != nil || len(p.inv.acts) > 0 {
		t.Fatalf("F10 over the ask: slots %v acts %v", p.slots, p.inv.acts)
	}
}

