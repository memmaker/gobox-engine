package ui

import (
	"testing"

	"gobox/game/play"
)

// TestNoFrontScreenOverAsk: F10 opens no slot list while the game's yes/no
// is up (BR2's "Quit to DOS?").
func TestNoFrontScreenOverAsk(t *testing.T) {
	p := &Presenter{}
	g, f := slotsGame(t)
	f.st = play.Exploring{Common: play.Common{Ask: &play.Question{Text: "Quit to DOS?", YesNo: true, Replies: []string{"No", "Yes"}}}}
	p.SetEngine(g)
	p.Key(KeyF10)
	if p.slots != nil || p.fullMap != nil || len(p.inv.acts) > 0 {
		t.Fatalf("F10 over the ask: slots %v acts %v", p.slots, p.inv.acts)
	}
}
