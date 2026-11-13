package combat

import (
	"gobox/ui/input"
	"image"
	"slices"
	"testing"
)

// Felix's ruling 2026-11-13: Move, View, Aim and Flee have no buttons; the
// board takes them (an empty square walks, a foe is attacked, a step off the
// edge flees, the turn's own square is its sheet), by mouse, keys and pad
// (the pad's d-pad is the cursor arrows, its A Enter: bindings.go, fightKey).

// boardScene is a turn whose unit (ID 1) stands on the west edge at (0, 8),
// a foe (7) at (3, 8), (1, 8) in reach; the camera looks at the unit.
func boardScene() *Presenter {
	foe := Unit{ID: 7, X: 3, Y: 8, Squares: []image.Point{{3, 8}}}
	opts := []Option{{"Move", 'M', DecideWalk, ""}, {"View", 'V', DecideView, ""}, {"Aim", 0, DecideAim, ""}, {"Cast", 'C', DecideCast, ""}, {"Flee", 0, DecideFlee, ""}, {"Done", 'D', DecideEnd, ""}}
	b := &Board{Units: []Unit{{ID: 1, X: 0, Y: 8, Squares: []image.Point{{0, 8}}, Party: true, Active: true}, foe}, Targets: []int{7},
		Reach: []image.Point{{1, 8}}, Turn: &Turn{Options: opts}}
	p := &Presenter{b: b, screen: image.Rect(0, 0, 800, 600)}
	p.cam = camera{eye: vec{2, 9, 16}, at: vec{1, 0, 8}, fovy: 0.8, aspect: 4.0 / 3}
	p.bar.items = shownItems(b, false)
	p.bar.rects = make([]image.Rectangle, len(p.bar.items))
	return p
}

func TestBarHasNoBoardActs(t *testing.T) {
	p := boardScene()
	var words []string
	for _, it := range p.bar.items {
		words = append(words, it.Text)
	}
	if !slices.Equal(words, []string{"Cast", "Done"}) {
		t.Errorf("bar %v, want [Cast Done]", words)
	}
	if !p.Key(input.None, 'v') || !slices.Equal(p.Actions(), []Decision{{Kind: DecideView}}) { // View's key stays
		t.Error("V: no View")
	}
}

// screenAt is the pixel showing the middle of square sq.
func screenAt(p *Presenter, sq image.Point) (int, int) {
	nx, ny, _ := p.cam.ndc(vec{float64(sq.X) + .5, 0, float64(sq.Y) + .5})
	return int((nx + 1) / 2 * 800), int((1 - ny) / 2 * 600)
}

func TestBoardActsMouse(t *testing.T) {
	for _, c := range []struct {
		name string
		sq   image.Point
		want Decision
	}{
		{"empty square: move", image.Pt(1, 8), Decision{Kind: DecideWalk, X: 1, Y: 8}},
		{"foe: attack", image.Pt(3, 8), Decision{Kind: DecideAttack, N: 7}},
		{"own square: view", image.Pt(0, 8), Decision{Kind: DecideView, N: -1}},
		{"past the edge: flee", image.Pt(-1, 8), Decision{Kind: DecideFlee}},
	} {
		p := boardScene()
		x, y := screenAt(p, c.sq)
		if !p.Click(x, y) {
			t.Errorf("%s: click not taken", c.name)
			continue
		}
		if got := p.Actions(); !slices.Equal(got, []Decision{c.want}) {
			t.Errorf("%s: %+v, want %+v", c.name, got, c.want)
		}
	}
	p := boardScene() // past the edge away from the unit: nothing
	if x, y := screenAt(p, image.Pt(-1, 2)); p.Click(x, y) || len(p.Actions()) != 0 {
		t.Error("a click past the edge away from the unit taken")
	}
}

// TestBoardActsKeys: Enter starts the cursor on the unit, the arrows move it,
// Enter takes the square; an arrow off the edge from the unit flees.
func TestBoardActsKeys(t *testing.T) {
	right, left := input.CursorRight, input.CursorLeft // yaw 0: right is +X
	for _, c := range []struct {
		name string
		keys []input.Act
		want Decision
	}{
		{"cursor to empty square: move", []input.Act{input.Enter, right, input.Enter}, Decision{Kind: DecideWalk, X: 1, Y: 8}},
		{"cursor onto foe: attack", []input.Act{input.Enter, right, right, right, input.Enter}, Decision{Kind: DecideAttack, N: 7}},
		{"cursor on the unit: view", []input.Act{input.Enter, input.Enter}, Decision{Kind: DecideView, N: -1}},
		{"step off the edge: flee", []input.Act{left}, Decision{Kind: DecideFlee}},
		{"cursor off the edge: flee", []input.Act{input.Enter, left}, Decision{Kind: DecideFlee}},
		{"step: move", []input.Act{right}, Decision{Kind: DecideWalk, X: 1, Y: 8}},
	} {
		p := boardScene()
		for _, a := range c.keys {
			p.Key(a, 0)
		}
		if got := p.Actions(); !slices.Equal(got, []Decision{c.want}) {
			t.Errorf("%s: %+v, want %+v", c.name, got, c.want)
		}
	}
}
