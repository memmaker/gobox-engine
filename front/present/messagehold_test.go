package present

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// TestMessagesHold: two fight messages in a row each stay up the
// game-speed delay (speed × 100 ms) before the next.
func TestMessagesHold(t *testing.T) {
	b := &play.Combat{}
	es := []play.Report{{Kind: game.EvMessage, Text: "Sylvaine is bandaged"}, {Kind: game.EvMessage, Text: "SWORDSMAN Attacks"}}
	out := playedEvents(es, true, func(play.Report) (*play.Combat, []game.Member) { return b, nil }, 400)
	if len(out) != 2 || out[0].Ms != 400 || out[1].Ms != 400 || out[0].Battle == nil || out[1].Battle == nil {
		t.Fatalf("%+v", out)
	}
	g := NewGame(&fake{st: play.Exploring{}}, game.Data{})
	if g.r.send(tSetSpeed{N: 2}); g.messageMs() != 200 {
		t.Fatalf("speed 2: %d ms", g.messageMs())
	}
}

// TestFightRefusalHeld: a refusal in a fight (POR's "Blocked") is a fight
// message too, held on the message row, not only the toast.
func TestFightRefusalHeld(t *testing.T) {
	r := &fake{st: play.Fighting{}, err: &play.Refusal{Text: "Blocked"}}
	g := NewGame(r, game.Data{})
	g.Settle()
	g.Answer(front.Action{Kind: front.ActWalk, X: 1, Y: 1})
	es := g.legacySettle().Events
	if len(es) != 1 || es[0].Text != "Blocked" || es[0].Ms == 0 {
		t.Fatalf("events %+v", es)
	}
}
