package combat

import (
	"image"
	"slices"
	"testing"

	"gobox/input"
)

// TestOptionKeys: the bar is the game's options and its letters, except a
// camera key (W A S D Q E): there the camera keeps it and the button takes the
// remaster's letter (Attack F, Quick U, as POR before gobox); an option the
// front has no act for is sent by its index.
func TestOptionKeys(t *testing.T) {
	b := &Board{Turn: &Turn{Options: []CombatOption{{"Attack", 'A', CombatActAttack, ""}, {"Quick", 'Q', CombatActQuick, ""}, {"Target", 'T', CombatActOption, ""}, {"Dodge", 'D', CombatActOption, ""}}}}
	for _, c := range []struct {
		a    input.Act
		k    uint16
		want Decision
	}{
		{input.None, 'u', Decision{Kind: CombatActQuick}},
		{input.None, 'f', Decision{Kind: CombatActAttack}},
		{input.None, 't', Decision{Kind: CombatActOption, N: 2}},
	} {
		p := &CombatPresenter{b: b}
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
	for _, k := range []uint16{'a', 'q', 'd'} { // the game keyed Attack A, Quick Q, Dodge D: the camera keeps them (Dodge: its number)
		p := &CombatPresenter{b: b}
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
	b := &Board{Units: []Unit{{X: 3, Y: 4, Active: true}}, Turn: &Turn{Options: []CombatOption{{"Move", 'M', CombatActWalk, ""}}}}
	p := &CombatPresenter{b: b}
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
	b := &Board{Units: []Unit{{ID: 1, X: 18, Y: 14, Active: true}, foe}, Targets: []int{7}, Turn: &Turn{Can: []Kind{CombatActAttack}}}
	p := &CombatPresenter{b: b, screen: image.Rect(0, 0, 800, 600)}
	p.cam = camera{eye: vec{20, 9, 20}, at: vec{14, 0, 8}, fovy: 0.8, aspect: 4.0 / 3}
	nx, ny, _ := p.cam.ndc(vec{14.5, 0.8, 8.5}) // the figure's head
	x, y := int((nx+1)/2*800), int((1-ny)/2*600)
	if sx, sy, _ := p.Pick(x, y); sx == 14 && sy == 8 {
		t.Fatal("the floor ray already lands on the foe's square: no test")
	}
	if !p.Click(x, y) {
		t.Fatal("click not taken")
	}
	if got := p.Actions(); len(got) != 1 || got[0] != (Decision{Kind: CombatActAttack, N: 7}) {
		t.Fatalf("%+v", got)
	}
}

// TestOptionChords: a game's chords (Alt-Q, Space) pick its option; its Quick (Q) is on U.
func TestOptionChords(t *testing.T) {
	b := &Board{Turn: &Turn{Options: []CombatOption{{"Quick", 'Q', CombatActQuick, ""}, {"Quick All", 0, CombatActOption, "alt+q, alt+a"}, {"Quick Off", 0, CombatActOption, "space"}}}}
	p := &CombatPresenter{b: b}
	p.bar.items = shownItems(b, false)
	p.bar.rects = make([]image.Rectangle, len(p.bar.items))
	p.Alt = true
	p.Key(input.None, 'a')
	p.Alt = false
	p.Key(input.Space, ' ')
	p.Key(input.None, 'u')
	want := []Decision{{Kind: CombatActOption, N: 1}, {Kind: CombatActOption, N: 2}, {Kind: CombatActQuick}}
	if got := p.Actions(); !slices.Equal(got, want) {
		t.Fatalf("%+v", got)
	}
}

// TestAimClickFoe: aiming, a click on a foe's figure answers the foe's square
// (not the cursor's, which starts on the shooter).
func TestAimClickFoe(t *testing.T) {
	foe := Unit{ID: 7, X: 14, Y: 8, Squares: []image.Point{{14, 8}}}
	b := &Board{Units: []Unit{{ID: 1, X: 18, Y: 14, Active: true}, foe}, Cursor: image.Pt(18, 14), Turn: &Turn{Aim: true, Can: []Kind{CombatActAim}}}
	p := &CombatPresenter{b: b, screen: image.Rect(0, 0, 800, 600)}
	p.cam = camera{eye: vec{20, 9, 20}, at: vec{14, 0, 8}, fovy: 0.8, aspect: 4.0 / 3}
	nx, ny, _ := p.cam.ndc(vec{14.5, 0.8, 8.5})
	x, y := int((nx+1)/2*800), int((1-ny)/2*600)
	p.Point(x, y)
	if !p.Click(x, y) {
		t.Fatal("click not taken")
	}
	got := p.Actions()
	if len(got) == 0 || got[len(got)-1] != (Decision{Kind: CombatActAim, X: 14, Y: 8}) {
		t.Fatalf("%+v", got)
	}
}

// TestAimCursorStarts: aiming, the game's Cursor with no CursorAt is drawn
// on the square it names, a big foe's squares when it is on one.
func TestAimCursorStarts(t *testing.T) {
	foe := Unit{ID: 7, X: 4, Y: 2, Squares: []image.Point{{4, 2}, {5, 2}}}
	b := &Board{Units: []Unit{foe}, Cursor: image.Pt(5, 2), CursorOn: true, Turn: &Turn{Aim: true}}
	var p CombatPresenter
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
	p := &CombatPresenter{b: b}
	if !p.Key(input.CursorLeft, 0x4B00) {
		t.Fatal("left not taken")
	}
	if got := p.Actions(); len(got) != 1 || got[0] != (Decision{Kind: CombatActWalk, X: 1, Y: 2}) {
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
	var p CombatPresenter
	got := p.aimBoard(b)
	if !got.CursorOn || got.Cursor != image.Pt(3, 3) || !slices.Equal(got.Targets, []int{9, 8}) {
		t.Fatalf("cursor %v on %v, targets %v", got.Cursor, got.CursorOn, got.Targets)
	}
	p.cycleTarget(false)
	if p.b.Cursor != image.Pt(7, 2) {
		t.Fatalf("Tab: cursor %v, want the next foe", p.b.Cursor)
	}
}

// TestChordLabel: a verb picked only by a chord shows the chord, not a blank.
func TestChordLabel(t *testing.T) {
	if h := hotkeysOf([]barItem{{Text: "Quick All", Keys: "alt+q,alt+a"}}); h[0].label != "A-Q" || h[0].key != 0 {
		t.Fatalf("%+v", h[0])
	}
}

// TestMoveThenArrowSteps: Move (M) puts the cursor on the unit; an arrow then
// steps the unit, not the cursor (POR castfight).
func TestMoveThenArrowSteps(t *testing.T) {
	u := Unit{ID: 1, X: 2, Y: 2, Active: true}
	b := &Board{Units: []Unit{u}, Reach: []image.Point{{1, 2}}, Turn: &Turn{}}
	p := &CombatPresenter{b: b}
	p.press(barItem{Kind: CombatActWalk})
	p.Key(input.CursorLeft, 0x4B00)
	if got := p.Actions(); len(got) != 1 || got[0] != (Decision{Kind: CombatActWalk, X: 1, Y: 2}) {
		t.Fatalf("M, left: %+v, want a walk to 1,2", got)
	}
	p.Key(input.CursorUp, 0x4800) // not in Reach (off the map): still sent, the game answers (POR: Flee?)
	if got := p.Actions(); len(got) != 1 || got[0].Kind != CombatActWalk {
		t.Fatalf("M, off-Reach step: %+v, want a walk", got)
	}
}
