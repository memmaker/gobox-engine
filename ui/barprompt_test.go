package ui

import (
	"testing"

	"gobox/game/play"
)

// TestBarOverPrompt: a game prompt on the creation's bar row must not show
// through the bar's gaps ("ROLLAMINIMALMKEEP EXIT").
func TestBarOverPrompt(t *testing.T) {
	g, _ := playGame(play.Assembling{Common: play.Common{Status: "xxxxAxxxxxxxMxxxxxxxxx"}, Draft: &play.Draft{}})
	p := &Presenter{}
	p.SetEngine(g)
	c := p.screenPage()
	row := ""
	for _, ce := range c.cells[c.rows-1][:22] {
		if ce.set {
			row += string(ce.ch)
		} else {
			row += " "
		}
	}
	if row != "Roll Minimal Keep Exit" {
		t.Fatalf("bar row %q", row)
	}
}
