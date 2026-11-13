package present

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
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

// A resize lays the creation sheet out to the page the game shows: a game that
// keeps 40x25 (no Cells) gets its icon on that page, not on the window's grid.
func TestCreationOnShownPage(t *testing.T) {
	g := NewGame(&fake{st: play.Assembling{Draft: &play.Draft{}}}, game.Data{})
	g.Answer(front.FrontAction{Kind: front.FrontActResize, X: 80, Y: 50})
	if ic := presenterOf(g).creation().Icon; ic.X2 >= 40 {
		t.Errorf("icon %+v off the 40-column page", ic)
	}
}
