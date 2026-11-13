package present

import (
	"strings"
	"testing"

	"gobox/front"
	"gobox/game"
)

// TestConfirmsDropMember: the game's DropMember confirm asks its texts one
// after another, on its default, and drops only after the last Yes.
func TestConfirmsDropMember(t *testing.T) {
	confirms = map[string]game.Confirm{"DropMember": {Asks: []string{"Drop %s?", "Really drop %s?"}}}
	defer func() { confirms = nil }()
	p := campPresenter()
	v := front.View{Camp: &front.Camp{}, Party: []front.Member{{Name: "Ann", Selected: true}}}
	p.memberAct(front.ActDropMember, 0)
	p.inv.acts = nil
	title, rows := p.campItems(v)
	if title != "Drop Ann?" || p.camp.sel != 1 {
		t.Fatalf("first ask %q on row %d", title, p.camp.sel)
	}
	rows[0].Do()
	title, rows = p.campItems(v)
	if title != "Really drop Ann?" || p.camp.sel != 1 || len(p.inv.acts) != 0 {
		t.Fatalf("second ask %q on row %d, sent %v", title, p.camp.sel, p.inv.acts)
	}
	rows[0].Do()
	if len(p.inv.acts) != 1 || p.inv.acts[0].Kind != front.ActDropMember || p.camp.screen != campMain {
		t.Fatalf("after both: %v screen %d", p.inv.acts, p.camp.screen)
	}
}

// TestConfirmText: %s is the name, %p the price.
func TestConfirmText(t *testing.T) {
	if s := confirmText("Sell %s for %p?", "a sword", "5 gp"); s != "Sell a sword for 5 gp?" {
		t.Fatal(s)
	}
}

// TestItemConfirm: Sell's variant by price and count, falling back to
// Sell; the inventory's question starts on the game's Confirm.Yes.
func TestItemConfirm(t *testing.T) {
	confirms = map[string]game.Confirm{"Sell": {Asks: []string{"Will you sell it?"}, Yes: false}, "SellFree": {Asks: []string{"Shall I dispose of it for you?"}},
		"SellFreeMany": {Asks: []string{"Would you like to discard them?"}}, "Drop": {Asks: []string{"Lose it? "}, Yes: true}}
	defer func() { confirms = nil }()
	paid, free, stack := &front.InvItem{Price: front.Money{Gold: 5}, Count: 3}, &front.InvItem{}, &front.InvItem{Count: 4}
	for it, want := range map[*front.InvItem]string{paid: "Sell", free: "SellFree", stack: "SellFreeMany"} {
		if got := itemConfirm(front.ActSell, it); got != want {
			t.Errorf("%+v: %s, want %s", it, got, want)
		}
	}
	inv := &inventory{view: &front.Inventory{Pool: []front.InvItem{{Item: 1, Price: front.Money{Gold: 1}}}}, sel: 1}
	inv.do(front.ActSell)
	if inv.ask == nil || inv.askSel != 1 {
		t.Fatalf("sell asks on row %d, want No (Confirm.Yes false)", inv.askSel)
	}
	inv.ask = nil
	inv.do(front.ActDrop)
	if inv.ask == nil || inv.askSel != 0 {
		t.Fatalf("drop asks on row %d, want Yes", inv.askSel)
	}
}

// TestSellAskOffer: a game's sell ask that is the question alone still has
// the offer before it; a free item has none.
func TestSellAskOffer(t *testing.T) {
	old := confirms
	defer func() { confirms = old }()
	confirms = map[string]game.Confirm{"SellMany": {Asks: []string{"Will you sell them? "}}, "SellFree": {Asks: []string{"Shall I dispose of it? "}}}
	got := sellAsk(&front.InvItem{Name: "Arrows", Count: 50, Price: front.Money{Gold: 7}})
	if !strings.HasPrefix(got, "I'll give you") || !strings.Contains(got, "Arrows") || !strings.HasSuffix(got, "Will you sell them? ") {
		t.Fatalf("%q", got)
	}
	if got := sellAsk(&front.InvItem{Name: "Rock"}); got != "Shall I dispose of it? " {
		t.Fatalf("free: %q", got)
	}
}
