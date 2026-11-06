package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

// TestTypedNumberN: a typed amount is the answer's N too, its Text kept.
func TestTypedNumberN(t *testing.T) {
	g := &Game{R: &loreRules{s: game.State{Mode: game.ModeAsk, Pending: &game.Ask{Kind: game.Number}}}}
	d, ok := g.decision(front.Action{Kind: front.ActText, Text: "100"})
	if a, _ := d.(game.Answer); !ok || a.N != 100 || a.Text != "100" {
		t.Fatalf("decision %+v", d)
	}
}
