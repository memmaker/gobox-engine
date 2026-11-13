package present

import (
	"gobox/front"
	"image"
	"image/color"
	"slices"
	"testing"
)

// The spell lists' title replaces the sheet's name row (SpellList: a box
// over row 1, the name's, the title after it), and the PNG portrait goes.
func TestSheetSpellsTitle(t *testing.T) {
	var ink [16]color.RGBA
	for i := range ink {
		ink[i] = col(i)
	}
	v := front.View{Cells: image.Pt(40, 25), Ink: ink, SheetPage: &front.SheetPage{Name: "Mirelle", NameColour: inkCyan, Spells: "Spells in Memory"}}
	c := (&Presenter{}).screenPage(v)
	got := c.lines()[1]
	want := []front.PaneRun{{X: 1, Text: "Mirelle's", Fg: ink[inkCyan], Bg: ink[0], Upcase: true}, {X: 11, Text: "Spells in Memory", Fg: ink[inkGreen], Bg: ink[0], Upcase: true}}
	if !slices.Equal(got, want) {
		t.Errorf("row 1: %+v, want %+v", got, want)
	}
	if !slices.Contains(c.boxes, front.CellRect{X1: 1, Y1: 1, X2: 38, Y2: 1}) {
		t.Errorf("no box round the title: %v", c.boxes)
	}
}

// The focus goes to the nearest part the arrow points at.
func TestNearestSpot(t *testing.T) {
	r := func(x, y int) spot { return spot{r: front.CellRect{X1: x, Y1: y, X2: x + 3, Y2: y}} }
	sp := []spot{r(1, 1), r(1, 3), r(8, 3), r(30, 2), r(0, 24)}
	for _, c := range []struct {
		from int
		d    image.Point
		want int
	}{{0, image.Pt(0, 1), 1}, {1, image.Pt(1, 0), 2}, {2, image.Pt(1, 0), 3}, {1, image.Pt(0, 1), 4}, {0, image.Pt(0, -1), 0}} {
		if got := nearest(sp, c.from, c.d); got != c.want {
			t.Errorf("from %d by %v: %d, want %d", c.from, c.d, got, c.want)
		}
	}
}

// The after-combat page is a screen of its own: drawn with no page under
// it (no classic page, no engine pictures).
func TestResultsOwnScreen(t *testing.T) {
	v := front.View{Cells: image.Pt(40, 25), Results: &front.Results{Outcome: "The party has won."}}
	if (&Presenter{}).ownLayout(v) {
		t.Fatal("the after-combat page still lays out on the classic page")
	}
}
