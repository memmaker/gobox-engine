package ui

import (
	"image"
	"testing"

	"gobox/game"
)

// TestWhoTakesCard: a spell's "Who?" takes member 2 from a click on its
// card, its name's line and F2, exploring and over a page (the items).
func TestWhoTakesCard(t *testing.T) {
	for _, page := range []bool{false, true} {
		for _, how := range []string{"click", "F2", "line"} {
			p := partyOn(game.Member{Name: "A"}, game.Member{Name: "B"}, game.Member{Name: "C"})
			p.active, p.pageMenu = true, page
			p.cards = []image.Rectangle{image.Rect(0, 0, 10, 10), image.Rect(20, 0, 30, 10), image.Rect(40, 0, 50, 10)}
			got := -1
			AskForMember(p, "Who?", 0, func(n int) { got = n })
			p.menuRects = []image.Rectangle{image.Rect(0, 50, 50, 60), image.Rect(0, 60, 50, 70), image.Rect(0, 70, 50, 80)}
			switch how {
			case "line":
				p.Click(25, 65)
			case "click":
				p.Click(25, 5)
			default:
				p.Key(KeyF1 + 1<<8)
			}
			for _, a := range p.inv.acts {
				AskAct(p, a)
			}
			if got != 1 {
				t.Errorf("page %v, %s: picked %d, want 1", page, how, got)
			}
		}
	}
}

// TestWhoCancels: Esc and a right click give a spell's "Who?" up: no
// member is taken and the question is gone.
func TestWhoCancels(t *testing.T) {
	for _, how := range []string{"esc", "right"} {
		p := partyOn(game.Member{Name: "A"}, game.Member{Name: "B"})
		p.active = true
		got := -1
		AskForMember(p, "Who?", 0, func(n int) { got = n })
		if how == "esc" {
			p.Key(27)
		} else {
			p.Mouse(5, 5, true, false, true, 0)
		}
		for _, a := range p.inv.acts {
			AskAct(p, a)
		}
		if got != -1 || p.choosing() {
			t.Errorf("%s: picked %d, still asking %v", how, got, p.choosing())
		}
	}
}
