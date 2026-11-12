package pool

import (
	"rules1e"
	"testing"

	"gobox/game/play"
)

// TestStackAndToasts: a second like stack bought joins the first and the
// listed name and count show the sum (BUG-CLASSES 73); a purchase and a
// refusal tell it in the outcome, a journal text adds to Lore (BUG-CLASSES 70).
func TestStackAndToasts(t *testing.T) {
	e := devPlay(t, "shop")
	if _, ok := e.State().(play.Trading); !ok {
		t.Fatalf("shop scene: %T", e.State())
	}
	a := e.state().RandomItem(rules1e.ItemArrow)
	a.Price, a.Count, a.Hidden = 1, 10, 0
	e.state().Treasure = []*rules1e.Item{a}
	e.state().Purse.Gold = 1
	arrows := func() (string, int, int) {
		n := 0
		var name string
		var count int
		for _, g := range commonOf(e.State()).Party[e.sel].Items {
			if g.Type == int(rules1e.ItemArrow) {
				n, name, count = n+1, g.Name, g.Count
			}
		}
		return name, count, n
	}
	said := func(o play.Outcome) string {
		for _, ev := range o.Events {
			switch ev := ev.(type) {
			case play.Narrated:
				return ev.Text
			case play.Refused:
				return ev.Why
			}
		}
		return ""
	}
	buy := play.Buy{Member: e.sel, Item: e.goods()[0]}
	if o, err := e.Do(buy); err != nil || said(o) != "Bought!" {
		t.Errorf("after a buy: %q %v", said(o), err)
	}
	if o, err := e.Do(buy); err == nil && said(o) == "" { // no gold left
		t.Errorf("a refused buy told nothing")
	}
	e.state().Purse.Gold = 1
	if o, err := e.Do(buy); err != nil {
		t.Fatalf("third buy: %v %+v", err, o)
	}
	if name, count, n := arrows(); n != 1 || count != 20 || name[:3] != "20 " {
		t.Errorf("two stacks bought: %d listed, %q count %d, want one of 20", n, name, count)
	}
	lore := len(commonOf(e.State()).Lore)
	for title := range e.journalTexts() {
		e.noteJournal(title)
		break
	}
	if len(commonOf(e.State()).Lore) != lore+1 {
		t.Errorf("Lore %d after a journal text, want %d", len(commonOf(e.State()).Lore), lore+1)
	}
}
