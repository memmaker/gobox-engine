package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

// askRules asks q (Pending, in ModeAsk) and keeps the decisions.
type askRules struct {
	seqRules
	q *game.Ask
}

func (r *askRules) State() game.State                 { return game.State{Mode: game.ModeAsk, Pending: r.q} }
func (r *askRules) Apply(d game.Decision) game.Result { r.got = append(r.got, d); return game.Result{} }

// Yes and No answer an ask's options labelled so; without options Yes is 1.
func TestYesNoOptions(t *testing.T) {
	opts := []game.Option{{Label: "Yes", Do: game.Search{}}, {Label: "No", Do: game.Cancel{}}}
	for _, c := range []struct {
		q    *game.Ask
		k    front.ActKind
		want game.Answer
	}{{&game.Ask{Kind: game.YesNo, Options: opts}, front.ActYes, game.Answer{N: 0}},
		{&game.Ask{Kind: game.YesNo, Options: opts}, front.ActNo, game.Answer{N: 1}},
		{&game.Ask{Kind: game.YesNo}, front.ActYes, game.Answer{N: 1}},
		{&game.Ask{Kind: game.YesNo}, front.ActNo, game.Answer{N: 0}}} {
		r := &askRules{q: c.q}
		NewGame(r, game.Data{}).Answer(front.Action{Kind: c.k})
		if len(r.got) != 1 || r.got[0] != c.want {
			t.Errorf("%v on %d options: %#v, want %#v", c.k, len(c.q.Options), r.got, c.want)
		}
	}
	// a Refusal's Yes/No (Attack Ally): Yes does the Yes option's Do
	r := &askRules{}
	g := NewGame(r, game.Data{})
	g.asked = &game.Ask{Kind: game.YesNo, Options: opts}
	g.Answer(front.Action{Kind: front.ActYes})
	if len(r.got) != 1 || r.got[0] != game.Decision(game.Search{}) {
		t.Errorf("refusal Yes: %#v", r.got)
	}
}
