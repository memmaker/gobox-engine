package app

import (
	"testing"

	"gobox/gpu"
)

// busyGame is busy for a set number of frames after each key, like a
// fade or a walk of varying pace.
type busyGame struct {
	busy  []int // frames busy after key i
	left  int
	keys  []Key
	early bool // a key came while busy
}

func (g *busyGame) Update()         {}
func (g *busyGame) Draw(*gpu.Image) {}
func (g *busyGame) Idle() bool      { return g.left == 0 }
func (g *busyGame) Step(k Key) {
	if g.left > 0 {
		g.left--
	}
	if k == 0 {
		return
	}
	g.early = g.early || g.left > 0 && k != Wait
	g.left = g.busy[len(g.keys)%len(g.busy)]
	g.keys = append(g.keys, k)
}

// The same script ends the same way however long the game is busy: every
// key waits for the game to be idle.
func TestScriptWaitsIdle(t *testing.T) {
	keys := []Key{'a', Wait, 'b', 'c'}
	for _, busy := range [][]int{{0}, {50, 3, 100, 7}} {
		g := &busyGame{busy: busy}
		s := &shell{cfg: Config{Game: g, Capture: Capture{Keys: keys}}}
		for i := 0; i < 10000 && !s.done(); i++ {
			s.step()
		}
		if !s.done() || len(g.keys) != len(keys) || g.early {
			t.Fatalf("busy %v: keys %v early %v done %v", busy, g.keys, g.early, s.done())
		}
	}
}

// altGame records each key with whether Alt was held as it came.
type altGame struct {
	busyGame
	alts []bool
}

func (g *altGame) Step(k Key) {
	if k != 0 {
		g.keys = append(g.keys, k)
		g.alts = append(g.alts, gpu.IsKeyPressed(gpu.KeyAlt))
	}
}

// \A<c> types c with Alt held, and only c.
func TestAltEscape(t *testing.T) {
	g := &altGame{busyGame: busyGame{busy: []int{0}}}
	s := &shell{cfg: Config{Game: g, Capture: Capture{Keys: ParseKeys(`\Aqx\Aa`, nil)}}}
	for i := 0; i < 1000 && !s.done(); i++ {
		s.step()
	}
	if string(rune(g.keys[0]))+string(rune(g.keys[1]))+string(rune(g.keys[2])) != "qxa" || !g.alts[0] || g.alts[1] || !g.alts[2] {
		t.Fatalf("keys %v alts %v", g.keys, g.alts)
	}
}

// A script ending on \t<ms> runs that long before it is done (its shot).
func TestTimedEndWaits(t *testing.T) {
	g := &busyGame{busy: []int{0}}
	s := &shell{cfg: Config{Game: g, Capture: Capture{Keys: ParseKeys(`\t2000`, nil)}}}
	n := 0
	for ; n < 10000 && !s.done(); n++ {
		s.step()
	}
	if n < 120 {
		t.Fatalf("done after %d frames, want 2 s (120)", n)
	}
}
