package combat

import (
	"image"
	"testing"

	"gobox/gpu"
	"gobox/ui/input"
)

// TestFreeAimSteps: aiming, the arrows step the cursor onto any square, a
// foe's or an empty one, from the game's highlighted foe; Enter on it
// (the game taking nothing) centres the view there.
func TestFreeAimSteps(t *testing.T) {
	dst := gpu.NewImage(320, 200)
	defer dst.Deallocate()
	icon := image.NewRGBA(image.Rect(0, 0, 24, 24))
	board := func() *Board {
		return &Board{Tiles: []image.Image{image.NewRGBA(image.Rect(0, 0, tilePx, tilePx))},
			Units: []Unit{{ID: 1, X: 5, Y: 5, Size: 1, Squares: []image.Point{{5, 5}}, Party: true, Active: true, Icon: icon},
				{ID: 2, X: 8, Y: 5, Size: 1, Squares: []image.Point{{8, 5}}, Icon: icon}},
			Turn: &Turn{Aim: true}, Targets: []int{2}, Cursor: image.Pt(8, 5), CursorOn: true, CursorAt: []image.Point{{8, 5}}}
	}
	var p Presenter
	p.Draw(dst, board())
	start := p.b.Cursor
	for range 3 {
		p.Key(input.CursorUp, 0)
		p.Draw(dst, board())
	}
	if p.b.Cursor == start {
		t.Fatalf("cursor stayed on %v", p.b.Cursor)
	}
	at := p.b.Cursor
	p.Key(input.Enter, '\r')
	p.Actions()
	p.Draw(dst, board())
	if p.b.Cursor != at {
		t.Errorf("cursor %v after Enter, want %v", p.b.Cursor, at)
	}
	if want := (vec{float64(at.X) + .5, 0, float64(at.Y) + .5}); p.focus != want {
		t.Errorf("focus %v, want the cursor %v", p.focus, want)
	}
}
