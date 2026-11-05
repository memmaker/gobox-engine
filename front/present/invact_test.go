package present

import (
	"image"
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
	for _, k := range []front.ActKind{front.ActHalve, front.ActJoin} {
		p.inv.acts = append(p.inv.acts, front.Action{Kind: k, N: 1})
	}
	p.inv.acts = append(p.inv.acts, front.Action{Kind: front.ActTrade, N: 1, X: 2, Y: 1}) // Trade, its member asked
	if as := p.Actions(); len(as) != 3 {
		t.Fatalf("passed %v", as)
	}
}

// Left highlights the shop's first ware, Enter buys it; a long name wraps.
func TestShopKeys(t *testing.T) {
	in := &front.Inventory{Wares: []front.InvItem{{Index: 2, Name: "Rope"}}, Pool: []front.InvItem{{Index: 0, Name: "Gun"}}, Sell: true}
	var p Presenter
	p.inv.view, p.inv.ware, p.inv.sel = in, -1, 0
	p.shownView = front.View{Inventory: in}
	p.inventoryKey(keyOf(input.Left))
	p.inventoryKey(keyOf(input.Enter))
	p.Actions()
	g := &buyGame{}
	for _, c := range p.Calls() {
		c(g)
	}
	if g.bought != 3 {
		t.Fatalf("bought %d", g.bought)
	}
	if a, b := wrapName("1 POISON ANTIDOTE", 12); a != "1 POISON" || b != "ANTIDOTE" {
		t.Fatalf("%q %q", a, b)
	}
}

// A drag off the doll unreadies, as a click there and a click on the list do;
// a readied item dragged from the list onto the doll stays as it is.
func TestInventoryDrag(t *testing.T) {
	gun := front.InvItem{Index: 1, Name: "Gun", Type: 7, Readied: true}
	in := &front.Inventory{Doll: []*front.InvItem{&gun}, Pool: []front.InvItem{gun}}
	var p Presenter
	inv := &p.inv
	inv.k, inv.kx, inv.ky, inv.view, inv.sel, inv.ware = 1, 1, 1, in, -1, -1
	inv.dollAt = image.Rect(0, 0, 100, 100)
	drag := func(from hit, x, y int) []front.Action {
		inv.lines = []hit{from}
		p.InventoryMouse(from.r.Min.X+5, from.r.Min.Y+5, true, false, 0)
		p.InventoryMouse(x, y, true, false, 0)
		p.InventoryMouse(x, y, false, false, 0)
		as := inv.acts
		inv.acts = nil
		return as
	}
	if as := drag(hit{r: image.Rect(0, 0, 20, 20), item: gun, doll: true}, 200, 200); len(as) != 1 || as[0].Kind != front.ActReady {
		t.Fatalf("off the doll: %v", as)
	}
	if as := drag(hit{r: image.Rect(150, 150, 300, 170), item: gun}, 50, 50); len(as) != 0 {
		t.Fatalf("readied onto the doll: %v", as)
	}
}

// Trade asks the member over the items: the pick, not the items, is drawn.
func TestTradeAsksOverItems(t *testing.T) {
	in := &front.Inventory{Pool: []front.InvItem{{Index: 1, Name: "Gun"}}, Trade: true}
	v := front.View{Inventory: in, Party: []front.Member{{Selected: true}, {}}}
	var p Presenter
	p.shownView = v
	p.inv.acts = []front.Action{{Kind: front.ActTrade, N: 1}}
	p.Actions()
	if p.q == nil || p.Inventory(nil, Note{}, v) {
		t.Fatal("no member pick over the items")
	}
}
