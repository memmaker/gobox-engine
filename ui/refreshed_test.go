package ui

import (
	"testing"
	"time"

	"gobox/game"
)

// TestRefreshedSaid: camp's Fix done (game.Refreshed) is one sound and the
// words of what was restored.
func TestRefreshedSaid(t *testing.T) {
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
		t.Fatalf("%d sounds, said %+v", sounds, m)
	}
	if m := newRefreshed(game.Refreshed{Healed: true}); len(m.lines) != 1 {
		t.Errorf("healed only: %v", m.lines)
	}
}

// TestFixMoment: picking Fix darkens the screen first; only once it is
// black is the rest sent; the box then waits for a key, whatever the time,
// and a key fades the screen in again. Keys before that act on nothing.
func TestFixMoment(t *testing.T) {
	f := &fake{st: game.Camping{}, out: game.Outcome{Events: []game.Event{game.Refreshed{Healed: true}}}}
	g := &classicGame{Game: NewGame(f, game.Data{})}
	p := &Presenter{}
	p.SetEngine(g)
	defer func(s bool, n int64) { Scripted, frames = s, n }(Scripted, frames)
	Scripted, frames = true, 0 // now() by the frame count
	wait := func(d time.Duration) { frames += int64(d * 60 / time.Second) }
	run := func() { // a frame's step, then the calls the screens queued
		p.fixStep()
		for _, c := range p.calls {
			c(g)
		}
		p.calls = nil
	}
	p.startFix(7)
	if run(); p.fix.phase != fixFadeOut || len(f.sent) != 0 || p.fix.dark() != 0 {
		t.Fatalf("at once: phase %v sent %v dark %v", p.fix.phase, f.sent, p.fix.dark())
	}
	if !p.fixInput() || p.fix.phase != fixFadeOut {
		t.Fatal("a key while it fades out is not swallowed")
	}
	wait(refreshFade + 50*time.Millisecond)
	if run(); p.fix.phase != fixWorking || len(f.sent) != 0 || p.fix.dark() != 1 {
		t.Fatalf("black: phase %v sent %v dark %v, want nothing sent before a black frame", p.fix.phase, f.sent, p.fix.dark())
	}
	run() // the black frame drawn
	run() // the rest sent and answered
	if len(f.sent) != 1 || f.sent[0] != game.Action(game.CampRest{Kind: 7}) {
		t.Fatalf("sent %v, want the rest", f.sent)
	}
	if run(); p.fix.phase != fixSaid || len(p.fix.lines) != 1 {
		t.Fatalf("answered: phase %v lines %v", p.fix.phase, p.fix.lines)
	}
	wait(time.Minute)
	if run(); p.fix.phase != fixSaid || p.fix.dark() != 1 {
		t.Fatalf("a minute on: phase %v, want the box still up", p.fix.phase)
	}
	if !p.fixInput() || p.fix.phase != fixFadeIn {
		t.Fatalf("a key on the box: phase %v, want the fade in", p.fix.phase)
	}
	wait(refreshFade + 50*time.Millisecond)
	if run(); p.fix.phase != fixNone || p.fixInput() {
		t.Fatalf("faded in: phase %v", p.fix.phase)
	}
}
