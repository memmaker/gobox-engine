package present

import (
	"testing"

	"gobox/front"
)

// With the loot's item list up, the gate menu (View/Take/Exit) must not go
// over it, else the list's Take is swallowed as ActOK (POR class 83).
func TestLootListNoGateMenu(t *testing.T) {
	p := &Presenter{}
	v := front.View{Loot: &front.Loot{Take: true}, Inventory: &front.Inventory{Loot: true}}
	if a := p.ownQuestion(v).Ask; a != nil {
		t.Fatalf("gate menu over the loot list: %+v", a.Menu)
	}
	v.Inventory = nil
	if p.ownQuestion(v).Ask == nil {
		t.Fatal("no gate menu on the bare treasure")
	}
}
