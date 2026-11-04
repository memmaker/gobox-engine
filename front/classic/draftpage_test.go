package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

// With no page from the rules, the creation sheet's page is the draft's.
func TestSheetFromDraft(t *testing.T) {
	c := &game.Creation{Draft: game.Member{Name: "Ann", Race: "Human", Class: "Fighter", Stats: map[string]int{"str": 18, "cha": 9}, Str00: 100},
		Fields: []game.Field{{Name: "gender", Options: []string{"Male", "Female"}, Pick: 1}}}
	p := Sheet(front.View{}, game.State{Creation: c}).SheetPage
	if p == nil || !p.Creating || p.Sex != "Female" || p.Stats[0] != 18 || p.Stats[5] != 9 || p.Str00 != "(00)" || p.Race != "Human" {
		t.Fatalf("%+v", p)
	}
}
