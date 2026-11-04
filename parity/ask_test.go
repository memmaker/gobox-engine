package parity

import (
	"reflect"
	"testing"

	"gobox/front"
	"gobox/front/classic"
	"gobox/game"
	"por/rules"
)

// sameAsk says classic's question equals por's (kind: the Pending's).
func sameAsk(t *testing.T, name string, e *rules.Engine, kind game.AskKind) {
	t.Helper()
	v, s := e.View(), e.Rules().State()
	if v.Ask == nil || s.Pending == nil || s.Pending.Kind != kind {
		t.Fatalf("%s: ask %+v pending %+v, want kind %d", name, v.Ask, s.Pending, kind)
	}
	if c := classic.View(v, s); !reflect.DeepEqual(c.Ask, v.Ask) {
		t.Errorf("%s: ask\n got %+v\nwant %+v", name, c.Ask, v.Ask)
	}
}

func TestAskParity(t *testing.T) {
	p := engine(t, true, "parlay")
	settled(p, front.Action{})
	sameAsk(t, "parlay encounter", p, game.Choice)
	p.Rules().Apply(game.Answer{N: 3}) // PARLAY
	settled(p, front.Action{})
	sameAsk(t, "parlay talk", p, game.Choice)

	w := engine(t, true, "password")
	settled(w, front.Action{})
	sameAsk(t, "password", w, game.Name)
	w.Rules().Apply(game.Answer{Text: "SWORD"})
	settled(w, front.Action{})
	sameAsk(t, "password said", w, game.Choice)

	m := engine(t, true, "who")
	settled(m, front.Action{})
	sameAsk(t, "who", m, game.PickMember)
	m.Rules().Apply(game.Answer{N: 1})
	if v := settled(m, front.Action{}); v.Ask != nil || classic.View(v, m.Rules().State()).Ask != nil {
		t.Errorf("who answered: ask %+v", v.Ask)
	}
}
