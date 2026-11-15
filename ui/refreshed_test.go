package ui

import (
	"testing"
	"time"

	"gobox/game"
)

// TestRefreshedMoment: camp's Fix done (game.Refreshed) is one sound and
// one moment: a fade to black, the words of what was restored, a fade in.
func TestRefreshedMoment(t *testing.T) {
	f := &fake{st: game.Camping{}, out: game.Outcome{Events: []game.Event{game.Refreshed{Healed: true, Studied: true}}}}
	r := newRules(f, game.Data{})
	res := r.send(game.CampRest{})
	sounds := 0
	for _, e := range res.Events {
		if e.Kind == game.EvSound {
			sounds++
			if e.Sound != refreshedCue {
				t.Errorf("sound %q, want %q", e.Sound, refreshedCue)
			}
		}
	}
	m := r.refreshed
	if sounds != 1 || m == nil || len(m.lines) != 2 || m.lines[0] != "HP restored." || m.lines[1] != "Spells memorized." {
		t.Fatalf("%d sounds, moment %+v", sounds, m)
	}
	at := func(d time.Duration) (float32, bool, bool) { return m.dark(m.at.Add(d)) }
	if a, words, over := at(refreshFade / 2); a <= 0 || a >= 1 || words || over {
		t.Errorf("fading out: dark %v words %v over %v", a, words, over)
	}
	if a, words, over := at(refreshFade + refreshHold/2); a != 1 || !words || over {
		t.Errorf("held: dark %v words %v over %v", a, words, over)
	}
	if a, words, over := at(refreshFade + refreshHold + refreshFade/2); a <= 0 || a >= 1 || words || over {
		t.Errorf("fading in: dark %v words %v over %v", a, words, over)
	}
	if _, _, over := at(2*refreshFade + refreshHold); !over {
		t.Error("the moment never ends")
	}
	if m := newRefreshed(game.Refreshed{Healed: true}, time.Time{}); len(m.lines) != 1 {
		t.Errorf("healed only: %v", m.lines)
	}
}
