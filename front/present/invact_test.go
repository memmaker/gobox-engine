package present

import (
	"image"
	"slices"
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
	"gobox/input"
	"gobox/rules"
)

// Down then B on the shop's wares buys the highlighted ware (a rule call);
// Trade, Halve and Join reach the game's Answer.
func TestInventoryActs(t *testing.T) {
	in := &invUp{kind: invShop, wares: []game.Item{{Handle: 4, Name: "Sword"}}, sell: true, trade: true, halve: true, join: true}
	g, f := playGame(play.Trading{Common: play.Common{Party: []game.Member{{}}}})
	g.r.items = game.ItemsShop
	var p Presenter
	p.SetEngine(g)
	p.inv.up, p.inv.ware, p.inv.sel = in, rules.NoItem, rules.NoItem
	p.inventoryKey(keyOf(input.Down))
	p.inventoryKey(0x42)
	as := p.Actions()
	for _, c := range p.Calls() {
		c(g)
	}
	if bought(f) != 4 || len(as) != 0 {
		t.Fatalf("sent %v, passed %v", f.sent, as)
	}
	for _, k := range []front.ActKind{front.ActHalve, front.ActJoin} {
		p.inv.acts = append(p.inv.acts, front.FrontAction{Kind: k, Item: 2})
	}
	p.inv.acts = append(p.inv.acts, front.FrontAction{Kind: front.ActTrade, Item: 2, X: 2, Y: 1}) // Trade, its member asked
	if as := p.Actions(); len(as) != 3 {
		t.Fatalf("passed %v", as)
	}
}

// Left highlights the shop's first ware, Enter buys it; a long name wraps.
func TestShopKeys(t *testing.T) {
	in := &invUp{kind: invShop, wares: []game.Item{{Handle: 3, Name: "Rope"}}, pool: []game.Item{{Handle: 1, Name: "Gun"}}, sell: true}
	g, f := playGame(play.Trading{Common: play.Common{Party: []game.Member{{}}}})
	g.r.items = game.ItemsShop
	var p Presenter
	p.SetEngine(g)
	p.inv.up, p.inv.ware, p.inv.sel = in, rules.NoItem, 1
	p.inventoryKey(keyOf(input.InputLeft))
	p.inventoryKey(keyOf(input.Enter))
	p.Actions()
	for _, c := range p.Calls() {
		c(g)
	}
	if bought(f) != 3 {
		t.Fatalf("sent %v", f.sent)
	}
	if a, b := wrapName("1 POISON ANTIDOTE", 12); a != "1 POISON" || b != "ANTIDOTE" {
		t.Fatalf("%q %q", a, b)
	}
}

// A drag off the doll unreadies, as a click there and a click on the list do;
// a readied item dragged from the list onto the doll stays as it is.
func TestInventoryDrag(t *testing.T) {
	gun := game.Item{Handle: 2, Name: "Gun", Type: 7, Readied: true}
	in := &invUp{kind: invOwn, doll: []*game.Item{&gun}, pool: []game.Item{gun}}
	var p Presenter
	inv := &p.inv
	inv.k, inv.kx, inv.ky, inv.up, inv.sel, inv.ware = 1, 1, 1, in, rules.NoItem, rules.NoItem
	inv.dollAt = image.Rect(0, 0, 100, 100)
	drag := func(from hit, x, y int) []front.FrontAction {
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
	g, _ := playGame(play.Exploring{Common: play.Common{Party: []game.Member{{Items: []game.Item{{Handle: 2, Name: "Gun"}}}, {}}, Allowed: play.Allowed{Plain: []play.Action{play.Give{}}}}})
	g.r.items = game.ItemsOwn
	var p Presenter
	p.SetEngine(g)
	p.inv.acts = []front.FrontAction{{Kind: front.ActTrade, Item: 2}}
	p.Actions()
	if p.q == nil || p.q.Ask.Kind != front.AskMember {
		t.Fatal("no member pick")
	}
}

// Down walks every item once, readied ones too (they are on the doll and
// in the list); PgDn moves a page.
func TestInventoryKeysReachAll(t *testing.T) {
	gun := game.Item{Handle: 1, Name: "Gun", Readied: true}
	pool := []game.Item{gun}
	for i := 1; i < 16; i++ {
		pool = append(pool, game.Item{Handle: rules.ItemHandle(i + 1), Name: "Knife"})
	}
	in := &invUp{kind: invOwn, doll: []*game.Item{&gun}, pool: pool}
	var p Presenter
	p.inv.up, p.inv.ware, p.inv.sel, p.inv.list.rows = in, rules.NoItem, 1, 10
	seen := map[rules.ItemHandle]bool{}
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
	listed := play.Scene{Common: play.Common{Ask: &play.Question{List: true, Text: "Items", Replies: []string{"a", "b"}}}}
	p, f := presenterOn(listed)
	p.openList()
	p.list.sel = 1
	f.st = play.Scene{} // the list gone a moment (Ready)
	p.openList()
	f.st = listed
	p.openList()
	if p.list.sel != 1 {
		t.Fatalf("back on line %d", p.list.sel)
	}
}

// The vault's loot menu has a Drop word (the Deposit) for every device: the D
// key, a click and the pad's focus all go through the menu's words; it sends
// ActDrop at once, with no "gone forever" question.
func TestLootDropWord(t *testing.T) {
	in := &invUp{kind: invLoot, treasure: &front.FrontLoot{Drop: true}, wares: []game.Item{{Handle: 4, Name: "Sword"}},
		pool: []game.Item{{Handle: 1, Name: "Gun"}}}
	var p Presenter
	p.inv.up, p.inv.ware, p.inv.sel = in, rules.NoItem, 1
	if !slices.ContainsFunc(p.inv.menu(), func(w word) bool { return w.act == front.ActDrop && w.on }) {
		t.Fatal("no live Drop word in the loot menu")
	}
	p.inventoryKey('D')
	if p.inv.ask != nil {
		t.Fatal("Drop asked a question in the treasure")
	}
	if as := p.inv.acts; len(as) != 1 || as[0].Kind != front.ActDrop || as[0].Item != 1 {
		t.Fatalf("acts %v, want ActDrop on item 1", as)
	}
	in.treasure.Drop = false // an ordinary treasure: no Drop
	if slices.ContainsFunc(p.inv.menu(), func(w word) bool { return w.act == front.ActDrop }) {
		t.Fatal("Drop offered outside the vault")
	}
}
