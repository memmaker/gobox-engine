package ui

import (
	"testing"

	"gobox/game"
)

// With no page from the rules, the creation sheet's page is the draft's.
func TestSheetFromDraft(t *testing.T) {
	c := &game.Creation{Draft: game.Member{Name: "Ann", Race: "Human", Class: "Fighter", Stats: map[string]int{"str": 18, "cha": 9}, Str00: 100},
		Fields: []game.Field{{Name: "gender", Options: []string{"Male", "Female"}, Pick: 1}}}
	p := draftPage(c, nil)
	if p == nil || !p.Creating || p.Sex != "Female" || p.Stats[0] != 18 || p.Stats[5] != 9 || p.Str00 != "(00)" || p.Race != "Human" {
		t.Fatalf("%+v", p)
	}
}

// A resize lays the creation sheet out to the window's grid at K: its
// icon on that page, not on the original's 40x25.
func TestCreationOnWindowGrid(t *testing.T) {
	g := NewGame(&fake{st: game.Assembling{Draft: &game.Draft{}}}, game.Data{})
	g.Answer(Action{Kind: ActResize, X: 80, Y: 50})
	if ic := presenterOf(g).creation().Icon; ic.X2 >= 80 || ic.Y2 >= 50 {
		t.Errorf("icon %+v off the 80x50 page", ic)
	}
}
