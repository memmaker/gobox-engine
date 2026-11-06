package combat

import (
	"image"
	"testing"

	"gobox/input"
)

// TestOptionKeys: the bar is the game's options; its own letters win over the
// camera's (Q turns the view otherwise), the remaster's (F, U) reach the same
// acts, and an option the front has no act for is sent by its index.
func TestOptionKeys(t *testing.T) {
	b := &Board{Turn: &Turn{Options: []Option{{"Attack", 'A', ActAttack}, {"Quick", 'Q', ActQuick}, {"Target", 'T', ActOption}, {"Dodge", 'D', ActOption}}}}
	for _, c := range []struct {
		a    input.Act
		k    uint16
		want Decision
	}{
		{input.TurnLeft, 'q', Decision{Kind: ActQuick}},
		{input.None, 'u', Decision{Kind: ActQuick}},
		{input.None, 'f', Decision{Kind: ActAttack}},
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
}

// TestMoveOption: the game's Move starts the square cursor on the active
// unit; it never sends a walk to (0,0).
func TestMoveOption(t *testing.T) {
	b := &Board{Units: []Unit{{X: 3, Y: 4, Active: true}}, Turn: &Turn{Options: []Option{{"Move", 'M', ActWalk}}}}
	p := &Presenter{b: b}
	p.bar.items = shownItems(b, false)
	p.bar.rects = make([]image.Rectangle, len(p.bar.items))
	if !p.Key(input.None, 'm') || len(p.Actions()) != 0 || p.free == nil || *p.free != image.Pt(3, 4) {
		t.Fatalf("free %v", p.free)
	}
}
