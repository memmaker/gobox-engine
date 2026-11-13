package present

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// With the loot's item list up, the gate menu (View/Take/Exit) must not go
// over it, else the list's Take is swallowed as ActOK (POR class 83).
func TestLootListNoGateMenu(t *testing.T) {
	g, _ := playGame(play.Looting{Common: play.Common{Party: []game.Member{{Name: "A"}}}, Items: []game.Item{{Handle: 4, Name: "Sword"}}})
	p := &Presenter{}
	p.SetEngine(g)
	v := front.View{Loot: &front.Loot{Take: true}}
	if a := p.ownQuestion(v).Ask; a != nil {
		t.Fatalf("gate menu over the loot list: %+v", a.Menu)
	}
	g, _ = playGame(play.Looting{Common: play.Common{Party: []game.Member{{Name: "A"}}}, Announce: true})
	p.SetEngine(g)
	if p.ownQuestion(v).Ask == nil {
		t.Fatal("no gate menu on the bare treasure")
	}
}
