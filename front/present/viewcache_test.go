package present

import (
	"testing"

	"gobox/front"
	"gobox/gpu"
	"image"
)

// countGame counts View builds (each is the rules' whole State).
type countGame struct {
	front.Game
	views int
	v     front.View
}

func (g *countGame) View() front.View { g.views++; return g.v }

// The shell asks for the View several times a frame (Screen, Free, Act's
// pacing, Draw): it is built once until something may change it.
func TestFrontBuildsViewOncePerFrame(t *testing.T) {
	g := &countGame{}
	f := &Front{Game: g, started: true}
	f.Screen()
	f.Free()
	f.Free()
	if g.views != 1 {
		t.Fatalf("Screen and Free built the View %d times, want 1", g.views)
	}
	f.Act() // nothing to run: its pacing builds the frame's View once
	f.Screen()
	f.Free()
	if g.views != 2 {
		t.Fatalf("after Act: %d builds, want 2", g.views)
	}
	f.Key(' ') // a key may change it: the next ask builds it again
	f.Screen()
	if g.views != 3 {
		t.Fatalf("after Key: %d builds, want 3", g.views)
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
