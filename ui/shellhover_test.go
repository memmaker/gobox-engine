package ui

import "testing"

// hoverFront counts the Hovers it is given.
type hoverFront struct {
	frontPort
	hovers int // Hover and Mouse calls
}

func (f *hoverFront) InventoryShown() bool      { return false }
func (f *hoverFront) Hover(x, y int, down bool) { f.hovers++ }
func (f *hoverFront) Click(x, y int) bool       { return false }

// TestStillPointerNoHover: a pointer that never moves moves no highlight
// (a scripted run's pointer resting over a menu row or a map square must not
// pick it under the keys); a moved one does.
func TestStillPointerNoHover(t *testing.T) {
	f := &hoverFront{}
	g := &frontGame{f: f, cfg: shellConfig{Capture: Capture{Shot: "x.png", Pointer: [][2]int{{100, 100}, {200, 200}}}}}
	for range 3 {
		g.mouse()
	}
	if f.hovers != 0 {
		t.Fatalf("a still pointer hovered %d times", f.hovers)
	}
	g.pointer = 1
	g.mouse()
	if f.hovers != 2 { // Hover and Mouse
		t.Fatalf("a moved pointer hovered %d times, want 2", f.hovers)
	}
}

func (f *hoverFront) Cell(x, y float64) (int, int, bool) { return 0, 0, true }
func (f *hoverFront) Free() bool                         { return false }
func (f *hoverFront) Mouse(cx, cy int, in, left, right bool, steps int) {
	f.hovers++
}
