package engine

import (
	"image"
	"testing"

	"gobox/game"
)

// TestWalkMapMarksEventSquares: New Phlan's walk map names the squares
// its step script dispatches on (the temples, the shops: the cell byte's
// number), so gobox's auto-walk goes round them; the streets are not marked.
func TestWalkMapMarksEventSquares(t *testing.T) {
	e := testEngine(t)
	e.Explore = true // straight into New Phlan, its map loaded
	e.Init()
	e.Start(e.Play)
	e.viewMode = 1 // the 3D town, as the game shows it
	m, ok := game.Map(walkMap{e}).(game.MarkedMap)
	if !ok {
		t.Fatal("the walk map is no game.MarkedMap")
	}
	marked := 0
	for y := range 16 {
		for x := range 16 {
			at := image.Pt(x, y)
			if got, want := m.Marked(at), e.MapCell(x, y)&0x7F != 0; got != want {
				t.Fatalf("%v: marked %v with cell byte %#x", at, got, e.MapCell(x, y))
			} else if got {
				marked++
			}
		}
	}
	if marked == 0 || marked > 128 {
		t.Fatalf("%d of 256 squares marked: want the buildings, not none or the streets too", marked)
	}
	e.viewMode = 0
	if m.Marked(image.Pt(11, 2)) {
		t.Error("the area map marks a square")
	}
}
