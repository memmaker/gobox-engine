package present

import (
	"testing"

	"gobox/front"
	"gobox/game/play"
)

// TestBarOverPrompt: a game prompt on the creation's bar row must not show
// through the bar's gaps ("ROLLAMINIMALMKEEP EXIT").
func TestBarOverPrompt(t *testing.T) {
	v := front.View{Prompt: []front.PaneRun{{X: 0, Text: "xxxxAxxxxxxxMxxxxxxxxx"}}}
	p, _ := draftOn(&play.Draft{})
	c := p.screenPage(v)
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
