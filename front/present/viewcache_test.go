package present

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
	"gobox/gpu"
	"image"
)

// counting is a fake that counts the State reads (every View build reads it).
type counting struct {
	*fake
	reads int
}

func (f *counting) State() play.State { f.reads++; return f.fake.State() }

// countGame is the presenter's game over a counting fake.
func countGame() (*classicGame, *counting) {
	f := &counting{fake: &fake{st: play.Exploring{}}}
	return &classicGame{Game: NewGame(f, game.Data{})}, f
}

// The shell asks for the View several times a frame (Screen, Free, Act's
// pacing, Draw): it is built once until something may change it.
func TestFrontBuildsViewOncePerFrame(t *testing.T) {
	g, c := countGame()
	f := &Front{Game: g, started: true}
	f.Screen()
	one := c.reads // one View build's reads
	f.Free()
	f.Free()
	if c.reads != one {
		t.Fatalf("Free built the View again: %d reads, want %d", c.reads, one)
	}
	f.Act() // nothing to run: the View stays
	f.Screen()
	f.Free()
	if c.reads != one {
		t.Fatalf("after an idle Act: %d reads, want %d", c.reads, one)
	}
	f.Key(' ') // a key may change it: the next ask builds it again
	f.Screen()
	if c.reads != 2*one {
		t.Fatalf("after Key: %d reads, want %d", c.reads, 2*one)
	}
}

// The minimap is drawn and uploaded when the map changes, not for every
// View (each builds a new MapSquares).
func TestMinimapUploadsOnChange(t *testing.T) {
	var p Presenter
	dst := gpu.NewImage(640, 400)
	font := func(byte, bool) [8]byte { return [8]byte{} }
	m := &front.MapSquares{Party: image.Pt(1, 1)}
	p.drawMap(dst, front.View{Map: m}, font)
	first := p.mapSrc
	same := *m
	p.drawMap(dst, front.View{Map: &same}, font)
	if p.mapSrc != first {
		t.Error("an unchanged map was drawn again")
	}
	moved := *m
	moved.Party.X = 2
	p.drawMap(dst, front.View{Map: &moved}, font)
	if p.mapSrc == first {
		t.Error("a changed map was not drawn again")
	}
}

// A key pressed while the fight's events play is kept, not dropped: it
// acts once they are played (Draw hands it on).
func TestKeyDuringPlayKept(t *testing.T) {
	g, _ := countGame()
	f := &Front{Game: g, started: true}
	f.P.Play([]front.CombatEvent{{Ms: 5000, Battle: &front.Battle{}}})
	f.Key('x')
	if len(f.held) != 1 || f.held[0] != 'x' || !f.P.Playing() {
		t.Fatalf("held %v, playing %v", f.held, f.P.Playing())
	}
	f.Key(' ') // Space still cuts the moment short
	if len(f.held) != 1 {
		t.Fatalf("held %v", f.held)
	}
}

// The camera's keys (Q, E, Home) turn it while the fight's events play: not
// held for after, and the moment plays on.
func TestCameraKeyDuringPlay(t *testing.T) {
	g, _ := countGame()
	f := &Front{Game: g, started: true}
	f.P.fight = true
	f.P.Play([]front.CombatEvent{{Ms: 5000, Battle: &front.Battle{}}})
	f.Key('q')
	if len(f.held) != 0 || !f.P.Playing() || f.P.replay.skip {
		t.Fatalf("held %v, playing %v, skip %v", f.held, f.P.Playing(), f.P.replay.skip)
	}
}
