package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

// TestMessagesHold: two fight messages in a row each stay up the
// game-speed delay (speed × 100 ms) before the next.
func TestMessagesHold(t *testing.T) {
	b := &front.Battle{}
	es := []game.Event{{Kind: game.EvMessage, Text: "Sylvaine is bandaged"}, {Kind: game.EvMessage, Text: "SWORDSMAN Attacks"}}
	out := combatEvents(es, front.View{Battle: b}, func(game.Event) *front.Battle { return b }, 400)
	if len(out) != 2 || out[0].Ms != 400 || out[1].Ms != 400 || out[0].Battle == nil || out[1].Battle == nil {
		t.Fatalf("%+v", out)
	}
	g := NewGame(&seqRules{}, game.Data{})
	if g.do(game.SetSpeed{N: 2}); g.messageMs() != 200 {
		t.Fatalf("speed 2: %d ms", g.messageMs())
	}
}
