package present

import (
	"image"
	"testing"

	"gobox/front"
)

type markGame struct {
	front.Game
	marked   string
	unmarked image.Point
}

func (g *markGame) FullMapView(at image.Point, label *string) *front.FullMap {
	return &front.FullMap{Mine: true}
}
func (g *markGame) Mark(x, y int, text string) { g.marked = text; g.unmarked = image.Pt(x, y) }
func (g *markGame) Unmark(x, y int)            { g.unmarked = image.Pt(-x, -y) }

// TestFullMapMarkUnmark: the cell menu's Mark types a label that Enter sends
// to Game.Mark; Remove calls Game.Unmark.
func TestFullMapMarkUnmark(t *testing.T) {
	g := &markGame{}
	p := &Presenter{}
	p.SetEngine(g)
	p.fullMap = &fullMapUI{cursor: image.Pt(3, 4)}
	p.mapPick(1) // Walk, Mark, Remove
	for _, k := range []uint16{'h', 'i', 13} {
		p.mapKey(k)
	}
	p.fullMap.menu = true
	p.mapPick(2)
	for _, f := range p.Calls() {
		f(g)
	}
	if g.marked != "hi" || g.unmarked != image.Pt(-3, -4) {
		t.Fatalf("Mark %q, Unmark %v", g.marked, g.unmarked)
	}
}
