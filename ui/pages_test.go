package ui

import (
	"gobox/game"
	"image"
	"slices"
	"testing"
)

// The spell lists' title replaces the sheet's name row (SpellList: a box
// over row 1, the name's, the title after it), and the PNG portrait goes.
func TestSheetSpellsTitle(t *testing.T) {
	p := partyOn(game.Member{Name: "Mirelle", Okay: true, Page: &game.SheetPage{Name: "Mirelle", Spells: "Spells in Memory"}})
	p.eng.r.sheet = 0
	ink := p.ink()
	c := p.screenPage()
	got := c.lines()[1]
	want := []PaneRun{{X: 1, Text: "Mirelle's", Fg: ink[inkCyan], Bg: ink[0]}, {X: 11, Text: "Spells in Memory", Fg: ink[inkGreen], Bg: ink[0]}}
	if !slices.Equal(got, want) {
		t.Errorf("row 1: %+v, want %+v", got, want)
	}
	if !slices.Contains(c.boxes, CellRect{X1: 1, Y1: 1, X2: 38, Y2: 1}) {
		t.Errorf("no box round the title: %v", c.boxes)
	}
}

// The focus goes to the nearest part the arrow points at.
func TestNearestSpot(t *testing.T) {
	r := func(x, y int) spot { return spot{r: CellRect{X1: x, Y1: y, X2: x + 3, Y2: y}} }
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
	p, _ := presenterOn(game.Scene{Kind: game.SceneResults, Results: &game.Results{Outcome: "The party has won."}})
	if p.ownLayout() {
		t.Fatal("the after-combat page still lays out on the classic page")
	}
}

// lines are the cells as runs of one colour, a row each.
func (c *cellPage) lines() [][]PaneRun {
	ls := make([][]PaneRun, c.rows)
	for y, row := range c.cells {
		var l []PaneRun
		for x, k := range row {
			if !k.set {
				continue
			}
			if n := len(l) - 1; n >= 0 && l[n].X+len(l[n].Text) == x && l[n].Fg == k.fg && l[n].Bg == k.bg {
				l[n].Text += string(k.ch)
				continue
			}
			l = append(l, PaneRun{X: x, Text: string(k.ch), Fg: k.fg, Bg: k.bg})
		}
		ls[y] = l
	}
	return ls
}

// The sheet lists the saving throws under the names the state gives, however many (no fixed five).
func TestSheetSaveNames(t *testing.T) {
	saves := []game.Throw{{Name: "Nerve", Value: 9}, {Name: "Luck", Value: 12}, {Name: "Grit", Value: 7}}
	p := partyOn(game.Member{Name: "Mirelle", Okay: true, Page: &game.SheetPage{Name: "Mirelle", Saves: saves}})
	p.eng.r.sheet = 0
	c := p.screenPage()
	seen := map[string]bool{}
	for _, l := range c.lines() {
		for _, r := range l {
			seen[r.Text] = true
		}
	}
	for _, s := range saves {
		if !seen[s.Name] {
			t.Errorf("save %q is not on the sheet", s.Name)
		}
	}
}
