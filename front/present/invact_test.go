package present

import (
	"testing"

	"gobox/front"
	"gobox/input"
)

// buyGame records the rule calls the inventory's decisions become.
type buyGame struct {
	front.Game
	bought int
}

func (g *buyGame) Buy(n int) error { g.bought = n + 1; return nil }

// Down then B on the shop's wares buys the highlighted ware (a rule call);
// Trade, Halve and Join reach the game's Answer.
func TestInventoryActs(t *testing.T) {
	in := &front.Inventory{Wares: []front.InvItem{{Index: 3, Name: "Sword"}}, Sell: true, Trade: true, Halve: true, Join: true}
	var p Presenter
	p.inv.view, p.inv.ware, p.inv.sel = in, -1, -1
	p.shownView = front.View{Inventory: in}
	p.inventoryKey(keyOf(input.Down))
	p.inventoryKey(0x42)
	as := p.Actions()
	g := &buyGame{}
	for _, c := range p.Calls() {
		c(g)
	}
	if g.bought != 4 || len(as) != 0 {
		t.Fatalf("bought %d, passed %v", g.bought, as)
	}
	for _, k := range []front.ActKind{front.ActTrade, front.ActHalve, front.ActJoin} {
		p.inv.acts = append(p.inv.acts, front.Action{Kind: k, N: 1})
	}
	if as := p.Actions(); len(as) != 3 {
		t.Fatalf("passed %v", as)
	}
}
