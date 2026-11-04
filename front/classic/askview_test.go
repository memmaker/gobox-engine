package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

// With Ask.View the View action on a highlighted member is ViewMember; without, it is not.
func TestAskView(t *testing.T) {
	for _, view := range []bool{true, false} {
		g := NewGame(&seqRules{}, game.Data{})
		g.asked = &game.Ask{Kind: game.PickMember, View: view}
		d, ok := g.decision(front.Action{Kind: front.ActView, N: 2})
		if v, isView := d.(game.ViewMember); view != (ok && isView && v.Member == 2) {
			t.Errorf("View %v: %v %v", view, d, ok)
		}
	}
}
