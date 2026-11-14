package ui

import (
	"testing"

	"gobox/gpu"
)

// TestBoxEdges: every box draws the theme's "box." edges where it has them
// (the Amiga Curse's chain alone), the 3D view's box its "view." ones, and
// a theme with neither its general edges.
func TestBoxEdges(t *testing.T) {
	general, box, view := []*gpu.Image{nil}, []*gpu.Image{nil, nil}, []*gpu.Image{nil, nil, nil}
	m := map[string][]*gpu.Image{"top": general, boxEdges + "top": box, "view.top": view}
	if got := sizedEdges(m, 4, 4).Top; len(got) != len(box) {
		t.Errorf("a box drew %d top cells, want the box's %d", len(got), len(box))
	}
	edgePrefix = "view."
	defer func() { edgePrefix = "" }()
	if got := sizedEdges(m, 4, 4).Top; len(got) != len(view) {
		t.Errorf("the 3D view's box drew %d top cells, want its own %d", len(got), len(view))
	}
	edgePrefix = ""
	if got := sizedEdges(map[string][]*gpu.Image{"top": general}, 4, 4).Top; len(got) != len(general) {
		t.Errorf("a theme with no box edges drew %d top cells, want the general %d", len(got), len(general))
	}
}
