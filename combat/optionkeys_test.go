package combat

import (
	"image"
	"slices"
	"testing"

	"gobox/input"
)

// TestOptionKeys: the bar is the game's options; its own letters win over the
// camera's (Q turns the view otherwise), no remaster letter (F, U) doubles
// them, and an option the front has no act for is sent by its index.
func TestOptionKeys(t *testing.T) {
	b := &Board{Turn: &Turn{Options: []Option{{"Attack", 'A', ActAttack, ""}, {"Quick", 'Q', ActQuick, ""}, {"Target", 'T', ActOption, ""}, {"Dodge", 'D', ActOption, ""}}}}
	for _, c := range []struct {
		a    input.Act
		k    uint16
		want Decision
	}{
		{input.TurnLeft, 'q', Decision{Kind: ActQuick}},
		{input.None, 'a', Decision{Kind: ActAttack}},
		{input.None, 't', Decision{Kind: ActOption, N: 2}},
		{input.None, 'D', Decision{Kind: ActOption, N: 3}},
	} {
		p := &Presenter{b: b}
		p.bar.items = shownItems(b, false)
		p.bar.rects = make([]image.Rectangle, len(p.bar.items))
		if !p.Key(c.a, c.k) {
			t.Errorf("%c: not taken", c.k)
			continue
		}
		if got := p.Actions(); len(got) != 1 || got[0] != c.want {
			t.Errorf("%c: %+v, want %+v", c.k, got, c.want)
		}
	}
	for _, k := range []uint16{'f', 'u'} { // the game keyed Attack A and Quick Q: F and U are no buttons
		p := &Presenter{b: b}
		p.bar.items = shownItems(b, false)
		p.bar.rects = make([]image.Rectangle, len(p.bar.items))
		p.Key(input.None, k)
		if got := p.Actions(); len(got) != 0 {
			t.Errorf("%c: %+v", k, got)
		}
	}
}

// TestMoveOption: the game's Move starts the square cursor on the active
// unit; it never sends a walk to (0,0).
func TestMoveOption(t *testing.T) {
	b := &Board{Units: []Unit{{X: 3, Y: 4, Active: true}}, Turn: &Turn{Options: []Option{{"Move", 'M', ActWalk, ""}}}}
	p := &Presenter{b: b}
	p.bar.items = shownItems(b, false)
	p.bar.rects = make([]image.Rectangle, len(p.bar.items))
	if !p.Key(input.None, 'm') || len(p.Actions()) != 0 || p.free == nil || *p.free != image.Pt(3, 4) {
		t.Fatalf("free %v", p.free)
	}
}

// TestClickHitsUnit: a click on a foe's figure (above its square, where the
// floor ray lands on the square behind) is on the foe.
func TestClickHitsUnit(t *testing.T) {
	foe := Unit{ID: 7, X: 14, Y: 8, Squares: []image.Point{{14, 8}}}
	b := &Board{Units: []Unit{{ID: 1, X: 18, Y: 14, Active: true}, foe}, Targets: []int{7}, Turn: &Turn{Can: []Kind{ActAttack}}}
	p := &Presenter{b: b, screen: image.Rect(0, 0, 800, 600)}
	p.cam = camera{eye: vec{20, 9, 20}, at: vec{14, 0, 8}, fovy: 0.8, aspect: 4.0 / 3}
	nx, ny, _ := p.cam.ndc(vec{14.5, 0.8, 8.5}) // the figure's head
	x, y := int((nx+1)/2*800), int((1-ny)/2*600)
	if sx, sy, _ := p.Pick(x, y); sx == 14 && sy == 8 {
		t.Fatal("the floor ray already lands on the foe's square: no test")
	}
	if !p.Click(x, y) {
		t.Fatal("click not taken")
	}
	if got := p.Actions(); len(got) != 1 || got[0] != (Decision{Kind: ActAttack, N: 7}) {
		t.Fatalf("%+v", got)
	}
}

// TestOptionChords: a game's chords (Alt-Q, Space) pick its option.
func TestOptionChords(t *testing.T) {
	b := &Board{Turn: &Turn{Options: []Option{{"Quick", 'Q', ActQuick, ""}, {"Quick All", 0, ActOption, "alt+q, alt+a"}, {"Quick Off", 0, ActOption, "space"}}}}
	p := &Presenter{b: b}
	p.bar.items = shownItems(b, false)
	p.bar.rects = make([]image.Rectangle, len(p.bar.items))
	p.Alt = true
	p.Key(input.None, 'a')
	p.Alt = false
	p.Key(input.Space, ' ')
	p.Key(input.None, 'q')
	want := []Decision{{Kind: ActOption, N: 1}, {Kind: ActOption, N: 2}, {Kind: ActQuick}}
	if got := p.Actions(); !slices.Equal(got, want) {
		t.Fatalf("%+v", got)
	}
}

// TestAimClickFoe: aiming, a click on a foe's figure answers the foe's square
// (not the cursor's, which starts on the shooter).
func TestAimClickFoe(t *testing.T) {
	foe := Unit{ID: 7, X: 14, Y: 8, Squares: []image.Point{{14, 8}}}
	b := &Board{Units: []Unit{{ID: 1, X: 18, Y: 14, Active: true}, foe}, Cursor: image.Pt(18, 14), Turn: &Turn{Aim: true, Can: []Kind{ActAim}}}
	p := &Presenter{b: b, screen: image.Rect(0, 0, 800, 600)}
	p.cam = camera{eye: vec{20, 9, 20}, at: vec{14, 0, 8}, fovy: 0.8, aspect: 4.0 / 3}
	nx, ny, _ := p.cam.ndc(vec{14.5, 0.8, 8.5})
	x, y := int((nx+1)/2*800), int((1-ny)/2*600)
	p.Point(x, y)
	if !p.Click(x, y) {
		t.Fatal("click not taken")
	}
	got := p.Actions()
	if len(got) == 0 || got[len(got)-1] != (Decision{Kind: ActAim, X: 14, Y: 8}) {
		t.Fatalf("%+v", got)
	}
}

// TestAimCursorStarts: aiming, the game's Cursor with no CursorAt is drawn
// on the square it names, a big foe's squares when it is on one.
func TestAimCursorStarts(t *testing.T) {
	foe := Unit{ID: 7, X: 4, Y: 2, Squares: []image.Point{{4, 2}, {5, 2}}}
	b := &Board{Units: []Unit{foe}, Cursor: image.Pt(5, 2), CursorOn: true, Turn: &Turn{Aim: true}}
	var p Presenter
	got := p.aimBoard(b)
	if !slices.Equal(got.CursorAt, foe.Squares) || len(b.CursorAt) != 0 {
		t.Fatalf("cursor at %v (game's board %v)", got.CursorAt, b.CursorAt)
	}
	if p.aimBoard(&Board{}); p.aimAt != nil {
		t.Fatal("the aim cursor outlived the aim")
	}
}

// TestArrowsStepUnit: on its turn the arrows walk the unit one square where
// the game lets it (Reach), as the original's; a blocked step moves the look
// cursor, and aiming they move the aim.
func TestArrowsStepUnit(t *testing.T) {
	u := Unit{ID: 1, X: 2, Y: 2, Active: true}
	b := &Board{Units: []Unit{u}, Reach: []image.Point{{1, 2}}, Turn: &Turn{}}
	p := &Presenter{b: b}
	if !p.Key(input.CursorLeft, 0x4B00) {
		t.Fatal("left not taken")
	}
	if got := p.Actions(); len(got) != 1 || got[0] != (Decision{Kind: ActWalk, X: 1, Y: 2}) {
		t.Fatalf("left: %+v, want a walk to 1,2", got)
	}
	p.Key(input.CursorRight, 0x4D00) // 3,2 is not in Reach: the cursor starts
	if got := p.Actions(); len(got) != 0 || p.free == nil {
		t.Fatalf("blocked right: %+v, cursor %v", got, p.free)
	}
}

// TestAimStartsOnNearestFoe: an aim the game gives no cursor starts on the
// nearest foe, drawn, and Tab goes on to the next one: never (0,0).
func TestAimStartsOnNearestFoe(t *testing.T) {
	b := &Board{Turn: &Turn{Aim: true}, Units: []Unit{{ID: 1, X: 2, Y: 2, Party: true, Active: true},
		{ID: 8, X: 7, Y: 2, Squares: []image.Point{{7, 2}}}, {ID: 9, X: 3, Y: 3, Squares: []image.Point{{3, 3}}}}}
	var p Presenter
	got := p.aimBoard(b)
	if !got.CursorOn || got.Cursor != image.Pt(3, 3) || !slices.Equal(got.Targets, []int{9, 8}) {
		t.Fatalf("cursor %v on %v, targets %v", got.Cursor, got.CursorOn, got.Targets)
	}
	p.cycleTarget(false)
	if p.b.Cursor != image.Pt(7, 2) {
		t.Fatalf("Tab: cursor %v, want the next foe", p.b.Cursor)
	}
}
