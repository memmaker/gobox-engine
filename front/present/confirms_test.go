package present

import (
	"strings"
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// TestConfirmsDropMember: the game's DropMember confirm asks its texts one
// after another, on its default, and drops only after the last Yes.
func TestConfirmsDropMember(t *testing.T) {
	confirms = map[string]game.Confirm{"DropMember": {Asks: []string{"Drop %s?", "Really drop %s?"}}}
	defer func() { confirms = nil }()
	p := campPresenter()
	p.memberAct(front.ActDropMember, 0)
	for _, c := range p.Calls() { // the member led first
		c(p.eng)
	}
	p.eng.G.(*fake).sent = nil
	title, rows := p.campItems()
	if title != "Drop Ann?" || p.camp.sel != 1 {
		t.Fatalf("first ask %q on row %d", title, p.camp.sel)
	}
	rows[0].Do()
	title, rows = p.campItems()
	if title != "Really drop Ann?" || p.camp.sel != 1 || len(p.inv.acts) != 0 {
		t.Fatalf("second ask %q on row %d, sent %v", title, p.camp.sel, p.inv.acts)
	}
	for _, c := range p.Calls() {
		c(p.eng)
	}
	if sent := p.eng.G.(*fake).sent; len(sent) != 0 {
		t.Fatalf("sent %v before the last Yes", sent)
	}
	rows[0].Do()
	for _, c := range p.Calls() {
		c(p.eng)
	}
	if sent := p.eng.G.(*fake).sent; len(sent) != 1 || sent[0] != (play.DropMember{Member: 0}) || p.camp.screen != campMain {
		t.Fatalf("after both: %v screen %d", sent, p.camp.screen)
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
	paid, free, stack := game.Item{Sale: map[string]int{Coin: 5}, Count: 3}, game.Item{}, game.Item{Count: 4}
	for want, it := range map[string]game.Item{"Sell": paid, "SellFree": free, "SellFreeMany": stack} {
		if got := itemConfirm(front.ActSell, it); got != want {
			t.Errorf("%+v: %s, want %s", it, got, want)
		}
	}
	inv := &inventory{up: &invUp{pool: []game.Item{{Handle: 1, Sale: map[string]int{Coin: 1}}}}, sel: 1}
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
	got := sellAsk(&game.Item{Name: "Arrows", Count: 50, Sale: map[string]int{Coin: 7}})
	if !strings.HasPrefix(got, "I'll give you") || !strings.Contains(got, "Arrows") || !strings.HasSuffix(got, "Will you sell them? ") {
		t.Fatalf("%q", got)
	}
	if got := sellAsk(&game.Item{Name: "Rock"}); got != "Shall I dispose of it? " {
		t.Fatalf("free: %q", got)
	}
}
