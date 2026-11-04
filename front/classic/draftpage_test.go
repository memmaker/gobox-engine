package classic

import (
	"image"
	"testing"

	"gobox/front"
	"gobox/game"
)

// With no page from the rules, the creation sheet's page is the draft's.
func TestSheetFromDraft(t *testing.T) {
	c := &game.Creation{Draft: game.Member{Name: "Ann", Race: "Human", Class: "Fighter", Stats: map[string]int{"str": 18, "cha": 9}, Str00: 100},
		Fields: []game.Field{{Name: "gender", Options: []string{"Male", "Female"}, Pick: 1}}}
	p := Sheet(front.View{}, game.State{Mode: game.ModeCreate, Creation: c}).SheetPage
	if p == nil || !p.Creating || p.Sex != "Female" || p.Stats[0] != 18 || p.Stats[5] != 9 || p.Str00 != "(00)" || p.Race != "Human" {
		t.Fatalf("%+v", p)
	}
}

// cellRules is a creation draft on a page of Cells (zero: the game keeps the original's 40x25).
type cellRules struct {
	seqRules
	cells image.Point
}

func (r *cellRules) State() game.State {
	return game.State{Mode: game.ModeCreate, Cells: r.cells, Creation: &game.Creation{}}
}

// A resize lays the creation sheet out to the page the game shows: a game that
// keeps 40x25 (no Cells) gets its icon on that page, not on the window's grid.
func TestCreationOnShownPage(t *testing.T) {
	for _, c := range []image.Point{{}, {80, 50}} {
		g := NewGame(&cellRules{cells: c}, game.Data{})
		g.Answer(front.Action{Kind: front.ActResize, X: 80, Y: 50})
		cols := max(c.X, 40)
		if ic := g.View().Creation.Icon; ic.X2 >= cols {
			t.Errorf("cells %v: icon %+v off the %d-column page", c, ic, cols)
		}
		if ic := g.View().Creation.Icon; c.X == 80 && ic.X2 < 40 {
			t.Errorf("cells %v: icon %+v not on the wide page", c, ic)
		}
	}
}
