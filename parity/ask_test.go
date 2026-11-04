package parity

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"por/rules"
)

// askIs says the State asks a question of kind (and at's stop is por's).
func askIs(t *testing.T, name string, e *rules.Engine, kind game.AskKind) {
	t.Helper()
	at(t, name, e)
	if s := e.Rules().State(); s.Mode != game.ModeAsk || s.Pending == nil || s.Pending.Kind != kind {
		t.Fatalf("%s: mode %v pending %+v, want kind %d", name, s.Mode, s.Pending, kind)
	}
}

func TestAskParity(t *testing.T) {
	p := engine(t, true, "parlay")
	settled(p, front.Action{})
	askIs(t, "parlay encounter", p, game.Choice)
	p.Rules().Apply(game.Answer{N: 3}) // PARLAY
	settled(p, front.Action{})
	askIs(t, "parlay talk", p, game.Choice)

	w := engine(t, true, "password")
	settled(w, front.Action{})
	askIs(t, "password", w, game.Name)
	w.Rules().Apply(game.Answer{Text: "SWORD"})
	settled(w, front.Action{})
	askIs(t, "password said", w, game.Choice)

	m := engine(t, true, "who")
	settled(m, front.Action{})
	askIs(t, "who", m, game.PickMember)
	m.Rules().Apply(game.Answer{N: 1})
	settled(m, front.Action{})
	at(t, "who answered", m)
	if s := m.Rules().State(); s.Pending != nil {
		t.Errorf("who answered: pending %+v", s.Pending)
	}
}

// TestAskSweep: State's question at each stop of the dev scenes, answered on
// with Enter; typed text, pauses and lists among them.
func TestAskSweep(t *testing.T) {
	seen := map[game.AskKind]int{}
	for _, scene := range []string{"password", "parlay", "who", "journal", "talljournal", "runes", "lockeddoor", "shop", "temple", "appraise", "loot", "effects", "autocombat"} {
		e := engine(t, true, scene)
		settled(e, front.Action{})
		for i := 0; i < 12; i++ {
			at(t, scene, e)
			a := front.Action{Kind: front.ActOK}
			if q := e.Rules().State().Pending; q != nil && e.Rules().State().Mode == game.ModeAsk {
				seen[q.Kind]++
				if q.Kind == game.Name && q.Typed == "" {
					a = front.Action{Kind: front.ActText, Text: "SWO"}
				}
			}
			settled(e, a)
		}
	}
	t.Logf("asks by kind: %v", seen)
	for _, k := range []game.AskKind{game.Choice, game.Name, game.PickMember} {
		if seen[k] == 0 {
			t.Errorf("no ask of kind %d seen: %v", k, seen)
		}
	}
}
