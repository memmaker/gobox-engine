package present

import (
	"image"
	"slices"
	"testing"

	"gobox/core"
	"gobox/front"
	"gobox/game"
	"gobox/input"
)

// buyGame records the rule calls the inventory's decisions become.
type buyGame struct {
	front.Game
	bought int
}

func (g *buyGame) Buy(h core.ItemHandle) error { g.bought = int(h); return nil }

// Down then B on the shop's wares buys the highlighted ware (a rule call);
// Trade, Halve and Join reach the game's Answer.
func TestInventoryActs(t *testing.T) {
	in := &front.Inventory{Wares: []front.InvItem{{Item: 4, Name: "Sword"}}, Sell: true, Trade: true, Halve: true, Join: true}
	var p Presenter
	p.inv.view, p.inv.ware, p.inv.sel = in, core.NoItem, core.NoItem
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
		p.inv.acts = append(p.inv.acts, front.Action{Kind: k, Item: 2})
	}
	p.inv.acts = append(p.inv.acts, front.Action{Kind: front.ActTrade, Item: 2, X: 2, Y: 1}) // Trade, its member asked
	if as := p.Actions(); len(as) != 3 {
		t.Fatalf("passed %v", as)
	}
}

// Left highlights the shop's first ware, Enter buys it; a long name wraps.
func TestShopKeys(t *testing.T) {
	in := &front.Inventory{Wares: []front.InvItem{{Item: 3, Name: "Rope"}}, Pool: []front.InvItem{{Item: 1, Name: "Gun"}}, Sell: true}
	var p Presenter
	p.inv.view, p.inv.ware, p.inv.sel = in, core.NoItem, 1
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
	gun := front.InvItem{Item: 2, Name: "Gun", Type: 7, Readied: true}
	in := &front.Inventory{Doll: []*front.InvItem{&gun}, Pool: []front.InvItem{gun}}
	var p Presenter
	inv := &p.inv
	inv.k, inv.kx, inv.ky, inv.view, inv.sel, inv.ware = 1, 1, 1, in, core.NoItem, core.NoItem
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

// Trade asks the member (drawn over the items).
func TestTradeAsksOverItems(t *testing.T) {
	in := &front.Inventory{Pool: []front.InvItem{{Item: 2, Name: "Gun"}}, Trade: true}
	v := front.View{Inventory: in, Party: []front.Member{{Selected: true}, {}}}
	var p Presenter
	p.shownView = v
	p.inv.acts = []front.Action{{Kind: front.ActTrade, Item: 2}}
	p.Actions()
	if p.q == nil || p.q.Ask.Kind != front.AskMember {
		t.Fatal("no member pick")
	}
}

// Down walks every item once, readied ones too (they are on the doll and
// in the list); PgDn moves a page.
func TestInventoryKeysReachAll(t *testing.T) {
	gun := front.InvItem{Item: 1, Name: "Gun", Readied: true}
	pool := []front.InvItem{gun}
	for i := 1; i < 16; i++ {
		pool = append(pool, front.InvItem{Item: core.ItemHandle(i + 1), Name: "Knife"})
	}
	in := &front.Inventory{Doll: []*front.InvItem{&gun}, Pool: pool}
	var p Presenter
	p.inv.view, p.inv.ware, p.inv.sel, p.inv.list.rows = in, core.NoItem, 1, 10
	seen := map[core.ItemHandle]bool{}
	for range 16 {
		p.inventoryKey(keyOf(input.Down))
		seen[p.inv.sel] = true
	}
	if len(seen) != 16 {
		t.Fatalf("Down reached %d of 16", len(seen))
	}
	p.inv.sel = 1
	p.inventoryKey(keyOf(input.PgDn))
	if p.inv.sel != 11 {
		t.Fatalf("PgDn to %d", p.inv.sel)
	}
}

// Buck Rogers' money is credits, shown as cr.
func TestCredits(t *testing.T) {
	defer func(s string) { Coin = s }(Coin)
	Coin = game.Credits
	if p := price(front.Money{Gold: 5}); p != "5cr" {
		t.Fatal(p)
	}
}

// After an act the same list comes back on the line acted on.
func TestListKeepsLine(t *testing.T) {
	v := front.View{Ask: &front.Ask{Kind: front.AskList, Menu: front.Menu{Prompt: "Items"}, Entries: []front.Entry{{Text: "a"}, {Text: "b"}}}}
	var p Presenter
	p.openList(v)
	p.list.sel = 1
	p.shownView = front.View{} // the list gone a moment (Ready)
	p.openList(v)
	if p.list.sel != 1 {
		t.Fatalf("back on line %d", p.list.sel)
	}
}

// The vault's loot menu has a Drop word (the Deposit) for every device: the D
// key, a click and the pad's focus all go through the menu's words; it sends
// ActDrop at once, with no "gone forever" question.
func TestLootDropWord(t *testing.T) {
	in := &front.Inventory{Loot: true, Treasure: &front.Loot{Drop: true}, Wares: []front.InvItem{{Item: 4, Name: "Sword"}},
		Pool: []front.InvItem{{Item: 1, Name: "Gun"}}}
	var p Presenter
	p.inv.view, p.inv.ware, p.inv.sel = in, core.NoItem, 1
	p.shownView = front.View{Inventory: in}
	if !slices.ContainsFunc(p.inv.menu(in), func(w word) bool { return w.act == front.ActDrop && w.on }) {
		t.Fatal("no live Drop word in the loot menu")
	}
	p.inventoryKey('D')
	if p.inv.ask != nil {
		t.Fatal("Drop asked a question in the treasure")
	}
	if as := p.inv.acts; len(as) != 1 || as[0].Kind != front.ActDrop || as[0].Item != 1 {
		t.Fatalf("acts %v, want ActDrop on item 1", as)
	}
	in.Treasure.Drop = false // an ordinary treasure: no Drop
	if slices.ContainsFunc(p.inv.menu(in), func(w word) bool { return w.act == front.ActDrop }) {
		t.Fatal("Drop offered outside the vault")
	}
}
