package ui

import (
	"testing"

	"gobox/game"
)

// With the loot's item list up, no question goes over it, else the list's
// Take is swallowed as ActOK (POR class 83); the bare treasure is announced.
func TestLootListNoGateMenu(t *testing.T) {
	g, _ := playGame(game.Looting{Common: game.Common{Party: []game.Member{{Name: "A"}}}, Items: []game.Item{{Handle: 4, Name: "Sword"}}})
	p := &Presenter{}
	p.SetEngine(g)
	if a := p.askUp(); a != nil {
		t.Fatalf("a question over the loot list: %+v", a.Menu)
	}
	g, _ = playGame(game.Looting{Common: game.Common{Party: []game.Member{{Name: "A"}}}, Announce: true})
	p.SetEngine(g)
	if p.askUp() == nil {
		t.Fatal("the bare treasure is not announced")
	}
}
